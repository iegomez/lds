package main

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"

    "gioui.org/layout"
    "gioui.org/widget/material"
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

func mqttForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
	return layout.Rigid( func() {
		th.Caption("mqtt").Layout(gtx)
	})
}

func beginMQTTForm() {
/*! //imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 25})
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
	if !cNSClient.IsConnected() {
		if imgui.Button("Connect") {
			connectClient()
		}
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
	imgui.End()*/
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
	log.Infoln("MQTT connecting...")
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("connection error: %s", token.Error())
		return token.Error()
	}
	log.Infoln("connection established")
	mqttClient.Subscribe(fmt.Sprintf(config.MQTT.DownlinkTopic, config.GW.MAC), 1, func(c paho.Client, msg paho.Message) {
		onIncomingDownlink(msg.Payload())
	})
	return nil
}
