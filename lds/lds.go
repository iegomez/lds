package lds

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//Message holds physical payload and rx info.
type Message struct {
	PhyPayload string  `json:"phyPayload"`
	RxInfo     *RxInfo `json:"rxInfo"`
}

//RxInfo holds all relevant information of a lora message.
type RxInfo struct {
	Channel   int       `json:"channel"`
	CodeRate  string    `json:"codeRate"`
	CrcStatus int       `json:"crcStatus"`
	DataRate  *DataRate `json:"dataRate"`
	Frequency int       `json:"frequency"`
	LoRaSNR   float32   `json:"loRaSNR"`
	Mac       string    `json:"mac"`
	RfChain   int       `json:"rfChain"`
	Rssi      int       `json:"rssi"`
	Size      int       `json:"size"`
	Time      string    `json:"time"`
	Timestamp int32     `json:"timestamp"`
}

//DataRate holds relevant info for data rate.
type DataRate struct {
	Bandwidth    int    `json:"bandwidth"`
	Modulation   string `json:"modulation"`
	SpreadFactor int    `json:"spreadFactor"`
	BitRate      int    `json:"bitrate"`
}

//Device holds device keys, addr, eui and fcnt.
type Device struct {
	DevEUI    lorawan.EUI64
	DevAddr   lorawan.DevAddr
	NwkSKey   lorawan.AES128Key
	AppSKey   lorawan.AES128Key
	AppKey    [16]byte
	AppEUI    lorawan.EUI64
	UlFcnt    uint32
	DlFcnt    uint32
	marshal   func(msg proto.Message) ([]byte, error)
	unmarshal func(b []byte, msg proto.Message) error
}

func (d *Device) SetMarshaler(opt string) {
	log.Printf("switching to marshaler: %s\n", opt)
	switch opt {
	case "json":
		d.marshal = func(msg proto.Message) ([]byte, error) {
			marshaler := &jsonpb.Marshaler{
				EnumsAsInts:  false,
				EmitDefaults: true,
			}
			str, err := marshaler.MarshalToString(msg)
			return []byte(str), err
		}

		d.unmarshal = func(b []byte, msg proto.Message) error {
			unmarshaler := &jsonpb.Unmarshaler{
				AllowUnknownFields: true, // we don't want to fail on unknown fields
			}
			return unmarshaler.Unmarshal(bytes.NewReader(b), msg)
		}

	case "protobuf":
		d.marshal = func(msg proto.Message) ([]byte, error) {
			return proto.Marshal(msg)
		}

		d.unmarshal = func(b []byte, msg proto.Message) error {
			return proto.Unmarshal(b, msg)
		}
	default:
		//Plain old json.
		d.marshal = func(msg proto.Message) ([]byte, error) {
			return json.Marshal(msg)
		}

		d.unmarshal = func(b []byte, msg proto.Message) error {
			return json.Unmarshal(b, msg)
		}
	}
}

//Join sends a join request for a given device (OTAA) and rxInfo.
func (d *Device) Join(client MQTT.Client, gwMac string, rxInfo RxInfo) error {

	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  d.AppEUI,
			DevEUI:   d.DevEUI,
			DevNonce: lorawan.DevNonce(uint16(65535)),
		},
	}

	if err := joinPhy.SetUplinkJoinMIC(d.AppKey); err != nil {
		return err
	}

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		return err
	}

	message := &Message{
		PhyPayload: string(joinStr),
		RxInfo:     &rxInfo,
	}

	pErr := publish(client, "gateway/"+rxInfo.Mac+"/rx", message)

	return pErr

}

//Uplink sends an uplink message for a given device, mType (UnconfirmedDataUp, ConfirmedDataUp), rxInfo and payload (unencrypted).
func (d *Device) Uplink(client MQTT.Client, mType lorawan.MType, fPort uint8, rxInfo *gw.UplinkRXInfo, txInfo *gw.UplinkTXInfo, payload []byte, gwMAC string) error {

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: mType,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: d.DevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt:  d.UlFcnt,
				FOpts: []lorawan.Payload{}, // you can leave this out when there is no MAC command to send
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: payload}},
		},
	}

	if err := phy.EncryptFRMPayload(d.AppSKey); err != nil {
		fmt.Printf("encrypt frm payload: %s", err)
		return err
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, d.NwkSKey, d.NwkSKey); err != nil {
		fmt.Printf("set uplink mic error: %s", err)
		return err
	}

	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		if err != nil {
			fmt.Printf("marshal binary error: %s", err)
			return err
		}
	}

	message := gw.UplinkFrame{
		PhyPayload: phyBytes,
		RxInfo:     rxInfo,
		TxInfo:     txInfo,
	}

	fmt.Printf("Message PHY payload: %v\n", string(message.PhyPayload))

	bytes, err := d.marshal(&message)
	if err != nil {
		return err
	}

	fmt.Printf("Marshaled message: %v\n", string(bytes))

	if token := client.Publish("gateway/"+gwMAC+"/rx", 0, false, bytes); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}

	//pErr := publish(client, "gateway/"+rxInfo.Mac+"/rx", bytes)

	//Increase uplink fcnt if unconfirmed
	/*if mType == lorawan.UnconfirmedDataUp {
		d.UlFcnt++
	}*/

	return nil

}

/////////////////////////
// Auxiliary functions //
/////////////////////////

//HexToDevAddress converts a string hex representation of a device address to a [4]byte.
func HexToDevAddress(hexAddress string) ([4]byte, error) {
	var devAddr ([4]byte)
	da, err := hex.DecodeString(hexAddress)
	if err != nil {
		return devAddr, err
	}
	copy(devAddr[:], da[:])
	return devAddr, nil
}

//HexToKey converts a string hex representation of an AES128Key to a [16]byte.
func HexToKey(hexKey string) ([16]byte, error) {
	var key ([16]byte)
	k, err := hex.DecodeString(hexKey)
	if err != nil {
		return key, err
	}
	copy(key[:], k[:])
	return key, nil
}

//HexToEUI converts a string hex representation of a devEUI to a lorawan.EUI.
func HexToEUI(hexEUI string) (lorawan.EUI64, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(hexEUI)); err != nil {
		fmt.Errorf("wron eui: %s\n", err)
		return eui, err
	}
	return eui, nil
}

//MACToGatewayID converts a string mac to a gateway id byte slice.
func MACToGatewayID(mac string) ([]byte, error) {
	bytes, err := hex.DecodeString(mac)
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}

func testMIC(appKey [16]byte, appEUI, devEUI [8]byte) error {
	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  appEUI,
			DevEUI:   devEUI,
			DevNonce: lorawan.DevNonce(uint16(65535)),
		},
	}

	if err := joinPhy.SetUplinkJoinMIC(appKey); err != nil {
		fmt.Printf("set uplink join mic error: %s", err)
		return err
	}

	fmt.Println("Printing MIC")
	fmt.Println(hex.EncodeToString(joinPhy.MIC[:]))

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		fmt.Printf("join marshal error: %s", err)
		return err
	}
	fmt.Println(joinStr)

	return nil
}

//publish publishes a message to the broker.
func publish(client MQTT.Client, topic string, v interface{}) error {

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println("Publishing:")
	fmt.Println(string(bytes))

	if token := client.Publish(topic, 0, false, bytes); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}

	return nil
}

/////////////////////////
// Custom data helpers //
/////////////////////////

func generateRisk(r int8) []byte {
	risk := uint8(r)
	bRep := make([]byte, 1)
	bRep[0] = risk
	return bRep
}

func generateTemp1byte(t int8) []byte {
	temp := uint8(t)
	bRep := make([]byte, 1)
	bRep[0] = temp
	return bRep
}

func generateTemp2byte(t int16) []byte {

	temp := uint16(float32(t/127.0) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, temp)
	return bRep
}

func generateLight(l int16) []byte {

	light := uint16(l)
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, light)
	return bRep
}

func generateAltitude(a float32) []byte {

	alt := uint16(float32(a/1200) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, alt)
	return bRep
}

func generateLat(l float32) []byte {
	lat := uint32((l / 90.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lat)
	return bRep
}

func generateLng(l float32) []byte {
	lng := uint32((l / 180.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lng)
	return bRep
}

func GenerateFloat(originalFloat, maxValue float32, numBytes int32) []byte {
	byteArray := make([]byte, numBytes)
	if numBytes == 4 {
		encodedFloat := uint32((originalFloat / maxValue) * float32(math.Pow(2, 31)))
		binary.BigEndian.PutUint32(byteArray, encodedFloat)
	} else if numBytes == 3 {
		encodedFloat := uint32((originalFloat / maxValue) * float32(math.Pow(2, 23)))
		temp := make([]byte, 4)
		binary.BigEndian.PutUint32(temp, encodedFloat)
		byteArray = temp[1:]
	} else if numBytes == 2 {
		encodedFloat := uint16((originalFloat / maxValue) * float32(math.Pow(2, 15)))
		binary.BigEndian.PutUint16(byteArray, encodedFloat)
	} else if numBytes == 1 {
		byteArray[0] = uint8(originalFloat)
	}
	return byteArray
}

func GenerateInt(originalInt, numBytes int32) []byte {

	bRep := make([]byte, numBytes)
	if numBytes == 4 {
		binary.BigEndian.PutUint32(bRep, uint32(originalInt))
	} else if numBytes == 3 {
		temp := make([]byte, 4)
		binary.BigEndian.PutUint32(temp, uint32(originalInt))
		bRep = temp[1:]
	} else if numBytes == 2 {
		binary.BigEndian.PutUint16(bRep, uint16(originalInt))
	} else if numBytes == 1 {
		bRep[0] = uint8(originalInt)
	}

	return bRep
}
