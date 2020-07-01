package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	xmat "github.com/scartill/giox/material"
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

var logEditor widget.Editor

func outputForm(th *material.Theme) l.FlexChild {
	logEditor.SetText(ow.Text)
	editorStyle := material.Editor(th, &logEditor, "")

	widgets := []l.FlexChild{
		xmat.RigidSection(th, "Output"),
		l.Rigid(func(gtx l.Context) l.Dimensions {
			return editorStyle.Layout(gtx)
		}),
	}

	inset := l.Inset{Left: unit.Px(30)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
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
