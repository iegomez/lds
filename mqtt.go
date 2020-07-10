package main

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	matx "github.com/scartill/giox/material"
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

var (
	mqttServerEdit       widget.Editor
	mqttUserEdit         widget.Editor
	mqttPasswordEdit     widget.Editor
	mqttMACEdit          widget.Editor
	mqttDownlinkEdit     widget.Editor
	mqttUplinkEdit       widget.Editor
	mqttConnectButton    widget.Clickable
	mqttDisconnectButton widget.Clickable
)

func mqttResetGuiValue() {
	mqttServerEdit.SetText(config.MQTT.Server)
	mqttUserEdit.SetText(config.MQTT.User)
	mqttPasswordEdit.SetText(config.MQTT.Password)
	mqttMACEdit.SetText(config.GW.MAC)
	mqttDownlinkEdit.SetText(config.MQTT.DownlinkTopic)
	mqttUplinkEdit.SetText(config.MQTT.UplinkTopic)
}

func mqttForm(th *material.Theme) l.FlexChild {

	config.MQTT.Server = mqttServerEdit.Text()
	config.MQTT.User = mqttUserEdit.Text()
	config.MQTT.Password = mqttPasswordEdit.Text()
	config.GW.MAC = mqttMACEdit.Text()
	config.MQTT.DownlinkTopic = mqttDownlinkEdit.Text()
	config.MQTT.UplinkTopic = mqttUplinkEdit.Text()

	for mqttConnectButton.Clicked() {
		connectClient()
	}

	for mqttDisconnectButton.Clicked() {
		mqttClient.Disconnect(200)
	}

	widgets := []l.FlexChild{
		matx.RigidSection(th, "MQTT & Gateway"),
		matx.RigidEditor(th, "MQTT Server:", "192.168.1.1", &mqttServerEdit),
		matx.RigidEditor(th, "MQTT User:", "<username>", &mqttUserEdit),
		matx.RigidEditor(th, "MQTT Password:", "<password>", &mqttPasswordEdit),
		matx.RigidEditor(th, "Gateway MAC:", "DEADBEEFDEADBEEF", &mqttMACEdit),
		matx.RigidEditor(th, "Downlink Topic:", "gateway/%s/command/down", &mqttDownlinkEdit),
		matx.RigidEditor(th, "Uplink Topic:", "gateway/%s/event/up", &mqttUplinkEdit)}

	if !cNSClient.IsConnected() {
		widgets = append(widgets, matx.RigidButton(th, "Connect", &mqttConnectButton))
	} else {
		widgets = append(widgets, matx.RigidLabel(th, "Forwarder connected"))
	}

	if mqttClient != nil && mqttClient.IsConnected() {
		widgets = append(widgets, matx.RigidButton(th, "Disconnect", &mqttDisconnectButton))
	}

	inset := l.Inset{Left: unit.Dp(30)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
	})
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
