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
	Connected bool
	Server    string
	Port      int
}

type pfpacket struct {
	Time string  `json:"time"`
	TMST uint64  `json:"tmst"`
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

func (client *NSClient) send(bytes []byte) error {
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
	defer conn.Close()

	_, err = conn.Write(bytes)
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
	packet.TMST = toMilliseconds(gps) / 1000
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

	version := byte(0x02)
	token := rand.Int()
	tokenlsb := byte(token & 0x00FF)
	tokenmsb := byte((token & 0xFF00) >> 8)
	id := byte(0x00)
	header := []byte{version, tokenmsb, tokenlsb, id}

	gwbytes, err := hex.DecodeString(gwMAC)

	if err != nil {
		return err
	}

	jsonbytes := []byte(packetJSON)
	datagram := bytes.Join([][]byte{header, gwbytes, jsonbytes}, []byte{})

	client.send(datagram)

	return nil
}
