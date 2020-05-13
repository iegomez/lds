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
	log "github.com/sirupsen/logrus"

	"github.com/iegomez/lds/lds"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/scartill/giox"
	xmat "github.com/scartill/giox/material"
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
	cDevice *lds.Device
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

// Widgets
var (
	deviceEUIEdit      widget.Editor
	deviceAddressEdit  widget.Editor
	nwkSEncKeyEdit     widget.Editor
	sNwkSIntKeyEdit    widget.Editor
	fNwkSIntKeyEdit    widget.Editor
	appSKeyEdit        widget.Editor
	nwkKeyEdit         widget.Editor
	appKeyEdit         widget.Editor
	joinEUIEdit        widget.Editor
	marshalerCombo     giox.Combo
	majorVersionCombo  giox.Combo
	macVersionCombo    giox.Combo
	mTypeCombo         giox.Combo
	profileCombo       giox.Combo
	disableFCWCheckbox widget.Bool
	joinButton         widget.Button
	resetButton        widget.Button
	setValuesButton    widget.Button

	ulFcntEdit    widget.Editor
	dlFcntEdit    widget.Editor
	devNonceEdit  widget.Editor
	joinNonceEdit widget.Editor

	resetDevice        bool
	resetCancelButton  widget.Button
	resetConfirmButton widget.Button

	setRedisValues              bool
	setRedisValuesCancelButton  widget.Button
	setRedisValuesConfirmButton widget.Button
)

func createDeviceForm() {
	marshalerItems := make([]string, len(marshalers))
	for i, v := range marshalers {
		marshalerItems[i] = string(v)
	}
	marshalerCombo = giox.MakeCombo(marshalerItems, "<select marshaler>")

	ki := 0
	majorVersionItems := make([]string, len(majorVersions))
	for _, v := range majorVersions {
		majorVersionItems[ki] = string(v)
		ki++
	}
	majorVersionCombo = giox.MakeCombo(majorVersionItems, "<select major version>")

	ki = 0
	macVersionItems := make([]string, len(macVersions))
	for _, v := range macVersions {
		macVersionItems[ki] = string(v)
		ki++
	}
	macVersionCombo = giox.MakeCombo(macVersionItems, "<select MAC version>")

	ki = 0
	mTypeItems := make([]string, len(mTypes))
	for _, v := range mTypes {
		mTypeItems[ki] = string(v)
		ki++
	}
	mTypeCombo = giox.MakeCombo(mTypeItems, "<select message type>")

	profileCombo = giox.MakeCombo([]string{"OTAA", "ABP"}, "<select profile>")
}

func deviceResetGuiValues() {
	deviceEUIEdit.SetText(config.Device.DevEUI)
	deviceAddressEdit.SetText(config.Device.DevAddress)
	nwkSEncKeyEdit.SetText(config.Device.NwkSEncKey)
	sNwkSIntKeyEdit.SetText(config.Device.SNwkSIntKey)
	fNwkSIntKeyEdit.SetText(config.Device.FNwkSIntKey)
	appSKeyEdit.SetText(config.Device.AppSKey)
	nwkKeyEdit.SetText(config.Device.NwkKey)
	appKeyEdit.SetText(config.Device.AppKey)
	joinEUIEdit.SetText(config.Device.JoinEUI)
	marshalerCombo.SelectItem(string(config.Device.Marshaler))
	majorVersionCombo.SelectItem(majorVersions[config.Device.Major])
	macVersionCombo.SelectItem(macVersions[config.Device.MACVersion])
	mTypeCombo.SelectItem(mTypes[config.Device.MType])
	profileCombo.SelectItem(config.Device.Profile)
	disableFCWCheckbox.Value = config.Device.SkipFCntCheck
}

func deviceForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
	config.Device.DevEUI = deviceEUIEdit.Text()
	config.Device.DevAddress = deviceAddressEdit.Text()
	config.Device.NwkSEncKey = nwkSEncKeyEdit.Text()
	config.Device.SNwkSIntKey = sNwkSIntKeyEdit.Text()
	config.Device.FNwkSIntKey = fNwkSIntKeyEdit.Text()
	config.Device.AppSKey = appSKeyEdit.Text()
	config.Device.NwkKey = nwkKeyEdit.Text()
	config.Device.AppKey = appKeyEdit.Text()
	config.Device.JoinEUI = joinEUIEdit.Text()

	config.Device.Marshaler = marshalerCombo.SelectedText()

	config.Device.Major = 0
	if majorVersionCombo.HasSelected() {
		for k, v := range majorVersions {
			if majorVersionCombo.SelectedText() == string(v) {
				config.Device.Major = k
			}
		}
	}

	config.Device.MACVersion = 0
	if macVersionCombo.HasSelected() {
		for k, v := range macVersions {
			if macVersionCombo.SelectedText() == string(v) {
				config.Device.MACVersion = k
			}
		}
	}

	config.Device.MType = lorawan.UnconfirmedDataUp
	if mTypeCombo.HasSelected() {
		for k, v := range mTypes {
			if mTypeCombo.SelectedText() == string(v) {
				config.Device.MType = k
			}
		}
	}

	config.Device.Profile = "OTAA"
	if profileCombo.HasSelected() {
		config.Device.Profile = profileCombo.SelectedText()
	}

	config.Device.SkipFCntCheck = disableFCWCheckbox.Value

	for joinButton.Clicked(gtx) {
		join()
	}

	for resetButton.Clicked(gtx) {
		resetDevice = true
	}

	for setValuesButton.Clicked(gtx) {
		setRedisValues = true
	}

	if resetDevice {
		if ok, subform := resetDeviceSubform(gtx, th); ok {
			return subform
		}
	}

	if setRedisValues {
		if ok, subform := setRedisValuesSubform(gtx, th); ok {
			return subform
		}
	}

	widgets := []layout.FlexChild{
		xmat.RigidSection(gtx, th, "Device"),
		xmat.RigidEditor(gtx, th, "DevEUI", "<device EUI>", &deviceEUIEdit),
		xmat.RigidEditor(gtx, th, "DevAddress", "device address", &deviceAddressEdit),
		xmat.RigidEditor(gtx, th, "NwkSEncKey", "network session key", &nwkSEncKeyEdit),
		xmat.RigidEditor(gtx, th, "SNwkSIntKey", "network session key", &sNwkSIntKeyEdit),
		xmat.RigidEditor(gtx, th, "FNwkSIntKey", "forward session key", &fNwkSIntKeyEdit),
		xmat.RigidEditor(gtx, th, "AppSKey", "application session key", &appSKeyEdit),
		xmat.RigidEditor(gtx, th, "NwkKey", "network key", &nwkKeyEdit),
		xmat.RigidEditor(gtx, th, "AppKey", "application key", &appKeyEdit),
		xmat.RigidEditor(gtx, th, "JoinEUI", "join EUI", &joinEUIEdit),
	}

	comboOpen := marshalerCombo.IsExpanded() ||
		majorVersionCombo.IsExpanded() ||
		macVersionCombo.IsExpanded() ||
		mTypeCombo.IsExpanded() ||
		profileCombo.IsExpanded()

	if !comboOpen || marshalerCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(gtx, th, "Marshaler", &marshalerCombo))
	}

	if !comboOpen || majorVersionCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(gtx, th, "LoRaWAN Major", &majorVersionCombo))
	}

	if !comboOpen || macVersionCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(gtx, th, "MAC Version", &macVersionCombo))
	}

	if !comboOpen || mTypeCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(gtx, th, "MType", &mTypeCombo))
	}

	if !comboOpen || profileCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(gtx, th, "Profile", &profileCombo))
	}

	if !comboOpen {
		widgets = append(widgets, []layout.FlexChild{
			xmat.RigidCheckBox(gtx, th, "Disable frame counter validation", &disableFCWCheckbox),
			xmat.RigidButton(gtx, th, "Join", &joinButton),
		}...)

		if cDevice != nil {
			widgets = append(widgets, []layout.FlexChild{
				xmat.RigidButton(gtx, th, "Reset device", &resetButton),
				xmat.RigidButton(gtx, th, "Set values", &setValuesButton),
			}...)
		}

		if cDevice != nil {
			widgets = append(widgets, []layout.FlexChild{
				xmat.RigidLabel(gtx, th, fmt.Sprintf("DlFCnt: %d - DevNonce:  %d", cDevice.DlFcnt, cDevice.DevNonce)),
				xmat.RigidLabel(gtx, th, fmt.Sprintf("UlFCnt: %d - JoinNonce: %d", cDevice.UlFcnt, cDevice.JoinNonce)),
				xmat.RigidLabel(gtx, th, fmt.Sprintf("Joined: %t", cDevice.Joined)),
			}...)
		}
	}

	inset := layout.Inset{Top: unit.Px(20)}
	return layout.Rigid(func() {
		inset.Layout(gtx, func() {
			layout.Flex{Axis: layout.Vertical}.Layout(gtx, widgets...)
		})
	})
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
			ulFcnt := int(cDevice.UlFcnt)
			dlFcnt := int(cDevice.DlFcnt)
			devNonce := int(cDevice.DevNonce)
			joinNonce := int(cDevice.JoinNonce)
			ulFcntEdit.SetText(strconv.Itoa(ulFcnt))
			dlFcntEdit.SetText(strconv.Itoa(dlFcnt))
			devNonceEdit.SetText(strconv.Itoa(devNonce))
			joinNonceEdit.SetText(strconv.Itoa(joinNonce))
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

func resetDeviceSubform(gtx *layout.Context, th *material.Theme) (bool, layout.FlexChild) {

	for resetCancelButton.Clicked(gtx) {
		resetDevice = false
		return false, layout.FlexChild{}
	}

	for resetConfirmButton.Clicked(gtx) {
		//Reset device.
		err := cDevice.Reset()
		if err != nil {
			log.Errorln(err)
		} else {
			setDevice()
			log.Warningln("Device was reset")
		}
		resetDevice = false
		return false, layout.FlexChild{}
	}

	widgets := []layout.FlexChild{
		xmat.RigidSection(gtx, th, "This will delete saved devNonce, joinNonce, downlink and uplink frame counters,\ndevice address and device keys.\nAre you sure you want to proceed?"),
		xmat.RigidButton(gtx, th, "Cancel", &resetCancelButton),
		xmat.RigidButton(gtx, th, "Confirm", &resetConfirmButton),
	}

	inset := layout.Inset{Top: unit.Px(20)}
	return true, layout.Rigid(func() {
		inset.Layout(gtx, func() {
			layout.Flex{Axis: layout.Vertical}.Layout(gtx, widgets...)
		})
	})
}

func setRedisValuesSubform(gtx *layout.Context, th *material.Theme) (bool, layout.FlexChild) {
	ulFcntEdit.SetText(strconv.FormatUint(uint64(cDevice.UlFcnt), 10))
	dlFcntEdit.SetText(strconv.FormatUint(uint64(cDevice.DlFcnt), 10))
	devNonceEdit.SetText(strconv.FormatUint(uint64(cDevice.DevNonce), 10))
	joinNonceEdit.SetText(strconv.FormatUint(uint64(cDevice.JoinNonce), 10))

	for setRedisValuesCancelButton.Clicked(gtx) {
		//Close popup.
		setRedisValues = false
		return false, layout.FlexChild{}
	}

	for setRedisValuesConfirmButton.Clicked(gtx) {
		//Set values.
		var (
			ulFcnt    int
			dlFcnt    int
			devNonce  int
			joinNonce int
		)
		extractInt(&ulFcntEdit, &ulFcnt, 0)
		extractInt(&dlFcntEdit, &dlFcnt, 0)
		extractInt(&devNonceEdit, &devNonce, 0)
		extractInt(&joinNonceEdit, &joinNonce, 0)
		log.Warningln("Setting Redis values")
		err := cDevice.SetValues(ulFcnt, dlFcnt, devNonce, joinNonce)
		if err != nil {
			log.Errorln(err)
		}
		//Close popup.
		setRedisValues = false
		return false, layout.FlexChild{}
	}

	widgets := []layout.FlexChild{
		xmat.RigidSection(gtx, th, "Set counters and nonces"),
		xmat.RigidLabel(gtx, th, "Warning: this will only work when device is activated; when not, values will be reset on program start. Modifying these values may result in failure of communication."),
		xmat.RigidEditor(gtx, th, fmt.Sprintf("DlFcnt"), "<downlink>", &dlFcntEdit),
		xmat.RigidEditor(gtx, th, fmt.Sprintf("UlFcnt"), "<uplink>", &ulFcntEdit),
		xmat.RigidEditor(gtx, th, fmt.Sprintf("DevNonce"), "<dev nonce>", &devNonceEdit),
		xmat.RigidEditor(gtx, th, fmt.Sprintf("JoinNonce"), "<join nonce>", &joinNonceEdit),
		xmat.RigidButton(gtx, th, "Cancel", &setRedisValuesCancelButton),
		xmat.RigidButton(gtx, th, "Confirm", &setRedisValuesConfirmButton),
	}

	inset := layout.Inset{Top: unit.Px(20)}
	return true, layout.Rigid(func() {
		inset.Layout(gtx, func() {
			layout.Flex{Axis: layout.Vertical}.Layout(gtx, widgets...)
		})
	})
}

func join() {

	if !cNSClient.IsConnected() {
		if mqttClient == nil || !mqttClient.IsConnected() {
			log.Errorln("Neither client is connected")
			return
		}
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

	if !cNSClient.IsConnected() {
		err = cDevice.Join(mqttClient, config.MQTT.UplinkTopic, config.GW.MAC, &urx, &utx)
	} else {
		err = cDevice.JoinUDP(cNSClient, config.GW.MAC, &urx, &utx)
	}

	if err != nil {
		log.Errorf("join error: %s", err)
	} else {
		log.Println("join sent")
	}
}

func run() {

	if !cNSClient.IsConnected() {
		if mqttClient == nil || !mqttClient.IsConnected() {
			log.Errorln("Neither client is connected")
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
			if macCommands[i].Use.Value {
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
		mqtt := mqttClient != nil && mqttClient.IsConnected()
		dlMessage, err := cDevice.ProcessDownlink(payload, cDevice.MACVersion, mqtt)
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
