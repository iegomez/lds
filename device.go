package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
	"github.com/golang/protobuf/ptypes"
	"github.com/iegomez/lds/lds"
	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
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
	}
	beginReset()
	imgui.Separator()
	if cDevice != nil {
		imgui.Text(fmt.Sprintf("DlFCnt: %d - UlFCnt: %d", cDevice.DlFcnt, cDevice.UlFcnt))
		imgui.Text(fmt.Sprintf("DevNonce: %d - JoinNonce: %d", cDevice.DevNonce, cDevice.JoinNonce))
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
	joinEUI, err := lds.HexToEUI(config.Device.JoinEUI)
	if err != nil {
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
			Major:         lorawan.Major(config.Device.Major),
			MACVersion:    lorawan.MACVersion(config.Device.MACVersion),
			SkipFCntCheck: config.Device.SkipFCntCheck,
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
		cDevice.Major = lorawan.Major(config.Device.Major)
		cDevice.MACVersion = lorawan.MACVersion(config.Device.MACVersion)
		cDevice.SkipFCntCheck = config.Device.SkipFCntCheck
	}
	cDevice.SetMarshaler(config.Device.Marshaler)
	//Get redis info.
	if cDevice.GetInfo() {
		config.Device.NwkSEncKey = lds.KeyToHex(cDevice.NwkSEncKey)
		config.Device.FNwkSIntKey = lds.KeyToHex(cDevice.FNwkSIntKey)
		config.Device.SNwkSIntKey = lds.KeyToHex(cDevice.SNwkSIntKey)
		config.Device.AppSKey = lds.KeyToHex(cDevice.AppSKey)
		config.Device.DevAddress = lds.DevAddressToHex(cDevice.DevAddr)
	} else {
		cDevice.Reset()
	}
}

func beginReset() {
	if resetDevice {
		imgui.OpenPopup("Reset device")
		resetDevice = false
	}
	imgui.SetNextWindowPos(imgui.Vec2{X: float32(windowWidth-190) / 2, Y: float32(windowHeight-90) / 2})
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
				config.Device.NwkSEncKey = lds.KeyToHex(cDevice.NwkSEncKey)
				config.Device.FNwkSIntKey = lds.KeyToHex(cDevice.FNwkSIntKey)
				config.Device.SNwkSIntKey = lds.KeyToHex(cDevice.SNwkSIntKey)
				config.Device.AppSKey = lds.KeyToHex(cDevice.AppSKey)
				config.Device.DevAddress = lds.DevAddressToHex(cDevice.DevAddr)
				log.Infoln("device was reset")
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
	}

	//Always set device to get any changes to the configuration.
	setDevice()

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

	setDevice()

	dataRate := &lds.DataRate{
		Bandwidth:    config.DR.Bandwith,
		Modulation:   "LORA",
		SpreadFactor: config.DR.SpreadFactor,
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
				return
			}
		} else if config.RawPayload.UseEncoder {
			payload, pErr = EncodeToBytes()
			if pErr != nil {
				log.Errorf("couldn't encode js object: %s", pErr)
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

		var fOpts []*lorawan.MACCommand
		for i := 0; i < len(macCommands); i++ {
			if macCommands[i].Use {
				fOpts = append(fOpts, &macCommands[i].MACCommand)
			}
		}

		//Now send an uplink
		ulfc, err := cDevice.Uplink(mqttClient, config.Device.MType, uint8(config.RawPayload.FPort), &urx, &utx, payload, config.GW.MAC, config.Band.Name, *dataRate, fOpts, fCtrl)
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
