package main

import (
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/iegomez/lds-gui/lds"

	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"

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

type device struct {
	EUI         string `toml:"eui"`
	Address     string `toml:"address"`
	NWSKey      string `toml:"network_session_key"`
	ASKey       string `toml:"application_session_key"`
	Marshaler   string `toml:"marshaler"`
	AppKey      string `toml:"app_key"`
	uiEUI       *ui.Entry
	uiAddress   *ui.Entry
	uiNWSKey    *ui.Entry
	uiASKey     *ui.Entry
	uiMarshaler *ui.Combobox
	uiAppKey    *ui.Entry
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
	Device      device      `timl:"device"`
	GW          gateway     `toml:"gateway"`
	DR          dataRate    `toml:"data_rate"`
	RXInfo      rxInfo      `toml:"rx_info"`
	DefaultData defaultData `toml:"default_data"`
}

//defaultData holds optional default data.
type defaultData struct {
	Names []string    `toml:"names"`
	Data  [][]float64 `toml:"data"`
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

var confFile string
var config tomlConfig
var dataBox *ui.Box
var dataForm *ui.Form
var data []SendableValue
var stop bool
var marshalers map[int]string

func checkConfig() {

	marshalers = map[int]string{0: "json", 1: "protobuf", 2: "v2_json"}

	cMqtt := mqtt{
		uiServer:   ui.NewEntry(),
		uiUser:     ui.NewEntry(),
		uiPassword: ui.NewPasswordEntry(),
	}

	cDev := device{
		uiEUI:       ui.NewEntry(),
		uiAddress:   ui.NewEntry(),
		uiNWSKey:    ui.NewEntry(),
		uiASKey:     ui.NewEntry(),
		uiMarshaler: ui.NewCombobox(),
		uiAppKey:    ui.NewEntry(),
	}

	cGw := gateway{
		uiMAC: ui.NewEntry(),
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

	config = tomlConfig{
		MQTT:        cMqtt,
		Device:      cDev,
		GW:          cGw,
		DR:          cDr,
		RXInfo:      cRx,
		DefaultData: dd,
	}

	if _, err := toml.DecodeFile(confFile, &config); err != nil {
		log.Println(err)
		return
	}

	config.MQTT.uiServer.SetText(config.MQTT.Server)
	config.MQTT.uiUser.SetText(config.MQTT.User)
	config.MQTT.uiPassword.SetText(config.MQTT.Password)

	config.GW.uiMAC.SetText(config.GW.MAC)

	config.Device.uiEUI.SetText(config.Device.EUI)
	config.Device.uiAddress.SetText(config.Device.Address)
	config.Device.uiNWSKey.SetText(config.Device.NWSKey)
	config.Device.uiASKey.SetText(config.Device.ASKey)
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
	config.Device.uiAppKey.SetText(config.Device.AppKey)

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

}

func makeMQTTForm() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)

	button := ui.NewButton("Open Configuration")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		if filename != "" {
			confFile = filename
			checkConfig()
		}
	})

	hbox.Append(button, false)

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
	entryForm.Append("Network session key:", config.Device.uiNWSKey, false)
	entryForm.Append("Application session key:", config.Device.uiASKey, false)
	entryForm.Append("Marshaler", config.Device.uiMarshaler, false)
	entryForm.Append("AppKey: ", config.Device.uiAppKey, false)

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

	data = make([]SendableValue, 0)

	dataBox := ui.NewVerticalBox()
	dataBox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	dataBox.Append(hbox, false)

	dataForm := ui.NewForm()
	dataForm.SetPadded(true)

	name := ui.NewEntry()

	button := ui.NewButton("Add value")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		v := SendableValue{
			value:    ui.NewEntry(),
			maxVal:   ui.NewEntry(),
			numBytes: ui.NewEntry(),
			isFloat:  ui.NewCheckbox("Float"),
			index:    len(data),
			del:      ui.NewButton("Delete"),
			name:     name.Text(),
		}
		addValue(v, dataForm)
		name.SetText("")
	})

	if len(config.DefaultData.Names) == len(config.DefaultData.Data) {
		for i, name := range config.DefaultData.Names {
			valueData := config.DefaultData.Data[i]
			v := SendableValue{
				value:    ui.NewEntry(),
				maxVal:   ui.NewEntry(),
				numBytes: ui.NewEntry(),
				isFloat:  ui.NewCheckbox("Float"),
				index:    i,
				del:      ui.NewButton("Delete"),
				name:     name,
			}
			v.value.SetText(fmt.Sprintf("%f", valueData[0]))
			v.maxVal.SetText(fmt.Sprintf("%f", valueData[1]))
			v.numBytes.SetText(fmt.Sprintf("%d", int(valueData[2])))
			v.isFloat.SetChecked(true)
			addValue(v, dataForm)
		}
	}

	runBtn := ui.NewButton("Run")
	entry2 := ui.NewEntry()
	entry2.SetReadOnly(true)

	stopBtn := ui.NewButton("Stop")
	entry3 := ui.NewEntry()
	entry3.SetReadOnly(true)

	runBtn.OnClicked(func(*ui.Button) {
		stopBtn.Enable()
		runBtn.Disable()
		go run()
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

	group := ui.NewGroup("Data")
	group.SetMargined(true)
	dataBox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())
	group.SetChild(dataForm)

	return dataBox
}

func addValue(v SendableValue, dataForm *ui.Form) {
	data = append(data, v)
	dataForm.Append(fmt.Sprintf("%s value", v.name), v.value, false)
	dataForm.Append(fmt.Sprintf("%s max value", v.name), v.maxVal, false)
	dataForm.Append(fmt.Sprintf("%s num bytes", v.name), v.numBytes, false)
	dataForm.Append(fmt.Sprintf("%s is float", v.name), v.isFloat, false)
	dataForm.Append("Delete", v.del, false)
	v.del.OnClicked(func(*ui.Button) {
		if len(data) == 1 {
			data = make([]SendableValue, 0)
		} else {
			copy(data[v.index:], data[v.index+1:])
			data[len(data)-1] = SendableValue{}
			data = data[:len(data)-1]
			for k := v.index; k < len(data); k++ {
				data[k].index--
			}
		}
		for i := 4; i >= 0; i-- {
			dataForm.Delete(5*v.index + i)
		}
	})
}

func setupUI() {
	//Set default conf file.
	confFile = "conf.toml"

	//Try to initialize default values.
	checkConfig()

	mainwin = ui.NewWindow("Loraserver device simulator", 640, 480, true)
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
	ui.Main(setupUI)
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
	nwsHexKey := config.Device.uiNWSKey.Text()
	appHexKey := config.Device.uiASKey.Text()
	devHexAddr := config.Device.uiAddress.Text()
	devAddr, err := lds.HexToDevAddress(devHexAddr)
	if err != nil {
		log.Printf("dev addr error: %s", err)
	}

	nwkSKey, err := lds.HexToKey(nwsHexKey)
	if err != nil {
		log.Printf("nwkskey error: %s", err)
	}

	appSKey, err := lds.HexToKey(appHexKey)
	if err != nil {
		log.Printf("appskey error: %s", err)
	}

	devEUI, err := lds.HexToEUI(config.Device.uiEUI.Text())
	if err != nil {
		return
	}

	appKey := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}

	device := &lds.Device{
		DevEUI:  devEUI,
		DevAddr: devAddr,
		NwkSKey: nwkSKey,
		AppSKey: appSKey,
		AppKey:  appKey,
		AppEUI:  appEUI,
		UlFcnt:  0,
		DlFcnt:  0,
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

		//Now send an uplink
		err = device.Uplink(client, lorawan.UnconfirmedDataUp, 1, &urx, &utx, payload, config.GW.uiMAC.Text())
		if err != nil {
			log.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(3 * time.Second)

	}

}

/*
func makeNumbersPage() ui.Control {
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)

	group := ui.NewGroup("Numbers")
	group.SetMargined(true)
	hbox.Append(group, true)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	group.SetChild(vbox)

	spinbox := ui.NewSpinbox(0, 100)
	slider := ui.NewSlider(0, 100)
	pbar := ui.NewProgressBar()
	spinbox.OnChanged(func(*ui.Spinbox) {
		slider.SetValue(spinbox.Value())
		pbar.SetValue(spinbox.Value())
	})
	slider.OnChanged(func(*ui.Slider) {
		spinbox.SetValue(slider.Value())
		pbar.SetValue(slider.Value())
	})
	vbox.Append(spinbox, false)
	vbox.Append(slider, false)
	vbox.Append(pbar, false)

	ip := ui.NewProgressBar()
	ip.SetValue(-1)
	vbox.Append(ip, false)

	group = ui.NewGroup("Lists")
	group.SetMargined(true)
	hbox.Append(group, true)

	vbox = ui.NewVerticalBox()
	vbox.SetPadded(true)
	group.SetChild(vbox)

	cbox := ui.NewCombobox()
	cbox.Append("Combobox Item 1")
	cbox.Append("Combobox Item 2")
	cbox.Append("Combobox Item 3")
	vbox.Append(cbox, false)

	ecbox := ui.NewEditableCombobox()
	ecbox.Append("Editable Item 1")
	ecbox.Append("Editable Item 2")
	ecbox.Append("Editable Item 3")
	vbox.Append(ecbox, false)

	rb := ui.NewRadioButtons()
	rb.Append("Radio Button 1")
	rb.Append("Radio Button 2")
	rb.Append("Radio Button 3")
	vbox.Append(rb, false)

	return hbox
}

func makeDataChoosersPage() ui.Control {
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	hbox.Append(vbox, false)

	vbox.Append(ui.NewDatePicker(), false)
	vbox.Append(ui.NewTimePicker(), false)
	vbox.Append(ui.NewDateTimePicker(), false)
	vbox.Append(ui.NewFontButton(), false)
	vbox.Append(ui.NewColorButton(), false)

	hbox.Append(ui.NewVerticalSeparator(), false)

	vbox = ui.NewVerticalBox()
	vbox.SetPadded(true)
	hbox.Append(vbox, true)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	vbox.Append(grid, false)

	button := ui.NewButton("Open File")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		if filename == "" {
			filename = "(cancelled)"
		}
		entry.SetText(filename)
	})
	grid.Append(button,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	button = ui.NewButton("Save File")
	entry2 := ui.NewEntry()
	entry2.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename := ui.SaveFile(mainwin)
		if filename == "" {
			filename = "(cancelled)"
		}
		entry2.SetText(filename)
	})
	grid.Append(button,
		0, 1, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry2,
		1, 1, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	msggrid := ui.NewGrid()
	msggrid.SetPadded(true)
	grid.Append(msggrid,
		0, 2, 2, 1,
		false, ui.AlignCenter, false, ui.AlignStart)

	button = ui.NewButton("Message Box")
	button.OnClicked(func(*ui.Button) {
		ui.MsgBox(mainwin,
			"This is a normal message box.",
			"More detailed information can be shown here.")
	})
	msggrid.Append(button,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	button = ui.NewButton("Error Box")
	button.OnClicked(func(*ui.Button) {
		ui.MsgBoxError(mainwin,
			"This message box describes an error.",
			"More detailed information can be shown here.")
	})
	msggrid.Append(button,
		1, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)

	return hbox
}
*/
