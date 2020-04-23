package lds

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/brocaar/chirpstack-api/go/gw"
	"github.com/golang/protobuf/ptypes/duration"
	log "github.com/sirupsen/logrus"
)

// NSClient is a raw UDP client
type NSClient struct {
	Server string
	Port   int

	connected bool
	connexion *net.UDPConn
}

type pfpacket struct {
	Time string  `json:"time"`
	TMMS uint64  `json:"tmms"`
	TMST uint32  `json:"tmst"`
	Chan uint32  `json:"chan"`
	RFCH uint32  `json:"rfch"`
	Freq float32 `json:"freq"`
	Stat int32   `json:"stat"`
	Modu string  `json:"modu"`
	DatR string  `json:"datr"`
	CorR string  `json:"codr"`
	RSSI int32   `json:"rssi"`
	LSNR float64 `json:"lsnr"`
	Size uint32  `json:"size"`
	Data string  `json:"data"`
}

type pfproto struct {
	RXPK []pfpacket `json:"rxpk"`
}

// IsConnected checks if listening for incoming UDP
func (client *NSClient) IsConnected() bool {
	return client.connected
}

type udpPacketCallback func(payload []byte) error

// Connect starts listening incoming UDP
func (client *NSClient) Connect(gwMAC string, onReceive udpPacketCallback) error {

	ip := net.ParseIP(client.Server)

	if ip == nil {
		return errors.New("bad network server IP")
	}

	addr := net.UDPAddr{
		IP:   ip,
		Port: client.Port,
	}

	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		return err
	}
	client.connexion = conn

	log.Infoln("UDP listening bindpoint=%s", conn.LocalAddr())
	go client.receiveUDP(onReceive)
	go client.sendPullData(gwMAC)

	client.connected = true
	return nil
}

func (client *NSClient) receiveUDP(onReceive udpPacketCallback) {
	defer client.connexion.Close()
	buffer := make([]byte, 2048)

	for true {
		size, _, err := client.connexion.ReadFromUDP(buffer)

		if size <= 0 {
			log.Warningf("Incoming packet size %d", size)
			continue
		}

		if err != nil {
			log.Errorf("Unable to receive incoming packet %s", err)
			continue
		}

		message := buffer[0:size]
		onReceive(message)
	}
}

func createGWHeader(id byte, gwMAC string) ([]byte, error) {
	version := byte(0x02)
	token := rand.Int()
	tokenlsb := byte(token & 0x00FF)
	tokenmsb := byte((token & 0xFF00) >> 8)
	header := []byte{version, tokenmsb, tokenlsb, id}

	gwbytes, err := hex.DecodeString(gwMAC)

	if err != nil {
		return nil, err
	}

	gwheader := bytes.Join([][]byte{header, gwbytes}, []byte{})

	return gwheader, nil
}

func (client *NSClient) sendPullData(gwMAC string) {
	for true {
		datagram, err := createGWHeader(byte(0x02), gwMAC)

		if err == nil {
			log.Infoln("Sending PULL_DATA heartbeat")
			client.send(datagram)
		} else {
			log.Errorf("Unable to create PULL_DATA packet: %s", err)
		}

		time.Sleep(time.Second * 10)
	}
}

func (client *NSClient) send(bytes []byte) error {
	_, err := client.connexion.Write(bytes)
	return err
}

func toMilliseconds(d *duration.Duration) uint64 {
	return uint64(d.Seconds)*1000 + uint64(d.Nanos)/1000
}

func (client *NSClient) sendWithPayload(payload []byte, gwMAC string, rxInfo *gw.UplinkRXInfo, txInfo *gw.UplinkTXInfo) error {

	phyBase := base64.StdEncoding.EncodeToString(payload)

	gps := rxInfo.GetTimeSinceGpsEpoch()
	utc := time.Now().Format(time.RFC3339)
	mod := txInfo.GetLoraModulationInfo()

	packet := pfpacket{}
	packet.Time = utc
	packet.TMMS = toMilliseconds(gps) / 1000
	packet.TMST = uint32(toMilliseconds(gps) / 1000 / 1000)
	packet.Chan = rxInfo.GetChannel()
	packet.RFCH = rxInfo.GetRfChain()
	packet.Freq = float32(txInfo.GetFrequency()) / 1000000.0
	packet.Stat = 1
	packet.Modu = "LORA"
	packet.DatR = fmt.Sprintf("SF%dBW%d", mod.SpreadingFactor, mod.GetBandwidth())
	packet.CorR = mod.GetCodeRate()
	packet.RSSI = rxInfo.GetRssi()
	packet.LSNR = rxInfo.GetLoraSnr()
	packet.Size = uint32(len(payload))
	packet.Data = phyBase

	proto := pfproto{RXPK: []pfpacket{packet}}

	packetJSON, err := json.Marshal(proto)

	log.Debugf("Marshalled upstream JSON %s", packetJSON)

	if err != nil {
		return err
	}

	gwheader, err := createGWHeader(0x00, gwMAC)

	if err != nil {
		return err
	}

	jsonbytes := []byte(packetJSON)
	datagram := bytes.Join([][]byte{gwheader, jsonbytes}, []byte{})

	client.send(datagram)

	return nil
}

// UDPParsePacket extract metadata and physial payload from a packet
func UDPParsePacket(packet []byte, result *map[string]interface{}) (bool, string, error) {

	if len(packet) < 4 {
		log.Warningf("Bad incoming packet len %d", len(packet))
		return false, "", nil
	}

	version := int8(packet[0])

	if version != 0x02 {
		log.Warningf("Bad incoming version %d", version)
		return false, "", nil
	}

	token := int16(packet[1]) + int16(packet[2])<<8
	id := int8(packet[3])

	log.Debugf("Incoming message {%d, %d}", id, token)

	// PULL_RESP == 0x03
	if id != 0x03 {
		return false, "", nil
	}

	jsonBytes := packet[4:]
	log.Debugf("Incoming JSON %s", string(jsonBytes))

	*result = make(map[string]interface{})
	json.Unmarshal(jsonBytes, result)

	if (*result)["txpk"] == nil {
		log.Warningf("BAD JSON 'txpk'")
		return false, "", nil
	}

	txpk := (*result)["txpk"].(map[string]interface{})

	if txpk["data"] == nil {
		log.Warningf("BAD JSON 'data'")
		return false, "", nil
	}

	phyBase := txpk["data"].(string)

	return true, phyBase, nil
}
