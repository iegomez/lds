package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/brocaar/chirpstack-api/go/gw"
	"github.com/brocaar/lorawan"
	lwBand "github.com/brocaar/lorawan/band"
	"github.com/golang/protobuf/ptypes"
	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"

	"github.com/iegomez/lds/lds"
)

// Marshalers, versions, and message types.
var (
	marshalers    = []string{"json", "protobuf", "v2_json"}
	majorVersions = map[lorawan.Major]string{0: "LoRaWANRev1"}
	macVersions   = map[lorawan.MACVersion]string{0: "LoRaWAN 1.0", 1: "LoRaWAN 1.1"}
	mTypes        = map[lorawan.MType]string{lorawan.UnconfirmedDataUp: "UnconfirmedDataUp", lorawan.ConfirmedDataUp: "ConfirmedDataUp"}
)

// lds device related vars.
var (
	cDevice        *lds.Device
	resetDevice    bool
	setRedisValues bool
)

// Frame counters and nonces.
var (
	ulFcntEdit     int
	dlFcntEdit     int
	devNonceEdit   int
	joinNonceEdit  int
	ulFcntEditS    string
	dlFcntEditS    string
	devNonceEditS  string
	joinNonceEditS string
)

type device struct {
	DevEUI        string             `toml:"eui"`
	DevAddress    string             `toml:"address"`
	NwkSEncKey    string             `toml:"network_session_encription_key"`
	SNwkSIntKey   string             `toml:"serving_network_session_integrity_key"`    //For Lorawan 1.0 this is the same as the NwkSEncKey
	FNwkSIntKey   string             `toml:"forwarding_network_session_integrity_key"` //For Lorawan 1.0 this is the same as the NwkSEncKey
	AppSKey       string             `toml:"application_session_key"`
	Marshaler     string             `toml:"marshaler"`
	NwkKey        string             `toml:"nwk_key"`  //Network key, used to be called application key for Lorawan 1.0
	AppKey        string             `toml:"app_key"`  //Application key, for Lorawan 1.1
	JoinEUI       string             `toml:"join_eui"` //JoinEUI for 1.1. (AppEUI on 1.0)
	Major         lorawan.Major      `toml:"-"`
	MACVersion    lorawan.MACVersion `toml:"mac_version"` //Lorawan MAC version
	MType         lorawan.MType      `toml:"-"`
	Profile       string             `toml:"profile"`
	Joined        bool               `toml:"joined"`
	SkipFCntCheck bool               `toml:"skip_fcnt_check"`
}

func beginDeviceForm() {
	//imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 205})
	//imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 435})
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
	imgui.InputTextV("Join EUI", &config.Device.JoinEUI, imgui.InputTextFlagsCharsHexadecimal|imgui.InputTextFlagsCallbackCharFilter, maxLength(config.Device.JoinEUI, 16))
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
		if imgui.SelectableV("UnconfirmedDataUp", config.Device.MType == lorawan.UnconfirmedDataUp || config.Device.MType == 0, 0, imgui.Vec2{}) {
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
	imgui.Checkbox("Disable frame counter validation", &config.Device.SkipFCntCheck)
	if imgui.Button("Join") {
		join()
	}
	imgui.SameLine()
	if cDevice != nil {
		if imgui.Button("Reset device") {
			resetDevice = true
		}
		imgui.SameLine()
		if imgui.Button("Set values") {
			setRedisValues = true
		}
	}
	beginReset()
	beginRedisValues()
	imgui.Separator()
	if cDevice != nil {
		imgui.Text(fmt.Sprintf("DlFCnt: %d - DevNonce:  %d", cDevice.DlFcnt, cDevice.DevNonce))
		imgui.Text(fmt.Sprintf("UlFCnt: %d - JoinNonce: %d", cDevice.UlFcnt, cDevice.JoinNonce))
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
		log.Errorf("devEUI error: %s", err)
		return
	}

	nwkHexKey := config.Device.NwkKey
	appHexKey := config.Device.AppKey

	appKey, err := lds.HexToKey(appHexKey)
	if err != nil {
		log.Errorf("appKey error: %s", err)
		return
	}
	nwkKey, err := lds.HexToKey(nwkHexKey)
	if err != nil {
		log.Errorf("nwkKey error: %s", err)
		return
	}
	joinEUI, err := lds.HexToEUI(config.Device.JoinEUI)
	if err != nil {
		log.Errorf("joinEUI error: %s", err)
		return
	}

	if cDevice == nil {
		cDevice = &lds.Device{
			DevEUI:        devEUI,
			DevAddr:       devAddr,
			NwkSEncKey:    nwkSEncKey,
			SNwkSIntKey:   sNwkSIntKey,
			FNwkSIntKey:   fNwkSIntKey,
			AppSKey:       appSKey,
			AppKey:        appKey,
			NwkKey:        nwkKey,
			JoinEUI:       joinEUI,
			Profile:       config.Device.Profile,
			Major:         lorawan.Major(config.Device.Major),
			MACVersion:    lorawan.MACVersion(config.Device.MACVersion),
			SkipFCntCheck: config.Device.SkipFCntCheck,
		}

		//Get redis info.
		if cDevice.GetInfo() {
			config.Device.NwkSEncKey = lds.KeyToHex(cDevice.NwkSEncKey)
			config.Device.FNwkSIntKey = lds.KeyToHex(cDevice.FNwkSIntKey)
			config.Device.SNwkSIntKey = lds.KeyToHex(cDevice.SNwkSIntKey)
			config.Device.AppSKey = lds.KeyToHex(cDevice.AppSKey)
			config.Device.DevAddress = lds.DevAddressToHex(cDevice.DevAddr)
			ulFcntEdit = int(cDevice.UlFcnt)
			ulFcntEditS = strconv.Itoa(ulFcntEdit)
			dlFcntEdit = int(cDevice.DlFcnt)
			dlFcntEditS = strconv.Itoa(dlFcntEdit)
			devNonceEdit = int(cDevice.DevNonce)
			devNonceEditS = strconv.Itoa(devNonceEdit)
			joinNonceEdit = int(cDevice.JoinNonce)
			joinNonceEditS = strconv.Itoa(joinNonceEdit)
		}
	} else {
		cDevice.DevEUI = devEUI
		cDevice.DevAddr = devAddr
		cDevice.NwkSEncKey = nwkSEncKey
		cDevice.SNwkSIntKey = sNwkSIntKey
		cDevice.FNwkSIntKey = fNwkSIntKey
		cDevice.AppSKey = appSKey
		cDevice.AppKey = appKey
		cDevice.NwkKey = nwkKey
		cDevice.JoinEUI = joinEUI
		cDevice.Profile = config.Device.Profile
		cDevice.Major = lorawan.Major(config.Device.Major)
		cDevice.MACVersion = lorawan.MACVersion(config.Device.MACVersion)
		cDevice.SkipFCntCheck = config.Device.SkipFCntCheck
	}
	cDevice.SetMarshaler(config.Device.Marshaler)
}

func beginReset() {
	if resetDevice {
		imgui.OpenPopup("Reset device")
		resetDevice = false
	}
	imgui.SetNextWindowPos(imgui.Vec2{X: float32(config.Window.Width-190) / 2, Y: float32(config.Window.Height-90) / 2})
	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 180})
	imgui.PushItemWidth(250.0)
	if imgui.BeginPopupModal("Reset device") {

		imgui.PushTextWrapPos()
		imgui.Text("This will delete saved devNonce, joinNonce, downlink and uplink frame counters, device address and device keys. Are you sure you want to proceed?")
		imgui.Separator()
		if imgui.Button("Cancel") {
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Confirm") {
			//Reset device.
			err := cDevice.Reset()
			if err != nil {
				log.Errorln(err)
			} else {
				setDevice()
				log.Infoln("device was reset")
			}
			imgui.CloseCurrentPopup()
			//Close popup.
		}
		imgui.EndPopup()
	}
}

func beginRedisValues() {
	if setRedisValues {
		imgui.OpenPopup("Set counters and nonces")
		setRedisValues = false
	}
	imgui.SetNextWindowPos(imgui.Vec2{X: float32(config.Window.Width-170) / 2, Y: float32(config.Window.Height-70) / 2})
	imgui.SetNextWindowSize(imgui.Vec2{X: 420, Y: 220})
	imgui.PushItemWidth(250.0)
	if imgui.BeginPopupModal("Set counters and nonces") {

		imgui.PushTextWrapPos()
		imgui.Text("Warning: this will only work when device is activated; when not, values will be reset on program start. Modifying these values may result in failure of communication.")
		imgui.InputTextV(fmt.Sprintf("DlFcnt    ##dlFcntEdit"), &dlFcntEditS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(dlFcntEditS, 10, &dlFcntEdit))
		imgui.InputTextV(fmt.Sprintf("UlFcnt    ##ulFcntEdit"), &ulFcntEditS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(ulFcntEditS, 10, &ulFcntEdit))
		imgui.InputTextV(fmt.Sprintf("DevNonce    ##devNonceEdit"), &devNonceEditS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(devNonceEditS, 10, &devNonceEdit))
		imgui.InputTextV(fmt.Sprintf("JoinNonce    ##joinNonceEdit"), &joinNonceEditS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(joinNonceEditS, 10, &joinNonceEdit))
		imgui.Separator()
		if imgui.Button("Cancel") {
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Save") {
			//Set values.
			err := cDevice.SetValues(ulFcntEdit, dlFcntEdit, devNonceEdit, joinNonceEdit)
			if err != nil {
				log.Errorln(err)
			}
			imgui.CloseCurrentPopup()
			//Close popup.
		}
		imgui.EndPopup()
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
		return
	}

	//Always set device to get any changes to the configuration.
	setDevice()

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
		Rssi:              int32(config.RXInfo.Rssi),
		LoraSnr:           float64(config.RXInfo.LoRaSNR),
		Channel:           uint32(config.RXInfo.Channel),
		RfChain:           uint32(config.RXInfo.RfChain),
		TimeSinceGpsEpoch: tsge,
		Time:              rxTime,
		Board:             0,
		Antenna:           0,
		Location:          nil,
		FineTimestamp:     nil,
		FineTimestampType: gw.FineTimestampType_NONE,
		Context:           make([]byte, 4),
	}

	lmi := &gw.LoRaModulationInfo{
		Bandwidth:       uint32(config.DR.Bandwidth),
		SpreadingFactor: uint32(config.DR.SpreadFactor),
		CodeRate:        config.RXInfo.CodeRate,
	}

	umi := &gw.UplinkTXInfo_LoraModulationInfo{
		LoraModulationInfo: lmi,
	}

	utx := gw.UplinkTXInfo{
		Frequency:      uint32(config.RXInfo.Frequency),
		ModulationInfo: umi,
	}

	err = cDevice.Join(mqttClient, config.MQTT.UplinkTopic, config.GW.MAC, &urx, &utx)

	if err != nil {
		log.Errorf("join error: %s", err)
	} else {
		log.Println("join sent")
	}
}

func run() {

	if !cNSClient.IsConnected() {
		if mqttClient == nil {
			err := connectClient()
			if err != nil {
				return
			}
		} else if !mqttClient.IsConnected() {
			log.Errorln("mqtt client not connected")
			return
		}
	}

	setDevice()

	/*dataRate := &lds.DataRate{
		Bandwidth:    config.DR.Bandwidth,
		Modulation:   "LORA",
		SpreadFactor: config.DR.SpreadFactor,
		BitRate:      config.DR.BitRate,
	}*/

	//Get DR index from a dr.
	dataRate := lwBand.DataRate{
		Modulation:   lwBand.Modulation("LORA"),
		SpreadFactor: config.DR.SpreadFactor,
		Bandwidth:    config.DR.Bandwidth,
		BitRate:      config.DR.BitRate,
	}

	running = true

	for {
		if stop {
			stop = false
			running = false
			return
		}
		payload := []byte{}
		var pErr error

		if config.RawPayload.UseRaw {
			payload, pErr = hex.DecodeString(config.RawPayload.Payload)
			if pErr != nil {
				log.Errorf("couldn't decode hex payload: %s", pErr)
				running = false
				return
			}
		} else if config.RawPayload.UseEncoder {
			payload, pErr = EncodeToBytes()
			if pErr != nil {
				log.Errorf("couldn't encode js object: %s", pErr)
				running = false
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

		gwID, err := lds.MACToGatewayID(config.GW.MAC)
		if err != nil {
			log.Errorf("gw mac error: %s", err)
			running = false
			return
		}
		now := time.Now()
		rxTime := ptypes.TimestampNow()
		tsge := ptypes.DurationProto(now.Sub(time.Time{}))

		urx := gw.UplinkRXInfo{
			GatewayId:         gwID,
			Rssi:              int32(config.RXInfo.Rssi),
			LoraSnr:           float64(config.RXInfo.LoRaSNR),
			Channel:           uint32(config.RXInfo.Channel),
			RfChain:           uint32(config.RXInfo.RfChain),
			TimeSinceGpsEpoch: tsge,
			Time:              rxTime,
			Board:             0,
			Antenna:           0,
			Location:          nil,
			FineTimestamp:     nil,
			FineTimestampType: gw.FineTimestampType_NONE,
			Context:           make([]byte, 4),
		}

		lmi := &gw.LoRaModulationInfo{
			Bandwidth:       uint32(config.DR.Bandwidth),
			SpreadingFactor: uint32(config.DR.SpreadFactor),
			CodeRate:        config.RXInfo.CodeRate,
		}

		umi := &gw.UplinkTXInfo_LoraModulationInfo{
			LoraModulationInfo: lmi,
		}

		utx := gw.UplinkTXInfo{
			Frequency:      uint32(config.RXInfo.Frequency),
			ModulationInfo: umi,
		}

		var fOpts []*lorawan.MACCommand
		for i := 0; i < len(macCommands); i++ {
			if macCommands[i].Use {
				fOpts = append(fOpts, &macCommands[i].MACCommand)
			}
		}

		//Now send an uplink
		var ulfc uint32

		if !cNSClient.IsConnected() {
			ulfc, err = cDevice.Uplink(mqttClient, config.MQTT.UplinkTopic, config.Device.MType, uint8(config.RawPayload.FPort), &urx, &utx, payload, config.GW.MAC, config.Band.Name, dataRate, fOpts, fCtrl)
		} else {
			ulfc, err = cDevice.UplinkUDP(cNSClient, config.Device.MType, uint8(config.RawPayload.FPort), &urx, &utx, payload, config.GW.MAC, config.Band.Name, dataRate, fOpts, fCtrl)
		}

		if err != nil {
			log.Errorf("couldn't send uplink: %s", err)
		} else {
			log.Infof("message sent, uplink framecounter is now %d", ulfc)
		}

		if !repeat || !running {
			stop = false
			running = false
			return
		}

		time.Sleep(time.Duration(interval) * time.Second)

	}
}

func onIncomingDownlink(payload []byte) error {
	log.Debugf("Incoming Downlink len=%d", len(payload))
	err := error(nil)
	if cDevice != nil {
		dlMessage, err := cDevice.ProcessDownlink(payload, cDevice.MACVersion)
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
		//Get redis info.
		cDevice.GetInfo()
	}
	return err
}
