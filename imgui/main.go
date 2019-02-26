package main

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/brocaar/lorawan"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"

	lwband "github.com/brocaar/lorawan/band"
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

type tomlConfig struct {
	MQTT       mqtt           `toml:"mqtt"`
	Band       band           `toml:"band"`
	Device     device         `timl:"device"`
	GW         gateway        `toml:"gateway"`
	DR         dataRate       `toml:"data_rate"`
	RXInfo     rxInfo         `toml:"rx_info"`
	DeviceData []*deviceDatum `toml:"default_data"`
	RawPayload rawPayload     `toml:"raw_payload"`
}

//deviceDatum holds optional default encoded data.
type deviceDatum struct {
	Name     string  `toml:"name"`
	Value    float64 `toml:"value"`
	MaxValue float64 `toml:"max_value"`
	MinValue float64 `toml:"min_value"`
	IsFloat  bool    `toml:"is_float"`
	NumBytes int     `toml:"num_bytes"`
	Index    int
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
var interval int

var output string

var str1 = "test"
var str2 = "prueba"
var str3 = "otro"
var testStrings = []*string{&str1, &str2, &str3}

func importConf() {

	if config == nil {
		cMqtt := mqtt{}

		cGw := gateway{}

		cDev := device{}

		cBand := band{}

		cDr := dataRate{}

		cRx := rxInfo{}

		dd := []*deviceDatum{}

		cPl := rawPayload{}

		config = &tomlConfig{
			MQTT:       cMqtt,
			Band:       cBand,
			Device:     cDev,
			GW:         cGw,
			DR:         cDr,
			RXInfo:     cRx,
			DeviceData: dd,
			RawPayload: cPl,
		}
	}

	if _, err := toml.DecodeFile(*confFile, &config); err != nil {
		log.Println(err)
		return
	}

	//Fill string representations of numeric values.
	config.DR.BitRateS = strconv.Itoa(config.DR.BitRate)
	config.RXInfo.ChannelS = strconv.Itoa(config.RXInfo.Channel)
	config.RXInfo.CrcStatusS = strconv.Itoa(config.RXInfo.CrcStatus)
	config.RXInfo.FrequencyS = strconv.Itoa(config.RXInfo.Frequency)
	config.RXInfo.LoRASNRS = strconv.FormatFloat(config.RXInfo.LoRaSNR, 'f', -1, 64)
	config.RXInfo.RfChainS = strconv.Itoa(config.RXInfo.RfChain)
	config.RXInfo.RssiS = strconv.Itoa(config.RXInfo.Rssi)

	//Set indexes of device data so we can delete fields.
	for i := 0; i < len(config.DeviceData); i++ {
		deviceDatum := config.DeviceData[i]
		deviceDatum.Index = i
		deviceDatum.ValueS = strconv.FormatFloat(deviceDatum.Value, 'f', -1, 64)
		deviceDatum.MaxValueS = strconv.FormatFloat(deviceDatum.MaxValue, 'f', -1, 64)
		deviceDatum.MinValueS = strconv.FormatFloat(deviceDatum.MinValue, 'f', -1, 64)
		deviceDatum.NumBytesS = strconv.Itoa(deviceDatum.NumBytes)
	}
}

func beginMQTTForm() {
	imgui.Begin("MQTT Configuration")
	imgui.Text("MQTT configuration")
	imgui.InputText("Server", &config.MQTT.Server)
	imgui.InputText("User", &config.MQTT.User)
	imgui.InputTextV("Password", &config.MQTT.Password, imgui.InputTextFlagsPassword, nil)
	imgui.Text("Gateway")
	imgui.InputText("MAC", &config.GW.MAC)
	imgui.Text(config.MQTT.Server)
	imgui.End()
}

func beginDeviceForm() {
	imgui.Begin("Device Configuration")
	imgui.InputTextV("Device EUI", &config.Device.EUI, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.EUI, 16))
	imgui.InputTextV("Device address", &config.Device.Address, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.Address, 8))
	imgui.InputTextV("Network session encryption key", &config.Device.NwkSEncKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.NwkSEncKey, 32))
	imgui.InputTextV("Serving network session integration key", &config.Device.SNwkSIntKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.SNwkSIntKey, 32))
	imgui.InputTextV("Forwarding network session integration key", &config.Device.FNwkSIntKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.FNwkSIntKey, 32))
	imgui.InputTextV("Application session key", &config.Device.AppSKey, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.AppSKey, 32))
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
	imgui.End()
}

func beginLoRaForm() {

	imgui.Begin("LoRa Configuration")
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

	imgui.Text(strconv.FormatFloat(config.RXInfo.LoRaSNR, 'f', -1, 64))

	imgui.End()
}

func beginDataForm() {
	imgui.Begin("Data")
	imgui.Text("Raw data")
	imgui.InputTextV("Raw bytes in hex", &config.RawPayload.Payload, imgui.InputTextFlagsCharsHexadecimal, nil)
	imgui.Checkbox("Send raw", &config.RawPayload.UseRaw)

	imgui.Text("Encoded data")

	/*for i := 0; i < len(config.DeviceData); i++ {
		deviceDatum := config.DeviceData[i]
		imgui.Text(strconv.Itoa(deviceDatum.Index))
		imgui.InputText("Name", &deviceDatum.Name)
		imgui.Checkbox("Float", &deviceDatum.IsFloat)
		imgui.InputTextV("Num bytes", &deviceDatum.NumBytesS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(deviceDatum.NumBytesS, 3, &deviceDatum.NumBytes))
		imgui.InputTextV("Value", &deviceDatum.ValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(deviceDatum.ValueS, &deviceDatum.Value))
		imgui.InputTextV("Max value", &deviceDatum.MaxValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(deviceDatum.MaxValueS, &deviceDatum.MaxValue))
		imgui.InputTextV("Min value", &deviceDatum.MinValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(deviceDatum.MinValueS, &deviceDatum.MinValue))
	}*/
	for i := 0; i < len(testStrings); i++ {
		imgui.InputText(fmt.Sprintf("test %d", i), testStrings[i])
	}

	/*
		//deviceDatum holds optional default encoded data.
		type deviceDatum struct {
			Name     string  `toml:"name"`
			Value    float64 `toml:"value"`
			MaxValue float64 `toml:"max_value"`
			MinValue float64 `toml:"min_value"`
			IsFloat  string  `toml:"is_float"`
			NumBytes int     `toml:"num_bytes"`
			Index    int
		}
	*/

	imgui.End()
}

func main() {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)

	window, err := glfw.CreateWindow(1280, 720, "ImGui-Go GLFW+OpenGL3 example", nil, nil)
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

func cb(data imgui.InputTextCallbackData) int32 {
	fmt.Printf("buff len: %d - char: %s - key: %d - flags: %d\n", len(string(data.Buffer())), string(data.EventChar()), data.EventKey(), data.EventFlag())
	if len(string(data.Buffer())) > 8 {
		data.SetEventChar(0)
		data.MarkBufferModified()
		return 1
	}
	return 0
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
