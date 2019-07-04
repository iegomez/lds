package main

import (
	"github.com/iegomez/lsp"
	"github.com/inkyblackness/imgui-go"
	log "github.com/sirupsen/logrus"
)

type provisioner struct {
	Hostname string `toml:"hostname"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Path     string `toml:"path"`
	Token    string
	Devices  []*lsp.Device
}

var openProvisioner bool

func beginProvisioner() {
	if openProvisioner {
		imgui.OpenPopup("Provisioner")
		openProvisioner = false
	}
	imgui.SetNextWindowPos(imgui.Vec2{X: float32(config.Window.Width-190) / 2, Y: 120.0})
	imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 180})
	imgui.PushItemWidth(250.0)

	if imgui.BeginPopupModal("Provisioner") {
		imgui.InputText("Hostname##provisioner-hostname", &config.Provisioner.Hostname)
		imgui.InputText("Username##provisioner-username", &config.Provisioner.Username)
		imgui.InputTextV("Password##provisioner-password", &config.Provisioner.Password, imgui.InputTextFlagsPassword, nil)
		imgui.InputText("Path##provisioner-path", &config.Provisioner.Path)

		imgui.Separator()
		if imgui.Button("Login") {
			token, err := lsp.Login(config.Provisioner.Username, config.Provisioner.Password, config.Provisioner.Hostname)
			if err != nil {
				log.Errorf("provisioner login error: %s", err)
			} else {
				config.Provisioner.Token = token
				log.Infoln("login successful")
			}
		}
		imgui.SameLine()
		if imgui.Button("Load") {
			devices, err := lsp.Load(config.Provisioner.Path)
			if err != nil {
				log.Errorf("provisioner load error: %s", err)
			} else {
				config.Provisioner.Devices = devices
				log.Infoln("devices successfully loaded")
			}
		}
		imgui.SameLine()
		if imgui.Button("Provision") {
			lsp.Provision(config.Provisioner.Devices, config.Provisioner.Hostname, config.Provisioner.Token)
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			imgui.CloseCurrentPopup()
		}

		imgui.EndPopup()
	}
}
