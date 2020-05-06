package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

    "gioui.org/layout"
    "gioui.org/widget/material"
)

func writeHistory() {
	f, err := os.Create(fmt.Sprintf("lds-%d.log", time.Now().UnixNano()))
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	defer f.Close()
	n, err := f.Write([]byte(ow.History))
	f.Sync()
	log.Infof("wrote %d bytes to %s", n, f.Name())
}

func setLevel(level log.Level) {
	log.SetLevel(level)
}

func outputForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
	return layout.Rigid( func() {
		material.Caption(th, "output").Layout(gtx)
	})
}

func beginOutput() {
/*! //imgui.SetNextWindowPos(imgui.Vec2{X: 400, Y: 650})
	//imgui.SetNextWindowSize(imgui.Vec2{X: 780, Y: 265})
	imgui.Begin("Output")
	imgui.PushTextWrapPos()
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.1, Y: 0.8, Z: 0.1, W: 0.5})
	imgui.Text(ow.Text)
	imgui.PopStyleColor()
	imgui.End()*/
}
