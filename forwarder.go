package main

import (
	"strconv"

	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
)

type nsclient struct {
	Connected bool
	Server string
	Port int
}

// NSClient is a direct NetworkServer connection handle
var NSClient nsclient

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
	if NSClient.Connected {
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
		log.Warn("Bad Network server UDP port")
		return err
	}

	NSClient.Server = config.Forwarder.Server
	NSClient.Port = port
	NSClient.Connected = true
	log.Infoln("UDP Forwarder Started (MQTT disabled)")
	// TODO subscribe to downlinks

	return nil
}

func forwarderDisconnect() error {
	NSClient.Connected = false
	log.Infoln("UDP Forwarder Stopped (MQTT back again")

	return nil
}

