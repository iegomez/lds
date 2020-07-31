package main

import (
	l "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/iegomez/lsp"
	xmat "github.com/scartill/giox/material"
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

// Widgets
var (
	provHostnameEdit widget.Editor
	provUsernameEdit widget.Editor
	provPasswordEdit widget.Editor
	provPathEdit     widget.Editor

	provLoginBtn  widget.Clickable
	provLoadBtn   widget.Clickable
	provProvBtn   widget.Clickable
	provCancelBtn widget.Clickable
)

func provResetGuiValues() {
	provHostnameEdit.SetText(config.Provisioner.Hostname)
	provUsernameEdit.SetText(config.Provisioner.Username)
	provPasswordEdit.SetText(config.Provisioner.Password)
	provPathEdit.SetText(config.Provisioner.Path)
}

func buildProvisioner(th *material.Theme) (l.FlexChild, bool) {
	config.Provisioner.Hostname = provHostnameEdit.Text()
	config.Provisioner.Username = provUsernameEdit.Text()
	config.Provisioner.Password = provPasswordEdit.Text()
	config.Provisioner.Path = provPathEdit.Text()

	for provLoginBtn.Clicked() {
		token, err := lsp.Login(config.Provisioner.Username, config.Provisioner.Password, config.Provisioner.Hostname)
		if err != nil {
			log.Errorf("provisioner login error: %s", err)
		} else {
			config.Provisioner.Token = token
			log.Infoln("login successful")
		}
	}

	for provLoadBtn.Clicked() {
		devices, err := lsp.Load(config.Provisioner.Path)
		if err != nil {
			log.Errorf("provisioner load error: %s", err)
		} else {
			config.Provisioner.Devices = devices
			log.Infoln("devices successfully loaded")
		}
	}

	for provProvBtn.Clicked() {
		lsp.Provision(config.Provisioner.Devices, config.Provisioner.Hostname, config.Provisioner.Token)
	}

	for provCancelBtn.Clicked() {
		openProvisioner = false
	}

	widgets := l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			xmat.RigidSection(th, "Provisioner"),
			xmat.RigidEditor(th, "Hostname", "<hostname>", &provHostnameEdit),
			xmat.RigidEditor(th, "Username", "<username>", &provUsernameEdit),
			xmat.RigidEditor(th, "Password", "<password>", &provPasswordEdit),
			xmat.RigidEditor(th, "Path", "<path>", &provPathEdit),
			l.Rigid(func(gtx l.Context) l.Dimensions {
				return l.Flex{Axis: l.Horizontal}.Layout(gtx,
					xmat.RigidButton(th, "Login", &provLoginBtn),
					xmat.RigidButton(th, "Load", &provLoadBtn),
					xmat.RigidButton(th, "Provision", &provProvBtn),
					xmat.RigidButton(th, "Cancel", &provCancelBtn),
				)
			}),
		)
	})

	return widgets, true
}
