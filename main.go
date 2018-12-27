package main

import (
	"encoding/hex"
	"os"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/iegomez/lds/lds"

	"flag"
	"fmt"
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
	Server     string `toml:"server"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	uiServer   *ui.Entry
	uiUser     *ui.Entry
	uiPassword *ui.Entry
}

type gateway struct {
	MAC   string `toml:"mac"`
	uiMAC *ui.Entry
}

type band struct {
	Name   lwband.Name `toml:"name"`
	uiName *ui.Combobox
}

type device struct {
	EUI           string             `toml:"eui"`
	Address       string             `toml:"address"`
	NwkSEncKey    string             `toml:"network_session_encription_key"`
	SNwkSIntKey   string             `toml:"serving_network_session_integrity_key"`    //For Lorawan 1.0 this is the same as the NwkSEncKey
	FNwkSIntKey   string             `toml:"forwarding_network_session_integrity_key"` //For Lorawan 1.0 this is the same as the NwkSEncKey
	AppSKey       string             `toml:"application_session_key"`
	Marshaler     string             `toml:"marshaler"`
	NwkKey        string             `toml:"nwk_key"`     //Network key, used to be called application key for Lorawan 1.0
	AppKey        string             `toml:"app_key"`     //Application key, for Lorawan 1.1
	Major         lorawan.Major      `toml:"major"`       //Lorawan major version
	MACVersion    lorawan.MACVersion `toml:"mac_version"` //Lorawan MAC version
	Type          lorawan.MType      `toml:"mtype"`       //LoRaWAN mtype (ConfirmedDataUp or UnconfirmedDataUp)
	uiEUI         *ui.Entry
	uiAddress     *ui.Entry
	uiNwkSEncKey  *ui.Entry
	uiSNwkSIntKey *ui.Entry
	uiFNwkSIntKey *ui.Entry
	uiAppSKey     *ui.Entry
	uiMarshaler   *ui.Combobox
	uiNwkKey      *ui.Entry
	uiAppKey      *ui.Entry
	uiMajor       *ui.Combobox
	uiMACVersion  *ui.Combobox
	uiMType       *ui.Combobox
}

type dataRate struct {
	Bandwith       int `toml:"bandwith"`
	SpreadFactor   int `toml:"spread_factor"`
	BitRate        int `toml:"bit_rate"`
	uiBandwith     *ui.Entry
	uiSpreadFactor *ui.Entry
	uiBitRate      *ui.Entry
}

type rxInfo struct {
	Channel     int     `toml:"channel"`
	CodeRate    string  `toml:"code_rate"`
	CrcStatus   int     `toml:"crc_status"`
	Frequency   int     `toml:"frequency"`
	LoRaSNR     float64 `toml:"lora_snr"`
	RfChain     int     `toml:"rf_chain"`
	Rssi        int     `toml:"rssi"`
	uiChannel   *ui.Entry
	uiCodeRate  *ui.Entry
	uiCrcStatus *ui.Entry
	uiFrequency *ui.Entry
	uiLoRaSNR   *ui.Entry
	uiRfChain   *ui.Entry
	uiRssi      *ui.Entry
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
	Names []string        `toml:"names"`
	Data  [][]interface{} `toml:"data"`
	Types []string        `toml:"types"`
}

type SendableValue struct {
	value    *ui.Entry
	maxVal   *ui.Entry
	numBytes *ui.Entry
	isFloat  *ui.Checkbox
	index    int
	del      *ui.Button
	name     string
}

//rawPayload holds optional raw bytes payload (hex encoded).
type rawPayload struct {
	Payload   string `toml:"payload"`
	UseRaw    bool   `toml:"use_raw"`
	uiPayload *ui.Entry
	uiUseRaw  *ui.Checkbox
}

var confFile *string
var config *tomlConfig
var dataBox *ui.Box
var dataFormBox *ui.Box
var dataForm *ui.Form
var data []*SendableValue
var stop bool
var marshalers = map[int]string{0: "json", 1: "protobuf", 2: "v2_json"}
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
var uiSendOnce *ui.RadioButtons
var uiInterval *ui.Slider
var runBtn *ui.Button
var stopBtn *ui.Button

func importConf() {

	if config == nil {
		cMqtt := mqtt{
			uiServer:   ui.NewEntry(),
			uiUser:     ui.NewEntry(),
			uiPassword: ui.NewPasswordEntry(),
		}

		cDev := device{
			uiEUI:         ui.NewEntry(),
			uiAddress:     ui.NewEntry(),
			uiNwkSEncKey:  ui.NewEntry(),
			uiSNwkSIntKey: ui.NewEntry(),
			uiFNwkSIntKey: ui.NewEntry(),
			uiAppSKey:     ui.NewEntry(),
			uiMarshaler:   ui.NewCombobox(),
			uiNwkKey:      ui.NewEntry(),
			uiAppKey:      ui.NewEntry(),
			uiMajor:       ui.NewCombobox(),
			uiMACVersion:  ui.NewCombobox(),
			uiMType:       ui.NewCombobox(),
		}

		cGw := gateway{
			uiMAC: ui.NewEntry(),
		}

		cBand := band{
			uiName: ui.NewCombobox(),
		}

		cDr := dataRate{
			uiBandwith:     ui.NewEntry(),
			uiBitRate:      ui.NewEntry(),
			uiSpreadFactor: ui.NewEntry(),
		}

		cRx := rxInfo{
			uiChannel:   ui.NewEntry(),
			uiCodeRate:  ui.NewEntry(),
			uiCrcStatus: ui.NewEntry(),
			uiFrequency: ui.NewEntry(),
			uiLoRaSNR:   ui.NewEntry(),
			uiRfChain:   ui.NewEntry(),
			uiRssi:      ui.NewEntry(),
		}

		dd := defaultData{}

		cPl := rawPayload{
			uiPayload: ui.NewEntry(),
			uiUseRaw:  ui.NewCheckbox("Select to send raw bytes (hex encoded) instead of encoded data."),
		}

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

	config.MQTT.uiServer.SetText(config.MQTT.Server)
	config.MQTT.uiUser.SetText(config.MQTT.User)
	config.MQTT.uiPassword.SetText(config.MQTT.Password)

	config.GW.uiMAC.SetText(config.GW.MAC)

	for i := 0; i < len(bands); i++ {
		config.Band.uiName.Append(string(bands[i]))
		if config.Band.Name == bands[i] {
			config.Band.uiName.SetSelected(i)
		}
	}

	config.Device.uiEUI.SetText(config.Device.EUI)
	config.Device.uiAddress.SetText(config.Device.Address)
	config.Device.uiNwkSEncKey.SetText(config.Device.NwkSEncKey)
	config.Device.uiSNwkSIntKey.SetText(config.Device.SNwkSIntKey)
	config.Device.uiFNwkSIntKey.SetText(config.Device.FNwkSIntKey)
	config.Device.uiAppSKey.SetText(config.Device.AppSKey)
	config.Device.uiMarshaler.Append(marshalers[0])
	config.Device.uiMarshaler.Append(marshalers[1])
	config.Device.uiMarshaler.Append(marshalers[2])
	if config.Device.Marshaler == marshalers[0] {
		config.Device.uiMarshaler.SetSelected(0)
	} else if config.Device.Marshaler == marshalers[1] {
		config.Device.uiMarshaler.SetSelected(1)
	} else {
		config.Device.uiMarshaler.SetSelected(2)
	}
	config.Device.uiNwkKey.SetText(config.Device.NwkKey)
	config.Device.uiAppKey.SetText(config.Device.AppKey)

	config.Device.uiMajor.Append("LoRaWANRev1")
	config.Device.uiMajor.SetSelected(0)
	config.Device.Major = lorawan.LoRaWANR1

	config.Device.uiMACVersion.Append("LoRaWAN 1.0")
	config.Device.uiMACVersion.Append("LoRaWAN 1.1")

	config.Device.uiMACVersion.OnSelected(func(*ui.Combobox) {
		if config.Device.uiMACVersion.Selected() == 0 {
			config.Device.uiSNwkSIntKey.Hide()
			config.Device.uiFNwkSIntKey.Hide()
			config.Device.uiAppKey.Hide()
		} else {
			config.Device.uiSNwkSIntKey.Show()
			config.Device.uiFNwkSIntKey.Show()
			config.Device.uiAppKey.Show()
		}
	})

	if config.Device.MACVersion == 0 {
		config.Device.uiMACVersion.SetSelected(0)
	} else if config.Device.MACVersion == 1 {
		config.Device.uiMACVersion.SetSelected(1)
	}

	config.Device.uiMType.Append("UnconfirmedDataUp")
	config.Device.uiMType.Append("ConfirmedDataUp")
	config.Device.uiMType.SetSelected(0)

	config.DR.uiBandwith.SetText(fmt.Sprintf("%d", config.DR.Bandwith))
	config.DR.uiBitRate.SetText(fmt.Sprintf("%d", config.DR.BitRate))
	config.DR.uiSpreadFactor.SetText(fmt.Sprintf("%d", config.DR.SpreadFactor))

	config.RXInfo.uiChannel.SetText(fmt.Sprintf("%d", config.RXInfo.Channel))
	config.RXInfo.uiCodeRate.SetText(config.RXInfo.CodeRate)
	config.RXInfo.uiCrcStatus.SetText(fmt.Sprintf("%d", config.RXInfo.CrcStatus))
	config.RXInfo.uiFrequency.SetText(fmt.Sprintf("%d", config.RXInfo.Frequency))
	config.RXInfo.uiLoRaSNR.SetText(fmt.Sprintf("%f", config.RXInfo.LoRaSNR))
	config.RXInfo.uiRfChain.SetText(fmt.Sprintf("%d", config.RXInfo.RfChain))
	config.RXInfo.uiRssi.SetText(fmt.Sprintf("%d", config.RXInfo.Rssi))

	config.RawPayload.uiPayload.SetText(config.RawPayload.Payload)
	config.RawPayload.uiUseRaw.SetChecked(config.RawPayload.UseRaw)

}

func exportConf(filename string) {
	f, err := os.Create(fmt.Sprintf("%s.toml", filename))
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	encoder := toml.NewEncoder(f)
	err = encoder.Encode(config)
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	log.Infof("exported conf file %s", f.Name())
}

func makeMQTTForm() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)

	importBtn := ui.NewButton("Import conf")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	importBtn.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		if filename != "" {
			confFile = &filename
			importConf()
		}
	})

	exportFile := ui.NewEntry()
	exportBtn := ui.NewButton("Export conf")

	exportBtn.OnClicked(func(*ui.Button) {
		outName := exportFile.Text()
		if outName == "" {
			outName = fmt.Sprintf("export_conf_%d", time.Now().UnixNano())
		}
		exportConf(outName)
	})

	hbox.Append(importBtn, false)
	hbox.Append(exportFile, false)
	hbox.Append(exportBtn, false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("MQTT and Gateway configuration")
	group.SetMargined(true)
	vbox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())
	group.SetChild(entryForm)

	entryForm.Append("Server:", config.MQTT.uiServer, false)
	entryForm.Append("User:", config.MQTT.uiUser, false)
	entryForm.Append("Password:", config.MQTT.uiPassword, false)

	entryForm.Append("Gateway MAC", config.GW.uiMAC, false)
	return vbox
}

func makeDeviceForm() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("Device configuration")
	group.SetMargined(true)
	vbox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())
	group.SetChild(entryForm)

	entryForm.Append("DevEUI:", config.Device.uiEUI, false)
	entryForm.Append("Device address:", config.Device.uiAddress, false)
	entryForm.Append("Network session encryption key:", config.Device.uiNwkSEncKey, false)
	entryForm.Append("Serving network session integration key:", config.Device.uiSNwkSIntKey, false)
	entryForm.Append("Forwarding network session integration key:", config.Device.uiFNwkSIntKey, false)
	entryForm.Append("Application session key:", config.Device.uiAppSKey, false)
	entryForm.Append("NwkKey: ", config.Device.uiNwkKey, false)
	entryForm.Append("AppKey: ", config.Device.uiAppKey, false)
	entryForm.Append("Marshaler", config.Device.uiMarshaler, false)
	entryForm.Append("LoRaWAN major: ", config.Device.uiMajor, false)
	entryForm.Append("MAC Version: ", config.Device.uiMACVersion, false)
	entryForm.Append("MType: ", config.Device.uiMType, false)

	return vbox
}

func makeLoRaForm() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("Data Rate configuration")
	group.SetMargined(true)
	vbox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())
	group.SetChild(entryForm)

	entryForm.Append("Band: ", config.Band.uiName, false)
	entryForm.Append("Bandwidth: ", config.DR.uiBandwith, false)
	entryForm.Append("Bit rate: ", config.DR.uiBitRate, false)
	entryForm.Append("Spread factor: ", config.DR.uiSpreadFactor, false)

	entryFormRX := ui.NewForm()
	entryFormRX.SetPadded(true)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	groupRX := ui.NewGroup("RX info configuration")
	groupRX.SetMargined(true)
	vbox.Append(groupRX, true)

	groupRX.SetChild(ui.NewNonWrappingMultilineEntry())
	groupRX.SetChild(entryFormRX)

	entryFormRX.Append("Channel: ", config.RXInfo.uiChannel, false)
	entryFormRX.Append("Code rate: ", config.RXInfo.uiCodeRate, false)
	entryFormRX.Append("CRC status: ", config.RXInfo.uiCrcStatus, false)
	entryFormRX.Append("Frequency: ", config.RXInfo.uiFrequency, false)
	entryFormRX.Append("LoRa SNR: ", config.RXInfo.uiLoRaSNR, false)
	entryFormRX.Append("RF chain: ", config.RXInfo.uiRfChain, false)
	entryFormRX.Append("RSSI: ", config.RXInfo.uiRssi, false)

	return vbox
}

func makeDataForm() ui.Control {

	data = make([]*SendableValue, 0)

	dataBox := ui.NewVerticalBox()
	dataBox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	dataBox.Append(hbox, false)

	/*dataForm := ui.NewForm()
	dataForm.SetPadded(true)*/

	dataFormBox = ui.NewVerticalBox()
	dataFormBox.SetPadded(true)

	name := ui.NewEntry()

	button := ui.NewButton("Add value")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		v := &SendableValue{
			value:    ui.NewEntry(),
			maxVal:   ui.NewEntry(),
			numBytes: ui.NewEntry(),
			isFloat:  ui.NewCheckbox("Float"),
			index:    len(data),
			del:      ui.NewButton("Delete"),
			name:     name.Text(),
		}
		addValue(v, dataFormBox)
		name.SetText("")
	})

	if len(config.DefaultData.Names) == len(config.DefaultData.Data) {
		for i, name := range config.DefaultData.Names {
			valueData := config.DefaultData.Data[i]
			v := &SendableValue{
				value:    ui.NewEntry(),
				maxVal:   ui.NewEntry(),
				numBytes: ui.NewEntry(),
				isFloat:  ui.NewCheckbox("Float"),
				index:    i,
				del:      ui.NewButton("Delete"),
				name:     name,
			}
			if config.DefaultData.Types[i] == "float" {
				v.value.SetText(fmt.Sprintf("%f", valueData[0].(float64)))
				v.maxVal.SetText(fmt.Sprintf("%f", valueData[1].(float64)))
				v.numBytes.SetText(fmt.Sprintf("%d", int(valueData[2].(float64))))
				v.isFloat.SetChecked(true)
			} else {
				v.value.SetText(fmt.Sprintf("%d", valueData[0].(int64)))
				v.maxVal.SetText(fmt.Sprintf("%d", valueData[1].(int64)))
				v.numBytes.SetText(fmt.Sprintf("%d", valueData[2].(int64)))
				v.isFloat.SetChecked(false)
			}
			addValue(v, dataFormBox)
		}
	}

	runBtn = ui.NewButton("Run")
	entry2 := ui.NewEntry()
	entry2.SetReadOnly(true)

	stopBtn = ui.NewButton("Stop")
	entry3 := ui.NewEntry()
	entry3.SetReadOnly(true)

	runBtn.OnClicked(func(*ui.Button) {
		stopBtn.Enable()
		runBtn.Disable()
		go handledRun()
	})

	stopBtn.OnClicked(func(*ui.Button) {
		stopBtn.Disable()
		runBtn.Enable()
		stop = true
	})

	hbox.Append(name, false)
	hbox.Append(button, false)
	hbox.Append(runBtn, false)
	hbox.Append(stopBtn, false)

	dataBox.Append(ui.NewHorizontalSeparator(), false)

	confBox := ui.NewHorizontalBox()
	confBox.SetPadded(true)
	dataBox.Append(confBox, false)

	uiSendOnce = ui.NewRadioButtons()
	uiSendOnce.Append("Send once")
	uiSendOnce.Append("Send every X seconds")
	uiSendOnce.SetSelected(0)
	uiInterval = ui.NewSlider(1, 60)

	confBox.Append(uiSendOnce, false)
	confBox.Append(uiInterval, true)

	dataBox.Append(ui.NewHorizontalSeparator(), false)

	rawBox := ui.NewHorizontalBox()
	rawBox.SetPadded(true)
	dataBox.Append(rawBox, false)

	rawBox.Append(config.RawPayload.uiPayload, false)
	rawBox.Append(config.RawPayload.uiUseRaw, false)

	dataBox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("Encoded data")
	group.SetMargined(true)

	dataBox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())
	group.SetChild(dataFormBox)

	return dataBox
}

func addValue(v *SendableValue, dataFormBox *ui.Box) {
	data = append(data, v)

	//dataFormWrapper := ui.NewVerticalBox()
	dataForm := ui.NewHorizontalBox()
	nameLbl := ui.NewLabel(v.name)

	dataForm.Append(nameLbl, true)
	//dataFormWrapper.Append(dataForm, false)

	dataForm.SetPadded(true)
	vLbl := ui.NewLabel("Value")
	dataForm.Append(vLbl, false)
	dataForm.Append(v.value, false)
	mvLbl := ui.NewLabel("Max value")
	dataForm.Append(mvLbl, false)
	dataForm.Append(v.maxVal, false)
	nbLbl := ui.NewLabel("# bytes")
	dataForm.Append(nbLbl, false)
	dataForm.Append(v.numBytes, false)
	dataForm.Append(v.isFloat, false)
	dataForm.Append(v.del, true)
	dataFormBox.Append(dataForm, false)
	v.del.OnClicked(func(*ui.Button) {
		if len(data) == 1 {
			data = make([]*SendableValue, 0)
		} else {
			copy(data[v.index:], data[v.index+1:])
			data[len(data)-1] = &SendableValue{}
			data = data[:len(data)-1]
			for k := v.index; k < len(data); k++ {
				data[k].index--
			}
		}

		dataFormBox.Delete(v.index)

	})
}

func setupUI() {

	//Try to initialize default values.
	importConf()

	mainwin = ui.NewWindow("Loraserver device simulator", 600, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	tab := ui.NewTab()
	mainwin.SetChild(tab)
	mainwin.SetMargined(true)

	tab.Append("MQTT", makeMQTTForm())
	tab.SetMargined(0, true)
	tab.Append("Device", makeDeviceForm())
	tab.SetMargined(1, true)
	tab.Append("DR and RX info", makeLoRaForm())
	tab.SetMargined(2, true)
	tab.Append("Run", makeDataForm())
	tab.SetMargined(3, true)

	mainwin.Show()
}

func main() {

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()

	ui.Main(setupUI)
}

func handledRun() {
	ui.QueueMain(run)
}

func run() {

	//Connect to the broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.MQTT.uiServer.Text())
	opts.SetUsername(config.MQTT.uiUser.Text())
	opts.SetPassword(config.MQTT.uiPassword.Text())

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Connection error")
		log.Println(token.Error())
	}

	log.Println("Connection established.")

	//Build your node with known keys (ABP).
	nwkSEncHexKey := config.Device.uiNwkSEncKey.Text()
	sNwkSIntHexKey := config.Device.uiSNwkSIntKey.Text()
	fNwkSIntHexKey := config.Device.uiFNwkSIntKey.Text()
	appSHexKey := config.Device.uiAppSKey.Text()
	devHexAddr := config.Device.uiAddress.Text()
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

	devEUI, err := lds.HexToEUI(config.Device.uiEUI.Text())
	if err != nil {
		return
	}

	nwkHexKey := config.Device.uiNwkKey.Text()
	appHexKey := config.Device.uiAppKey.Text()

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
		Major:       lorawan.Major(config.Device.uiMajor.Selected()),
		MACVersion:  lorawan.MACVersion(config.Device.uiMACVersion.Selected()),
	}

	device.SetMarshaler(marshalers[config.Device.uiMarshaler.Selected()])
	log.Printf("using marshaler: %s\n", marshalers[config.Device.uiMarshaler.Selected()])

	bw, err := strconv.Atoi(config.DR.uiBandwith.Text())
	if err != nil {
		return
	}

	sf, err := strconv.Atoi(config.DR.uiSpreadFactor.Text())
	if err != nil {
		return
	}

	br, err := strconv.Atoi(config.DR.uiBitRate.Text())
	if err != nil {
		return
	}

	dataRate := &lds.DataRate{
		Bandwidth:    bw,
		Modulation:   "LORA",
		SpreadFactor: sf,
		BitRate:      br,
	}

	for {
		if stop {
			stop = false
			return
		}
		payload := []byte{}

		if config.RawPayload.uiUseRaw.Checked() {
			var pErr error
			payload, pErr = hex.DecodeString(config.RawPayload.uiPayload.Text())
			if err != nil {
				log.Errorf("couldn't decode hex payload: %s\n", pErr)
				return
			}
		} else {
			for _, v := range data {
				if v.isFloat.Checked() {
					val, err := strconv.ParseFloat(v.value.Text(), 32)
					if err != nil {
						log.Errorf("wrong conversion: %s\n", err)
						return
					}
					maxVal, err := strconv.ParseFloat(v.maxVal.Text(), 32)
					if err != nil {
						log.Errorf("wrong conversion: %s\n", err)
						return
					}
					numBytes, err := strconv.Atoi(v.numBytes.Text())
					if err != nil {
						log.Errorf("wrong conversion: %s\n", err)
						return
					}
					arr := lds.GenerateFloat(float32(val), float32(maxVal), int32(numBytes))
					payload = append(payload, arr...)
				} else {
					val, err := strconv.Atoi(v.value.Text())
					if err != nil {
						log.Errorf("wrong conversion: %s\n", err)
						return
					}

					numBytes, err := strconv.Atoi(v.numBytes.Text())
					if err != nil {
						log.Errorf("wrong conversion: %s\n", err)
						return
					}
					arr := lds.GenerateInt(int32(val), int32(numBytes))
					payload = append(payload, arr...)
				}
			}
		}

		log.Printf("Bytes: %v\n", payload)

		//Construct DataRate RxInfo with proper values according to your band (example is for US 915).

		channel, err := strconv.Atoi(config.RXInfo.uiChannel.Text())
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		crc, err := strconv.Atoi(config.RXInfo.uiCrcStatus.Text())
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		frequency, err := strconv.Atoi(config.RXInfo.uiFrequency.Text())
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		rfChain, err := strconv.Atoi(config.RXInfo.uiRfChain.Text())
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		rssi, err := strconv.Atoi(config.RXInfo.uiRssi.Text())
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		snr, err := strconv.ParseFloat(config.RXInfo.uiLoRaSNR.Text(), 64)
		if err != nil {
			log.Errorf("wrong conversion: %s\n", err)
			return
		}

		rxInfo := &lds.RxInfo{
			Channel:   channel,
			CodeRate:  config.RXInfo.uiCodeRate.Text(),
			CrcStatus: crc,
			DataRate:  dataRate,
			Frequency: frequency,
			LoRaSNR:   float32(snr),
			Mac:       config.GW.uiMAC.Text(),
			RfChain:   rfChain,
			Rssi:      rssi,
			Size:      len(payload),
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
		}

		//////

		gwID, err := lds.MACToGatewayID(config.GW.uiMAC.Text())
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
		if config.Device.uiMType.Selected() > 0 {
			mType = lorawan.ConfirmedDataUp
		}

		//Now send an uplink
		err = device.Uplink(client, mType, 1, &urx, &utx, payload, config.GW.uiMAC.Text(), bands[config.Band.uiName.Selected()], *dataRate)
		if err != nil {
			log.Printf("couldn't send uplink: %s\n", err)
		}

		if uiSendOnce.Selected() == 0 {
			stop = false
			//Let mqtt client publish first, then stop it.
			//time.Sleep(2 * time.Second)
			ui.QueueMain(stopBtn.Disable)
			ui.QueueMain(runBtn.Enable)
			return
		}

		time.Sleep(time.Duration(uiInterval.Value()) * time.Second)

	}

}
