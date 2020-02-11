package main

import (
	"strconv"

	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
	"github.com/iegomez/lds/lds"
)

// cNSClient is a direct NetworkServer connection handle
var cNSClient lds.NSClient

type forwarder struct {
	Server		string `toml:"nserver"`
	Port		string `toml:"nsport"`
}

func beginForwarderForm() {
	imgui.Begin("Forwarder")
	imgui.Separator()
	imgui.PushItemWidth(250.0)
	imgui.InputText("Network Server", &config.Forwarder.Server)
	imgui.InputText("UDP Port", &config.Forwarder.Port)

	if imgui.Button("Connect") {
		forwarderConnect()
	}
	if cNSClient.Connected {
		if imgui.Button("Disconnect") {
			forwarderDisconnect()
		}
	}
	//Add popus for file administration.
	beginOpenFile()
	beginSaveFile()
	imgui.End()
}

func forwarderConnect() error {
	port, err := strconv.Atoi(config.Forwarder.Port)

	if err != nil {
		log.Warn("network server UDP port must be a number")
		return err
	}

	cNSClient.Server = config.Forwarder.Server
	cNSClient.Port = port
	cNSClient.Connected = true
	log.Infoln("UDP Forwarder started (MQTT disabled)")
	// TODO subscribe to downlinks

	return nil
}

func forwarderDisconnect() error {
	cNSClient.Connected = false
	log.Infoln("UDP Forwarder STopped (MQTT back again)")

	return nil
}

