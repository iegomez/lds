package main

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/iegomez/lds/lds"
	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
)

var mqttClient paho.Client

type mqtt struct {
	Server        string `toml:"server"`
	User          string `toml:"user"`
	Password      string `toml:"password"`
	DownlinkTopic string `toml:"downlink_topic"`
	UplinkTopic   string `toml:"uplink_topic"`
}

type gateway struct {
	MAC           string `toml:"mac"`
	BridgeVersion string `toml:"bridge_version"`
}

func beginMQTTForm() {
	//imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 25})
	//imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 170})
	imgui.Begin("MQTT & Gateway")
	imgui.Separator()
	imgui.PushItemWidth(250.0)
	imgui.InputText("Server", &config.MQTT.Server)
	imgui.InputText("User", &config.MQTT.User)
	imgui.InputTextV("Password", &config.MQTT.Password, imgui.InputTextFlagsPassword, nil)
	imgui.InputText("MAC", &config.GW.MAC)
	imgui.InputText("Downlink topic", &config.MQTT.DownlinkTopic)
	imgui.InputText("Uplink topic", &config.MQTT.UplinkTopic)
	if imgui.Button("Connect") {
		connectClient()
	}
	if mqttClient != nil && mqttClient.IsConnected() {
		if imgui.Button("Disconnect") {
			mqttClient.Disconnect(200)
			log.Infoln("mqtt client disconnected")
		}
	}
	//Add popus for file administration.
	beginOpenFile()
	beginSaveFile()
	imgui.End()
}

func connectClient() error {
	//Connect to the broker
	opts := paho.NewClientOptions()
	opts.AddBroker(config.MQTT.Server)
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Password)
	opts.SetAutoReconnect(true)
	opts.SetClientID(fmt.Sprintf("lds-%d", time.Now().UnixNano()))

	mqttClient = paho.NewClient(opts)
	log.Infoln("connecting...")
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("connection error: %s", token.Error())
		return token.Error()
	}
	log.Infoln("connection established")
	mqttClient.Subscribe(fmt.Sprintf(config.MQTT.DownlinkTopic, config.GW.MAC), 1, func(c paho.Client, msg paho.Message) {
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
			//Get redis info.
			cDevice.GetInfo()
		}
	})
	return nil
}
