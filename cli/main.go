package main

import (
	"encoding/hex"
	"math/rand"

	"github.com/golang/protobuf/ptypes"
	"github.com/iegomez/lds/lds"

	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
	lwband "github.com/brocaar/lorawan/band"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

var mainwin *ui.Window

type mqtt struct {
	Server   string `toml:"server"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type gateway struct {
	MAC string `toml:"mac"`
}

type band struct {
	Name lwband.Name `toml:"name"`
}

type device struct {
	EUI         string             `toml:"eui"`
	Address     string             `toml:"address"`
	NwkSEncKey  string             `toml:"network_session_encription_key"`
	SNwkSIntKey string             `toml:"serving_network_session_integrity_key"`    //For Lorawan 1.0 this is the same as the NwkSEncKey
	FNwkSIntKey string             `toml:"forwarding_network_session_integrity_key"` //For Lorawan 1.0 this is the same as the NwkSEncKey
	AppSKey     string             `toml:"application_session_key"`
	Marshaler   string             `toml:"marshaler"`
	NwkKey      string             `toml:"nwk_key"`     //Network key, used to be called application key for Lorawan 1.0
	AppKey      string             `toml:"app_key"`     //Application key, for Lorawan 1.1
	Major       lorawan.Major      `toml:"major"`       //Lorawan major version
	MACVersion  lorawan.MACVersion `toml:"mac_version"` //Lorawan MAC version
	MType       lorawan.MType      `toml:"mtype"`       //LoRaWAN mtype (ConfirmedDataUp or UnconfirmedDataUp)
}

type dataRate struct {
	Bandwith     int `toml:"bandwith"`
	SpreadFactor int `toml:"spread_factor"`
	BitRate      int `toml:"bit_rate"`
}

type rxInfo struct {
	Channel   int     `toml:"channel"`
	CodeRate  string  `toml:"code_rate"`
	CrcStatus int     `toml:"crc_status"`
	Frequency int     `toml:"frequency"`
	LoRaSNR   float64 `toml:"lora_snr"`
	RfChain   int     `toml:"rf_chain"`
	Rssi      int     `toml:"rssi"`
}

type tomlConfig struct {
	MQTT        mqtt        `toml:"mqtt"`
	Band        band        `toml:"band"`
	Device      device      `timl:"device"`
	GW          gateway     `toml:"gateway"`
	DR          dataRate    `toml:"data_rate"`
	RXInfo      rxInfo      `toml:"rx_info"`
	DefaultData defaultData `toml:"default_data"`
	RawPayload  rawPayload  `toml:"raw_payload"`
}

//defaultData holds optional default encoded data.
type defaultData struct {
	Names    []string    `toml:"names"`
	Data     [][]float64 `toml:"data"`
	Interval int32       `toml:"interval"`
}

//rawPayload holds optional raw bytes payload (hex encoded).
type rawPayload struct {
	Payload string `toml:"payload"`
	UseRaw  bool   `toml:"use_raw"`
}

var confFile *string
var config *tomlConfig
var stop bool
var marshalers = map[string]int{"json": 0, "protobuf": 1, "v2_json": 2}
var bands = []lwband.Name{
	lwband.AS_923,
	lwband.AU_915_928,
	lwband.CN_470_510,
	lwband.CN_779_787,
	lwband.EU_433,
	lwband.EU_863_870,
	lwband.IN_865_867,
	lwband.KR_920_923,
	lwband.US_902_928,
	lwband.RU_864_870,
}
var sendOnce bool
var interval int

func importConf() {

	if config == nil {
		cMqtt := mqtt{}

		cDev := device{}

		cGw := gateway{}

		cBand := band{}

		cDr := dataRate{}

		cRx := rxInfo{}

		dd := defaultData{}

		cPl := rawPayload{}

		config = &tomlConfig{
			MQTT:        cMqtt,
			Band:        cBand,
			Device:      cDev,
			GW:          cGw,
			DR:          cDr,
			RXInfo:      cRx,
			DefaultData: dd,
			RawPayload:  cPl,
		}
	}

	if _, err := toml.DecodeFile(*confFile, &config); err != nil {
		log.Println(err)
		return
	}
}

func main() {

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()
	importConf()
	run()
}

func run() {

	//Connect to the broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.MQTT.Server)
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Password)

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Connection error")
		log.Println(token.Error())
	}

	log.Println("Connection established.")

	//Build your node with known keys (ABP).
	nwkSEncHexKey := config.Device.NwkSEncKey
	sNwkSIntHexKey := config.Device.SNwkSIntKey
	fNwkSIntHexKey := config.Device.FNwkSIntKey
	appSHexKey := config.Device.AppSKey
	devHexAddr := config.Device.Address
	devAddr, err := lds.HexToDevAddress(devHexAddr)
	if err != nil {
		log.Printf("dev addr error: %s", err)
	}

	nwkSEncKey, err := lds.HexToKey(nwkSEncHexKey)
	if err != nil {
		log.Printf("nwkSEncKey error: %s", err)
	}

	sNwkSIntKey, err := lds.HexToKey(sNwkSIntHexKey)
	if err != nil {
		log.Printf("sNwkSIntKey error: %s", err)
	}

	fNwkSIntKey, err := lds.HexToKey(fNwkSIntHexKey)
	if err != nil {
		log.Printf("fNwkSIntKey error: %s", err)
	}

	appSKey, err := lds.HexToKey(appSHexKey)
	if err != nil {
		log.Printf("appskey error: %s", err)
	}

	devEUI, err := lds.HexToEUI(config.Device.EUI)
	if err != nil {
		return
	}

	nwkHexKey := config.Device.NwkKey
	appHexKey := config.Device.AppKey

	appKey, err := lds.HexToKey(appHexKey)
	if err != nil {
		return
	}
	nwkKey, err := lds.HexToKey(nwkHexKey)
	if err != nil {
		return
	}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}

	device := &lds.Device{
		DevEUI:      devEUI,
		DevAddr:     devAddr,
		NwkSEncKey:  nwkSEncKey,
		SNwkSIntKey: sNwkSIntKey,
		FNwkSIntKey: fNwkSIntKey,
		AppSKey:     appSKey,
		AppKey:      appKey,
		NwkKey:      nwkKey,
		AppEUI:      appEUI,
		UlFcnt:      0,
		DlFcnt:      0,
		Major:       lorawan.Major(config.Device.Major),
		MACVersion:  lorawan.MACVersion(config.Device.MACVersion),
	}

	device.SetMarshaler(config.Device.Marshaler)

	dataRate := &lds.DataRate{
		Bandwidth:    config.DR.Bandwith,
		Modulation:   "LORA",
		SpreadFactor: config.DR.SpreadFactor,
		BitRate:      config.DR.BitRate,
	}

	mult := 1

	for {
		if stop {
			stop = false
			return
		}
		payload := []byte{}

		if config.RawPayload.UseRaw {
			var pErr error
			payload, pErr = hex.DecodeString(config.RawPayload.Payload)
			if err != nil {
				log.Errorf("couldn't decode hex payload: %s\n", pErr)
				return
			}
		} else {
			for _, v := range config.DefaultData.Data {
				rand.Seed(time.Now().UnixNano() / 10000)
				if rand.Intn(10) < 5 {
					mult *= -1
				}
				arr := lds.GenerateFloat(float32(mult)*float32(v[0]+rand.Float64()/100), float32(v[1]), int32(v[2]))
				payload = append(payload, arr...)

			}
		}

		log.Printf("Bytes: %v\n", payload)

		rxInfo := &lds.RxInfo{
			Channel:   config.RXInfo.Channel,
			CodeRate:  config.RXInfo.CodeRate,
			CrcStatus: config.RXInfo.CrcStatus,
			DataRate:  dataRate,
			Frequency: config.RXInfo.Frequency,
			LoRaSNR:   float32(config.RXInfo.LoRaSNR),
			Mac:       config.GW.MAC,
			RfChain:   config.RXInfo.RfChain,
			Rssi:      config.RXInfo.RfChain,
			Size:      len(payload),
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
		}

		//////

		gwID, err := lds.MACToGatewayID(config.GW.MAC)
		if err != nil {
			log.Errorf("gw mac error: %s\n", err)
			return
		}
		now := time.Now()
		rxTime := ptypes.TimestampNow()
		tsge := ptypes.DurationProto(now.Sub(time.Time{}))

		urx := gw.UplinkRXInfo{
			GatewayId:         gwID,
			Rssi:              int32(rxInfo.Rssi),
			LoraSnr:           float64(rxInfo.LoRaSNR),
			Channel:           uint32(rxInfo.Channel),
			RfChain:           uint32(rxInfo.RfChain),
			TimeSinceGpsEpoch: tsge,
			Time:              rxTime,
			Timestamp:         uint32(rxTime.GetSeconds()),
			Board:             0,
			Antenna:           0,
			Location:          nil,
			FineTimestamp:     nil,
			FineTimestampType: gw.FineTimestampType_NONE,
		}

		lmi := &gw.LoRaModulationInfo{
			Bandwidth:       uint32(rxInfo.DataRate.Bandwidth),
			SpreadingFactor: uint32(rxInfo.DataRate.SpreadFactor),
			CodeRate:        rxInfo.CodeRate,
		}

		umi := &gw.UplinkTXInfo_LoraModulationInfo{
			LoraModulationInfo: lmi,
		}

		utx := gw.UplinkTXInfo{
			Frequency:      uint32(rxInfo.Frequency),
			ModulationInfo: umi,
		}

		//////
		mType := lorawan.UnconfirmedDataUp
		if config.Device.MType > 0 {
			mType = lorawan.ConfirmedDataUp
		}

		//Now send an uplink
		err = device.Uplink(client, mType, 1, &urx, &utx, payload, config.GW.MAC, config.Band.Name, *dataRate)
		if err != nil {
			log.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(time.Duration(config.DefaultData.Interval) * time.Second)

	}

}
