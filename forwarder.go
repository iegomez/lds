package main

import (
	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
)

type nsclient struct {
	Connected bool
}

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
	NSClient.Connected = true
	log.Infoln("UDP Forwared Started")
	// TODO subscribe to downlinks

	return nil
}

func forwarderDisconnect() error {
	log.Infoln("UDP Forwared Stopped")

	return nil
}

