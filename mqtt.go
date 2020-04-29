package main

import (
    "fmt"
    "time"

    paho "github.com/eclipse/paho.mqtt.golang"
    log "github.com/sirupsen/logrus"

    "gioui.org/layout"
    "gioui.org/widget"
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

var (
    mqttServerEdit widget.Editor
    mqttUserEdit widget.Editor
    mqttPasswordEdit widget.Editor
    mqttMACEdit widget.Editor
    mqttDownlinkEdit widget.Editor
    mqttUplinkEdit widget.Editor
    mqttConnectButton widget.Button
    mqttDisconnectButton widget.Button
)

func mqttResetGuiValue() {
    mqttServerEdit.SetText(config.MQTT.Server)
    mqttUserEdit.SetText(config.MQTT.User)
    mqttPasswordEdit.SetText(config.MQTT.Password)
    mqttMACEdit.SetText(config.GW.MAC)
    mqttDownlinkEdit.SetText(config.MQTT.DownlinkTopic)
    mqttUplinkEdit.SetText(config.MQTT.UplinkTopic)
}

func mqttForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {

    config.MQTT.Server = mqttServerEdit.Text()
    config.MQTT.User = mqttUserEdit.Text()
    config.MQTT.Password = mqttPasswordEdit.Text()
    config.GW.MAC = mqttMACEdit.Text()
    config.MQTT.DownlinkTopic = mqttDownlinkEdit.Text()
    config.MQTT.UplinkTopic = mqttUplinkEdit.Text()
    
    for mqttConnectButton.Clicked(gtx) {
        connectClient()
    }

    for mqttDisconnectButton.Clicked(gtx) {
        mqttClient.Disconnect(200)
    }
    
    widgets := []layout.FlexChild {
        layout.Rigid(gioSection(gtx, th, "MQTT & Gateway")),
        layout.Rigid(gioEditor(gtx, th, "MQTT Server", "192.168.1.1", &mqttServerEdit)),
        layout.Rigid(gioEditor(gtx, th, "MQTT User", "", &mqttUserEdit)),
        layout.Rigid(gioEditor(gtx, th, "MQTT Password", "", &mqttPasswordEdit)),
        layout.Rigid(gioEditor(gtx, th, "Gateway MAC", "DEADBEEFDEADBEEF", &mqttMACEdit)),
        layout.Rigid(gioEditor(gtx, th, "Downlink Topic", "gateway/%s/command/down", &mqttDownlinkEdit)),
        layout.Rigid(gioEditor(gtx, th, "Uplink Topic", "gateway/%s/event/up", &mqttUplinkEdit)) }

    if !cNSClient.IsConnected() {
        widgets = append(widgets, layout.Rigid(gioButton(gtx, th, "Connect", &mqttConnectButton)))
    } else {
        widgets = append(widgets, layout.Rigid(gioLabel(gtx, th, "MQTT Connected")))
    }
    
    if mqttClient != nil && mqttClient.IsConnected() {
        widgets = append(widgets, layout.Rigid(gioButton(gtx, th, "Disconnect", &mqttDisconnectButton)))
    }
    
    return layout.Rigid(func() { 
        layout.Flex{Axis: layout.Vertical}.Layout(gtx, widgets...)
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
