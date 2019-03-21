package lds

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan/band"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis"
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
	DevEUI      lorawan.EUI64      `json:"devEUI"`
	DevAddr     lorawan.DevAddr    `json:"devAddr"`
	NwkSEncKey  lorawan.AES128Key  `json:"nwkSEncKey"`
	SNwkSIntKey lorawan.AES128Key  `json:"sNwkSIntKey"`
	FNwkSIntKey lorawan.AES128Key  `json:"fNwksSIntKey"`
	AppSKey     lorawan.AES128Key  `json:"appSKey"`
	NwkKey      [16]byte           `json:"nwkKey"`
	AppKey      [16]byte           `json:"appKey"`
	JoinEUI     lorawan.EUI64      `json:"joinEUI"`
	Major       lorawan.Major      `json:"major"`
	MACVersion  lorawan.MACVersion `json:"macVersion"`
	UlFcnt      uint32             `json:"ulFcnt"`
	DlFcnt      uint32             `json:"dlFcnt"`
	marshal     func(msg proto.Message) ([]byte, error)
	unmarshal   func(b []byte, msg proto.Message) error
	Profile     string            `json:"profile"`
	Joined      bool              `json:"joined"`
	DevNonce    lorawan.DevNonce  `json:"devNonce"`
	JoinNonce   lorawan.JoinNonce `json:"joinNonce"`
}

var redisClient *redis.Client

//StartRedis tries to connect to Redis (used for DevNonce and JoinNonce).
func StartRedis(addr, password string, db int) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Errorf("couldn't start Redis, only ABP profile will work. error: %s", err)
		return err
	}
	return nil
}

//SetMarshaler sets marshaling and unmarshaling functions according to the given option.
func (d *Device) SetMarshaler(opt string) {
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

	d.Joined = false
	devNonceKey := fmt.Sprintf("dev-nonce-%s", d.DevEUI[:])
	var devNonce uint16
	sdn, err := redisClient.Get(devNonceKey).Result()
	if err == nil {
		dn, err := strconv.Atoi(sdn)
		if err == nil {
			devNonce = uint16(dn + 1)
		}
	}
	redisClient.Set(devNonceKey, devNonce, 0)

	d.DevNonce = lorawan.DevNonce(devNonce)

	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  d.JoinEUI,
			DevEUI:   d.DevEUI,
			DevNonce: d.DevNonce,
		},
	}

	if err := joinPhy.SetUplinkJoinMIC(d.NwkKey); err != nil {
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

//Uplink sends an uplink message as if it was sent from a lora-gateway-bridge. Works only for ABP devices with relaxed frame counter.
func (d *Device) Uplink(client MQTT.Client, mType lorawan.MType, fPort uint8, rxInfo *gw.UplinkRXInfo, txInfo *gw.UplinkTXInfo, payload []byte, gwMAC string, bandName band.Name, dr DataRate, macCommands []*lorawan.MACCommand) (uint32, error) {

	//Get uplink frame counter.
	ulFcntKey := fmt.Sprintf("ul-fcnt-%s", d.DevEUI[:])
	uf, err := redisClient.Get(ulFcntKey).Result()
	if err == nil {
		ufn, err := strconv.Atoi(uf)
		if err == nil {
			d.UlFcnt = uint32(ufn) + 1
		}
	}

	var fOpts = make([]lorawan.Payload, len(macCommands))
	for i := 0; i < len(fOpts); i++ {
		fOpts[i] = macCommands[i]
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: mType,
			Major: d.Major,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: d.DevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       true,
				},
				FCnt:  d.UlFcnt,
				FOpts: fOpts,
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: payload}},
		},
	}

	if err := phy.EncryptFRMPayload(d.AppSKey); err != nil {
		log.Debugf("encrypt frm payload: %s", err)
		return d.UlFcnt, err
	}

	if d.MACVersion == lorawan.LoRaWAN1_0 {
		if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, d.NwkSEncKey, d.NwkSEncKey); err != nil {
			log.Debugf("set uplink mic error: %s", err)
			return d.UlFcnt, err
		}
		phy.ValidateUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, d.NwkSEncKey, d.NwkSEncKey)
	} else if d.MACVersion == lorawan.LoRaWAN1_1 {
		//Get the band.
		b, err := band.GetConfig(bandName, false, lorawan.DwellTime400ms)
		if err != nil {
			return d.UlFcnt, err
		}
		//Get DR index from a dr.
		dataRate := band.DataRate{
			Modulation:   band.Modulation(dr.Modulation),
			SpreadFactor: dr.SpreadFactor,
			Bandwidth:    dr.Bandwidth,
			BitRate:      dr.BitRate,
		}
		txDR, err := b.GetDataRateIndex(true, dataRate)
		if err != nil {
			return d.UlFcnt, err
		}
		//Get tx ch.
		var txCh int
		for _, defaultChannel := range []bool{true, false} {
			i, err := b.GetUplinkChannelIndex(int(txInfo.Frequency), defaultChannel)
			if err != nil {
				continue
			}

			c, err := b.GetUplinkChannel(i)
			if err != nil {
				return d.UlFcnt, err
			}

			// there could be multiple channels using the same frequency, but with different data-rates.
			// eg EU868:
			//  channel 1 (868.3 DR 0-5)
			//  channel x (868.3 DR 6)
			if c.MinDR <= txDR && c.MaxDR >= txDR {
				txCh = i
			}
		}
		//Encrypt fOPts.
		if err := phy.EncryptFOpts(d.NwkSEncKey); err != nil {
			log.Errorf("encrypt fopts error: %s", err)
			return d.UlFcnt, err
		}

		//Now set the MIC.
		if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_1, d.UlFcnt, uint8(txDR), uint8(txCh), d.FNwkSIntKey, d.SNwkSIntKey); err != nil {
			log.Errorf("set uplink mic error: %s", err)
			return d.UlFcnt, err
		}

	} else {
		return d.UlFcnt, errors.New("unknown lorawan version")
	}

	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		if err != nil {
			log.Debugf("marshal binary error: %s\n", err)
			return d.UlFcnt, err
		}
	}

	message := gw.UplinkFrame{
		PhyPayload: phyBytes,
		RxInfo:     rxInfo,
		TxInfo:     txInfo,
	}

	log.Debugf("payload: %v\n", string(message.PhyPayload))

	bytes, err := d.marshal(&message)
	if err != nil {
		return d.UlFcnt, err
	}

	log.Debugf("marshaled message: %v\n", string(bytes))

	if token := client.Publish("gateway/"+gwMAC+"/rx", 0, false, bytes); token.Wait() && token.Error() != nil {
		log.Errorf("publish error: %s", token.Error())
		return d.UlFcnt, token.Error()
	}

	//Message was sent, UlFcnt can be set.
	redisClient.Set(ulFcntKey, d.UlFcnt, 0)

	return d.UlFcnt, nil

}

//ProcessDownlink processes a downlink message from the loraserver.
func (d *Device) ProcessDownlink(dlMessage []byte, mv lorawan.MACVersion) (string, error) {
	log.Debugf("original dlmessage: %s", string(dlMessage))
	var df map[string]interface{}
	err := json.Unmarshal(dlMessage, &df)
	if err != nil {
		return "", err
	}
	var phy lorawan.PHYPayload
	payload := []byte(df["phyPayload"].(string))
	//log.Infof("encrypted payload: %s", string(payload))
	if err := phy.UnmarshalText(payload); err != nil {
		log.Error("failed at unmarshal")
		return "", err
	}

	//Now we need to check the profile and if we are joined.
	if d.Profile == "ABP" || d.Joined {
		return d.processDownlink(phy, payload, mv)
	}

	//If we are not joined, we need to process the join response.
	return d.processJoinResponse(phy, payload, mv)

}

func (d *Device) processJoinResponse(phy lorawan.PHYPayload, payload []byte, mv lorawan.MACVersion) (string, error) {
	log.Infoln("processing join response")

	err := phy.DecryptJoinAcceptPayload(d.NwkKey)
	if err != nil {
		log.Errorf("can't decrypt join accept: %s", err)
	}

	ok, err := phy.ValidateDownlinkJoinMIC(lorawan.JoinRequestType, d.DevEUI, d.DevNonce, d.NwkKey)
	if err != nil {
		log.Error("failed at join mic function")
		return "", err
	}
	if !ok {
		return "", errors.New("join invalid mic")
	}

	if err := phy.DecryptFOpts(d.NwkKey); err != nil {
		log.Error("failed at opts decryption")
		return "", err
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		log.Error("failed at json marshal")
		return "", err
	}

	jap := phy.MACPayload.(*lorawan.JoinAcceptPayload)

	log.Debugf("join accept payload: %+v", jap)

	//Check that JoinNonce is greater than the one already stored.
	joinNonceKey := fmt.Sprintf("join-nonce-%s", d.DevEUI[:])
	var joinNonce lorawan.JoinNonce
	sjn, err := redisClient.Get(joinNonceKey).Result()
	if err == nil {
		jn, err := strconv.Atoi(sjn)
		if err == nil {
			joinNonce = lorawan.JoinNonce(jn + 1)
		}
	}

	if jap.JoinNonce <= joinNonce {
		return "", errors.New("got lower or equal JoinNonce from server")
	}

	redisClient.Set(joinNonceKey, joinNonce, 0)

	d.FNwkSIntKey, err = getFNwkSIntKey(jap.DLSettings.OptNeg, d.NwkKey, jap.HomeNetID, d.DevEUI, jap.JoinNonce, d.DevNonce)
	if d.MACVersion == 0 {
		d.NwkSEncKey = d.FNwkSIntKey
		d.SNwkSIntKey = d.FNwkSIntKey
	} else {
		d.NwkSEncKey, err = getNwkSEncKey(jap.DLSettings.OptNeg, d.NwkKey, jap.HomeNetID, d.DevEUI, jap.JoinNonce, d.DevNonce)
		d.SNwkSIntKey, err = getSNwkSIntKey(jap.DLSettings.OptNeg, d.NwkKey, jap.HomeNetID, d.DevEUI, jap.JoinNonce, d.DevNonce)
	}
	if jap.DLSettings.OptNeg {
		d.AppSKey, err = getAppSKey(jap.DLSettings.OptNeg, d.AppKey, jap.HomeNetID, d.DevEUI, jap.JoinNonce, d.DevNonce)
	} else {
		d.AppSKey, err = getAppSKey(jap.DLSettings.OptNeg, d.NwkKey, jap.HomeNetID, d.DevEUI, jap.JoinNonce, d.DevNonce)
	}
	d.DevAddr = jap.DevAddr
	d.Joined = true
	d.UlFcnt = 0
	d.DlFcnt = 0

	//Set frame counters to 0.
	ulFcntKey := fmt.Sprintf("ul-fcnt-%s", d.DevEUI[:])
	dlFcntKey := fmt.Sprintf("dl-fcnt-%s", d.DevEUI[:])

	redisClient.Set(ulFcntKey, d.UlFcnt, 0)
	redisClient.Set(dlFcntKey, d.DlFcnt, 0)

	return string(phyJSON), nil
}

func (d *Device) processDownlink(phy lorawan.PHYPayload, payload []byte, mv lorawan.MACVersion) (string, error) {

	//Get downlink frame counter and icnrease it by 1.
	dlFcntKey := fmt.Sprintf("dl-fcnt-%s", d.DevEUI[:])
	df, err := redisClient.Get(dlFcntKey).Result()
	if err == nil {
		dfn, err := strconv.Atoi(df)
		if err == nil {
			d.DlFcnt = uint32(dfn)
		}
	}

	ok, err := phy.ValidateDownlinkDataMIC(mv, 0, d.NwkSEncKey)
	if err != nil {
		log.Error("failed at downlink mic function")
		return "", err
	}
	if !ok {
		return "", errors.New("downlink error: invalid mic")
	}

	if err := phy.DecryptFOpts(d.NwkSEncKey); err != nil {
		log.Error("failed at downlink opts decryption")
		return "", err
	}

	if err := phy.DecryptFRMPayload(d.AppSKey); err != nil {
		log.Error("failed at downlink frm payload decryption")
		return "", err
	}

	/*if err := phy.DecodeFOptsToMACCommands(); err != nil {
		log.Error("failed at downlink opts to mac commands decoding")
		return "", err
	}*/

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		log.Error("failed at downlink json marshal")
		return "", err
	}

	macPayload := phy.MACPayload.(*lorawan.MACPayload)
	log.Infof("mac payload: %+v", macPayload)

	for _, fOpt := range macPayload.FHDR.FOpts {
		opt := fOpt.(*lorawan.MACCommand)
		log.Infof("fOpt: %+v", opt.Payload)
	}

	log.Infof("fctrl: %+v", macPayload.FHDR.FCtrl)

	for _, frmPayload := range macPayload.FRMPayload {
		log.Infof("data payload: %+v", frmPayload.(*lorawan.DataPayload))
	}

	log.Infof("dlFcnt: %d / received Fcnt: %d", d.DlFcnt, macPayload.FHDR.FCnt)

	//Set downlink frame counter.
	d.DlFcnt++
	redisClient.Set(dlFcntKey, d.DlFcnt, 0)
	return string(phyJSON), nil
}

//Reset clears all data from redis for a given device.
func (d *Device) Reset() error {
	dlFcntKey := fmt.Sprintf("dl-fcnt-%s", d.DevEUI[:])
	ulFcntKey := fmt.Sprintf("dl-fcnt-%s", d.DevEUI[:])
	joinNonceKey := fmt.Sprintf("join-nonce-%s", d.DevEUI[:])
	devNonceKey := fmt.Sprintf("dev-nonce-%s", d.DevEUI[:])
	_, err := redisClient.Del(dlFcntKey, ulFcntKey, joinNonceKey, devNonceKey).Result()
	return err
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
		return eui, err
	}
	return eui, nil
}

//KeyToHex converts an AES128Key to its hex representation.
func KeyToHex(key lorawan.AES128Key) string {
	h := hex.EncodeToString(key[:])
	return h
}

//DevAddressToHex converts a device address to its hex representation.
func DevAddressToHex(devAddr [4]byte) string {
	return hex.EncodeToString(devAddr[:])
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
		log.Debugf("set uplink join mic error: %s", err)
		return err
	}

	_, err := joinPhy.MarshalText()
	if err != nil {
		log.Debugf("join marshal error: %s", err)
		return err
	}

	return nil
}

//publish publishes a message to the broker.
func publish(client MQTT.Client, topic string, v interface{}) error {

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

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

//GenerateFloat generates a byte array representation of a float.
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

//GenerateInt generates a byte array representation of an int.
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
