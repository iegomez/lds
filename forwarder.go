package main

import (
    "strconv"

    log "github.com/sirupsen/logrus"

    "github.com/iegomez/lds/lds"

    "gioui.org/layout"
    "gioui.org/widget"
    "gioui.org/widget/material"
)

// cNSClient is a direct NetworkServer connection handle
var cNSClient lds.NSClient

type forwarder struct {
    Server string `toml:"nserver"`
    Port   string `toml:"nsport"`
}

var (
    nserverEdit widget.Editor
    nportEdit widget.Editor
    nsConnectButton widget.Button
)

func forwarderResetGuiValues() {
	nserverEdit.SetText(config.Forwarder.Server)
	nportEdit.SetText(config.Forwarder.Port)
}

func forwarderForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {

    config.Forwarder.Server = nserverEdit.Text()
    config.Forwarder.Port = nportEdit.Text()

    for nsConnectButton.Clicked(gtx) {
        forwarderConnect()
    }

    wLabel := layout.Rigid(gioSection(gtx, th, "Forwarder"))
    wNS := layout.Rigid(gioEditor(gtx, th, "Network Server", "192.168.1.1", &nserverEdit))
    wNP := layout.Rigid(gioEditor(gtx, th, "UDP Port", "1680", &nportEdit))

    var wConnect layout.FlexChild
    if mqttClient == nil || !mqttClient.IsConnected() {
        if !cNSClient.IsConnected() {
            wConnect = layout.Rigid(gioButton(gtx, th, "Connect", &nsConnectButton))
        } else {
            wConnect = layout.Rigid(gioLabel(gtx, th, "UDP Listening"))
        }
    }
    
    return layout.Rigid(func() { 
        layout.Flex{Axis: layout.Vertical}.Layout(gtx, wLabel, wNS, wNP, wConnect)
    })
}

func forwarderConnect() error {
    port, err := strconv.Atoi(config.Forwarder.Port)

    if err != nil {
        log.Warn("network server UDP port must be a number")
        return err
    }

    cNSClient.Server = config.Forwarder.Server
    cNSClient.Port = port
    cNSClient.Connect(config.GW.MAC, onIncomingDownlink)
    log.Infoln("UDP Forwarder started (MQTT disabled)")

    return nil
}
