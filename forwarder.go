package main

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/iegomez/lds/lds"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	matx "github.com/scartill/giox/material"
)

// cNSClient is a direct NetworkServer connection handle
var cNSClient lds.NSClient

type forwarder struct {
	Server string `toml:"nserver"`
	Port   string `toml:"nsport"`
}

var (
	nserverEdit     widget.Editor
	nportEdit       widget.Editor
	nsConnectButton widget.Clickable
)

func forwarderResetGuiValues() {
	nserverEdit.SetText(config.Forwarder.Server)
	nportEdit.SetText(config.Forwarder.Port)
}

func forwarderForm(th *material.Theme) l.FlexChild {

	config.Forwarder.Server = nserverEdit.Text()
	config.Forwarder.Port = nportEdit.Text()

	for nsConnectButton.Clicked() {
		forwarderConnect()
	}

	wLabel := matx.RigidSection(th, "Forwarder")
	wNS := matx.RigidEditor(th, "Network Server:", "192.168.1.1", &nserverEdit)
	wNP := matx.RigidEditor(th, "UDP Port:", "1680", &nportEdit)

	var wConnect l.FlexChild
	if mqttClient == nil || !mqttClient.IsConnected() {
		if !cNSClient.IsConnected() {
			wConnect = matx.RigidButton(th, "Connect", &nsConnectButton)
		} else {
			wConnect = matx.RigidLabel(th, "UDP Listening")
		}
	}

	inset := l.Inset{Top: unit.Px(20)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, wLabel, wNS, wNP, wConnect)
		})
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
