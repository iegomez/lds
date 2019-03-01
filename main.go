package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/protobuf/ptypes"
	"github.com/iegomez/lds/lds"
	"github.com/inkyblackness/imgui-go"

	lwband "github.com/brocaar/lorawan/band"
	paho "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

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
	DevEUI      string             `toml:"eui"`
	DevAddress  string             `toml:"address"`
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
	Profile     string             `toml:"profile"`
	Joined      bool               `toml:"joined"`
}

type dataRate struct {
	Bandwith     int `toml:"bandwith"`
	SpreadFactor int `toml:"spread_factor"`
	BitRate      int `toml:"bit_rate"`
	BitRateS     string
}

type rxInfo struct {
	Channel   int     `toml:"channel"`
	CodeRate  string  `toml:"code_rate"`
	CrcStatus int     `toml:"crc_status"`
	Frequency int     `toml:"frequency"`
	LoRaSNR   float64 `toml:"lora_snr"`
	RfChain   int     `toml:"rf_chain"`
	Rssi      int     `toml:"rssi"`
	//String representations for numeric values so that we can manage them with input texts.
	ChannelS   string
	CrcStatusS string
	FrequencyS string
	LoRASNRS   string
	RfChainS   string
	RssiS      string
}

type encodedType struct {
	Name     string  `toml:"name"`
	Value    float64 `toml:"value"`
	MaxValue float64 `toml:"max_value"`
	MinValue float64 `toml:"min_value"`
	IsFloat  bool    `toml:"is_float"`
	NumBytes int     `toml:"num_bytes"`
	//String representations.
	ValueS    string
	MinValueS string
	MaxValueS string
	NumBytesS string
}

//rawPayload holds optional raw bytes payload (hex encoded).
type rawPayload struct {
	Payload string `toml:"payload"`
	UseRaw  bool   `toml:"use_raw"`
}

type redisConf struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type tomlConfig struct {
	MQTT        mqtt           `toml:"mqtt"`
	Band        band           `toml:"band"`
	Device      device         `timl:"device"`
	GW          gateway        `toml:"gateway"`
	DR          dataRate       `toml:"data_rate"`
	RXInfo      rxInfo         `toml:"rx_info"`
	RawPayload  rawPayload     `toml:"raw_payload"`
	EncodedType []*encodedType `toml:"encoded_type"`
	LogLevel    string         `toml:"log_level"`
	RedisConf   redisConf      `toml:"redis"`
}

var confFile *string
var config *tomlConfig
var stop bool
var marshalers = []string{"json", "protobuf", "v2_json"}
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
var majorVersions = map[lorawan.Major]string{0: "LoRaWANRev1"}
var macVersions = map[lorawan.MACVersion]string{0: "LoRaWAN 1.0", 1: "LoRaWAN 1.1"}
var mTypes = map[lorawan.MType]string{lorawan.UnconfirmedDataUp: "UnconfirmedDataUp", lorawan.ConfirmedDataUp: "ConfirmedDataUp"}

var bandwidths = []int{50, 125, 250, 500}
var spreadFactors = []int{7, 8, 9, 10, 11, 12}

var sendOnce bool
var interval int32

type outputWriter struct {
	Text string
}

func (o *outputWriter) Write(p []byte) (n int, err error) {
	o.Text = fmt.Sprintf("%s%s", o.Text, string(p))
	return len(p), nil
}

var ow = &outputWriter{Text: ""}
var repeat bool
var running bool

var mqttClient paho.Client
var cDevice *lds.Device

func importConf() {

	if config == nil {
		cMqtt := mqtt{}

		cGw := gateway{}

		cDev := device{}

		cBand := band{}

		cDr := dataRate{}

		cRx := rxInfo{}

		cPl := rawPayload{}

		et := []*encodedType{}

		config = &tomlConfig{
			MQTT:        cMqtt,
			Band:        cBand,
			Device:      cDev,
			GW:          cGw,
			DR:          cDr,
			RXInfo:      cRx,
			RawPayload:  cPl,
			EncodedType: et,
		}
	}

	if _, err := toml.DecodeFile(*confFile, &config); err != nil {
		log.Println(err)
		return
	}

	l, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(l)
	}

	//Try to set redis.
	lds.StartRedis(config.RedisConf.Addr, config.RedisConf.Password, config.RedisConf.DB)

	for i := 0; i < len(config.EncodedType); i++ {
		config.EncodedType[i].ValueS = strconv.FormatFloat(config.EncodedType[i].Value, 'f', -1, 64)
		config.EncodedType[i].MaxValueS = strconv.FormatFloat(config.EncodedType[i].MaxValue, 'f', -1, 64)
		config.EncodedType[i].MinValueS = strconv.FormatFloat(config.EncodedType[i].MinValue, 'f', -1, 64)
		config.EncodedType[i].NumBytesS = strconv.Itoa(config.EncodedType[i].NumBytes)
	}

	//Fill string representations of numeric values.
	config.DR.BitRateS = strconv.Itoa(config.DR.BitRate)
	config.RXInfo.ChannelS = strconv.Itoa(config.RXInfo.Channel)
	config.RXInfo.CrcStatusS = strconv.Itoa(config.RXInfo.CrcStatus)
	config.RXInfo.FrequencyS = strconv.Itoa(config.RXInfo.Frequency)
	config.RXInfo.LoRASNRS = strconv.FormatFloat(config.RXInfo.LoRaSNR, 'f', -1, 64)
	config.RXInfo.RfChainS = strconv.Itoa(config.RXInfo.RfChain)
	config.RXInfo.RssiS = strconv.Itoa(config.RXInfo.Rssi)
}

func beginMQTTForm() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 10})
	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 180})
	imgui.Begin("MQTT & Gateway")
	imgui.PushItemWidth(250.0)
	imgui.InputText("Server", &config.MQTT.Server)
	imgui.InputText("User", &config.MQTT.User)
	imgui.InputTextV("Password", &config.MQTT.Password, imgui.InputTextFlagsPassword, nil)
	imgui.InputText("MAC", &config.GW.MAC)
	if imgui.Button("Connect") {
		connectClient()
	}
	if mqttClient != nil && mqttClient.IsConnected() {
		if imgui.Button("Disconnect") {
			mqttClient.Disconnect(200)
			log.Infoln("mqtt client disconnected")
		}
	}
	imgui.End()
}

func connectClient() error {
	//Connect to the broker
	opts := paho.NewClientOptions()
	opts.AddBroker(config.MQTT.Server)
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Password)
	opts.SetAutoReconnect(true)

	mqttClient = paho.NewClient(opts)
	log.Infoln("connecting...")
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("connection error: %s", token.Error())
		return token.Error()
	}
	log.Infoln("connection established")
	mqttClient.Subscribe(fmt.Sprintf("gateway/%s/tx", config.GW.MAC), 1, func(c paho.Client, msg paho.Message) {
		if cDevice != nil {
			dlMessage, err := cDevice.ProcessDownlink(msg.Payload(), cDevice.MACVersion)
			//Update keys when necessary.
			config.Device.AppSKey = lds.KeyToHex(cDevice.AppSKey)
			config.Device.FNwkSIntKey = lds.KeyToHex(cDevice.FNwkSIntKey)
			config.Device.NwkSEncKey = lds.KeyToHex(cDevice.NwkSEncKey)
			config.Device.SNwkSIntKey = lds.KeyToHex(cDevice.SNwkSIntKey)
			config.Device.DevAddress = lds.DevAddressToHex(cDevice.DevAddr)
			config.Device.Joined = cDevice.Joined
			if err != nil {
				log.Errorf("downlink error: %s", err)
			} else {
				log.Infof("received message: %s", dlMessage)
			}
		}
	})
	return nil
}

func beginDeviceForm() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 200})
	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 370})
	imgui.Begin("Device")
	imgui.PushItemWidth(250.0)
	imgui.InputTextV("Device EUI", &config.Device.DevEUI, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.DevEUI, 16))
	imgui.InputTextV("Device address", &config.Device.DevAddress, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.DevAddress, 8))
	imgui.InputTextV("NwkSEncKey", &config.Device.NwkSEncKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.NwkSEncKey, 32))
	imgui.InputTextV("SNwkSIntkey", &config.Device.SNwkSIntKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.SNwkSIntKey, 32))
	imgui.InputTextV("FNwkSIntKey", &config.Device.FNwkSIntKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.FNwkSIntKey, 32))
	imgui.InputTextV("AppSKey", &config.Device.AppSKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.AppSKey, 32))
	imgui.InputTextV("NwkKey", &config.Device.NwkKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.NwkKey, 32))
	imgui.InputTextV("AppKey", &config.Device.AppKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.AppKey, 32))
	if imgui.BeginCombo("Marshaler", config.Device.Marshaler) {
		for _, marshaler := range marshalers {
			if imgui.SelectableV(marshaler, marshaler == config.Device.Marshaler, 0, imgui.Vec2{}) {
				config.Device.Marshaler = marshaler
			}
		}
		imgui.EndCombo()
	}
	if imgui.BeginCombo("LoRaWAN major", majorVersions[config.Device.Major]) {
		if imgui.SelectableV("LoRaWANRev1", config.Device.Major == 0, 0, imgui.Vec2{}) {
			config.Device.MACVersion = 0
		}
		imgui.EndCombo()
	}
	if imgui.BeginCombo("MAC Version", macVersions[config.Device.MACVersion]) {

		if imgui.SelectableV("LoRaWAN 1.0", config.Device.MACVersion == 0, 0, imgui.Vec2{}) {
			config.Device.MACVersion = 0
		}
		if imgui.SelectableV("LoRaWAN 1.1", config.Device.MACVersion == 1, 0, imgui.Vec2{}) {
			config.Device.MACVersion = 1
		}
		imgui.EndCombo()
	}
	if imgui.BeginCombo("MType", mTypes[config.Device.MType]) {
		if imgui.SelectableV("UnconfirmedDataUp", config.Device.MType == lorawan.UnconfirmedDataUp, 0, imgui.Vec2{}) {
			config.Device.MType = lorawan.UnconfirmedDataUp
		}
		if imgui.SelectableV("ConfirmedDataUp", config.Device.MType == lorawan.ConfirmedDataUp, 0, imgui.Vec2{}) {
			config.Device.MType = lorawan.ConfirmedDataUp
		}
		imgui.EndCombo()
	}
	if imgui.BeginCombo("Profile", config.Device.Profile) {
		if imgui.SelectableV("OTAA", config.Device.Profile == "OTAA", 0, imgui.Vec2{}) {
			config.Device.Profile = "OTAA"
		}
		if imgui.SelectableV("ABP", config.Device.Profile == "ABP", 0, imgui.Vec2{}) {
			config.Device.Profile = "ABP"
		}
		imgui.EndCombo()
	}
	if cDevice == nil {
		if imgui.Button("Set device") {
			setDevice()
		}
	} else {
		if imgui.Button("Reset device") {
			setDevice()
		}
	}
	imgui.SameLine()
	if imgui.Button("Join") {
		join()
	}
	imgui.End()
}

func setDevice() {
	//Build your node with known keys (ABP).
	nwkSEncHexKey := config.Device.NwkSEncKey
	sNwkSIntHexKey := config.Device.SNwkSIntKey
	fNwkSIntHexKey := config.Device.FNwkSIntKey
	appSHexKey := config.Device.AppSKey
	devHexAddr := config.Device.DevAddress
	devAddr, err := lds.HexToDevAddress(devHexAddr)
	if err != nil {
		log.Errorf("dev addr error: %s", err)
	}

	nwkSEncKey, err := lds.HexToKey(nwkSEncHexKey)
	if err != nil {
		log.Errorf("nwkSEncKey error: %s", err)
	}

	sNwkSIntKey, err := lds.HexToKey(sNwkSIntHexKey)
	if err != nil {
		log.Errorf("sNwkSIntKey error: %s", err)
	}

	fNwkSIntKey, err := lds.HexToKey(fNwkSIntHexKey)
	if err != nil {
		log.Errorf("fNwkSIntKey error: %s", err)
	}

	appSKey, err := lds.HexToKey(appSHexKey)
	if err != nil {
		log.Errorf("appskey error: %s", err)
	}

	devEUI, err := lds.HexToEUI(config.Device.DevEUI)
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

	cDevice = &lds.Device{
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
	cDevice.SetMarshaler(config.Device.Marshaler)
	log.Infof("using marshaler: %s", config.Device.Marshaler)
}

func beginLoRaForm() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 580})
	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 350})
	imgui.Begin("LoRa Configuration")
	imgui.PushItemWidth(250.0)
	if imgui.BeginCombo("Band", string(config.Band.Name)) {
		for _, band := range bands {
			if imgui.SelectableV(string(band), band == config.Band.Name, 0, imgui.Vec2{}) {
				config.Band.Name = band
			}
		}
		imgui.EndCombo()
	}

	if imgui.BeginCombo("Bandwidth", strconv.Itoa(config.DR.Bandwith)) {
		for _, bandwidth := range bandwidths {
			if imgui.SelectableV(strconv.Itoa(bandwidth), bandwidth == config.DR.Bandwith, 0, imgui.Vec2{}) {
				config.DR.Bandwith = bandwidth
			}
		}
		imgui.EndCombo()
	}

	if imgui.BeginCombo("SpreadFactor", strconv.Itoa(config.DR.SpreadFactor)) {
		for _, sf := range spreadFactors {
			if imgui.SelectableV(strconv.Itoa(sf), sf == config.DR.SpreadFactor, 0, imgui.Vec2{}) {
				config.DR.SpreadFactor = sf
			}
		}
		imgui.EndCombo()
	}

	imgui.InputTextV("Bit rate", &config.DR.BitRateS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.DR.BitRateS, 6, &config.DR.BitRate))

	imgui.InputTextV("Channel", &config.RXInfo.ChannelS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.ChannelS, 10, &config.RXInfo.Channel))

	imgui.InputTextV("CrcStatus", &config.RXInfo.CrcStatusS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.CrcStatusS, 10, &config.RXInfo.CrcStatus))

	imgui.InputTextV("Frequency", &config.RXInfo.FrequencyS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.FrequencyS, 14, &config.RXInfo.Frequency))

	imgui.InputTextV("LoRaSNR", &config.RXInfo.LoRASNRS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.RXInfo.LoRASNRS, &config.RXInfo.LoRaSNR))

	imgui.InputTextV("RfChain", &config.RXInfo.RfChainS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(config.RXInfo.RfChainS, 10, &config.RXInfo.RfChain))

	imgui.InputTextV("Rssi", &config.RXInfo.RssiS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.RssiS, 10, &config.RXInfo.Rssi))

	imgui.End()
}

func beginDataForm() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 400, Y: 10})
	imgui.SetNextWindowSize(imgui.Vec2{X: 750, Y: 520})
	imgui.Begin("Data")
	imgui.Text("Raw data")
	imgui.PushItemWidth(150.0)
	imgui.InputTextV("Raw bytes in hex", &config.RawPayload.Payload, imgui.InputTextFlagsCharsHexadecimal, nil)
	imgui.SameLine()
	imgui.Checkbox("Send raw", &config.RawPayload.UseRaw)
	imgui.SliderInt("X", &interval, 1, 60)
	imgui.SameLine()
	imgui.Checkbox("Send every X seconds", &repeat)
	if !running {
		if imgui.Button("Send") {
			run()
		}
	}
	if repeat {
		if imgui.Button("Stop") {
			running = false
		}
	}

	imgui.Text("Encoded data")
	if imgui.Button("Add encoded type") {
		et := &encodedType{
			Name:      "New type",
			ValueS:    "0",
			MaxValueS: "0",
			MinValueS: "0",
			NumBytesS: "0",
		}
		config.EncodedType = append(config.EncodedType, et)
		log.Println("added new type")
	}

	for i := 0; i < len(config.EncodedType); i++ {
		delete := false
		imgui.Separator()
		imgui.InputText(fmt.Sprintf("Name     ##%d", i), &config.EncodedType[i].Name)
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Bytes    ##%d", i), &config.EncodedType[i].NumBytesS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(config.EncodedType[i].NumBytesS, 10, &config.EncodedType[i].NumBytes))
		imgui.SameLine()
		imgui.Checkbox(fmt.Sprintf("Float##%d", i), &config.EncodedType[i].IsFloat)
		imgui.SameLine()
		if imgui.Button(fmt.Sprintf("Delete##%d", i)) {
			delete = true
		}
		imgui.InputTextV(fmt.Sprintf("Value    ##%d", i), &config.EncodedType[i].ValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].ValueS, &config.EncodedType[i].Value))
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Max value##%d", i), &config.EncodedType[i].MaxValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].MaxValueS, &config.EncodedType[i].MaxValue))
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Min value##%d", i), &config.EncodedType[i].MinValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].MinValueS, &config.EncodedType[i].MinValue))
		if delete {
			if len(config.EncodedType) == 1 {
				config.EncodedType = make([]*encodedType, 0)
			} else {
				copy(config.EncodedType[i:], config.EncodedType[i+1:])
				config.EncodedType[len(config.EncodedType)-1] = &encodedType{}
				config.EncodedType = config.EncodedType[:len(config.EncodedType)-1]
			}
		}
	}
	imgui.Separator()

	imgui.End()
}

func beginOutput() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 400, Y: 540})
	imgui.SetNextWindowSize(imgui.Vec2{X: 750, Y: 350})
	imgui.Begin("Output")
	if imgui.Button("Clear") {
		ow.Text = ""
	}
	//imgui.PushStyleColor(imgui.StyleColorID(1), imgui.Vec4{X: 0.2, Y: 0.2, Z: 0.2, W: 0.5})
	imgui.PushTextWrapPos()
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.4, Y: 0.4, Z: 0.4, W: 0.5})
	imgui.Text(ow.Text)
	imgui.PopStyleColor()
	imgui.End()
}

func main() {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	log.SetOutput(ow)

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)

	window, err := glfw.CreateWindow(1200, 900, "LoRaServer ABP device simulator", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	err = gl.Init()
	if err != nil {
		panic(err)
	}

	context := imgui.CreateContext(nil)
	defer context.Destroy()

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()

	importConf()

	//imgui.CurrentStyle().ScaleAllSizes(2.0)
	//imgui.CurrentIO().SetFontGlobalScale(2.0)

	impl := imguiGlfw3Init(window)
	defer impl.Shutdown()

	var clearColor imgui.Vec4

	for !window.ShouldClose() {
		glfw.PollEvents()
		impl.NewFrame()
		beginMQTTForm()
		beginDeviceForm()
		beginLoRaForm()
		beginDataForm()
		beginOutput()
		displayWidth, displayHeight := window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(displayWidth), int32(displayHeight))
		gl.ClearColor(clearColor.X, clearColor.Y, clearColor.Z, clearColor.W)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		imgui.Render()
		impl.Render(imgui.RenderedDrawData())

		window.SwapBuffers()
		<-time.After(time.Millisecond * 25)
	}
}

func join() {
	if mqttClient == nil {
		err := connectClient()
		if err != nil {
			return
		}
	} else if !mqttClient.IsConnected() {
		log.Errorln("mqtt client not connected")
	}

	if cDevice == nil {
		setDevice()
	}

	dataRate := &lds.DataRate{
		Bandwidth:    config.DR.Bandwith,
		Modulation:   "LORA",
		SpreadFactor: config.DR.SpreadFactor,
		BitRate:      config.DR.BitRate,
	}

	rxInfo := &lds.RxInfo{
		Channel:   config.RXInfo.Channel,
		CodeRate:  config.RXInfo.CodeRate,
		CrcStatus: config.RXInfo.CrcStatus,
		DataRate:  dataRate,
		Frequency: config.RXInfo.Frequency,
		LoRaSNR:   float32(config.RXInfo.LoRaSNR),
		Mac:       config.GW.MAC,
		RfChain:   config.RXInfo.RfChain,
		Rssi:      config.RXInfo.Rssi,
		Time:      time.Now().Format(time.RFC3339),
		Timestamp: int32(time.Now().UnixNano() / 1000000000),
	}

	gwID, err := lds.MACToGatewayID(config.GW.MAC)
	if err != nil {
		log.Errorf("gw mac error: %s", err)
		return
	}

	err = cDevice.Join(mqttClient, string(gwID), *rxInfo)

	if err != nil {
		log.Errorf("join error: %s", err)
	} else {
		log.Println("join sent")
	}

}

func run() {

	if mqttClient == nil {
		err := connectClient()
		if err != nil {
			return
		}
	} else if !mqttClient.IsConnected() {
		log.Errorln("mqtt client not connected")
	}

	if cDevice == nil {
		setDevice()
	}

	dataRate := &lds.DataRate{
		Bandwidth:    config.DR.Bandwith,
		Modulation:   "LORA",
		SpreadFactor: config.DR.SpreadFactor,
		BitRate:      config.DR.BitRate,
	}

	for {
		if stop {
			stop = false
			return
		}
		payload := []byte{}

		if config.RawPayload.UseRaw {
			var pErr error
			payload, pErr = hex.DecodeString(config.RawPayload.Payload)
			if pErr != nil {
				log.Errorf("couldn't decode hex payload: %s", pErr)
				return
			}
		} else {
			for _, v := range config.EncodedType {
				if v.IsFloat {
					arr := lds.GenerateFloat(float32(v.Value), float32(v.MaxValue), int32(v.NumBytes))
					payload = append(payload, arr...)
				} else {
					arr := lds.GenerateInt(int32(v.Value), int32(v.NumBytes))
					payload = append(payload, arr...)
				}
			}
		}

		//Construct DataRate RxInfo with proper values according to your band (example is for US 915).

		rxInfo := &lds.RxInfo{
			Channel:   config.RXInfo.Channel,
			CodeRate:  config.RXInfo.CodeRate,
			CrcStatus: config.RXInfo.CrcStatus,
			DataRate:  dataRate,
			Frequency: config.RXInfo.Frequency,
			LoRaSNR:   float32(config.RXInfo.LoRaSNR),
			Mac:       config.GW.MAC,
			RfChain:   config.RXInfo.RfChain,
			Rssi:      config.RXInfo.Rssi,
			Size:      len(payload),
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
		}

		//////

		gwID, err := lds.MACToGatewayID(config.GW.MAC)
		if err != nil {
			log.Errorf("gw mac error: %s", err)
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

		//Now send an uplink
		ulfc, err := cDevice.Uplink(mqttClient, config.Device.MType, 1, &urx, &utx, payload, config.GW.MAC, config.Band.Name, *dataRate)
		if err != nil {
			log.Errorf("couldn't send uplink: %s", err)
		} else {
			log.Infof("message sent, uplink framecounter is now %d", ulfc)
		}

		if !repeat || !running {
			stop = false
			return
		}

		time.Sleep(time.Duration(interval) * time.Second)

	}

}

func maxLength(input string, limit int) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if len(input) >= limit {
			return 1
		}
		return 0
	}
}

func handleInt(input string, limit int, uValue *int) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
			if len(input) > limit || data.EventChar() == rune('.') {
				return 1
			}
			return 0
		}
		v, err := strconv.Atoi(input)
		if err == nil {
			*uValue = v
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleFloat32(input string, uValue *float32) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		v, err := strconv.ParseFloat(input, 32)
		if err == nil {
			*uValue = float32(v)
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleFloat64(input string, uValue *float64) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		v, err := strconv.ParseFloat(input, 64)
		if err == nil {
			*uValue = v
		} else {
			*uValue = 0
		}
		return 0
	}
}
