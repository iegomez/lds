package main

import (
	"fmt"
	"os"
	"time"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	xmat "github.com/scartill/giox/material"
	log "github.com/sirupsen/logrus"
)

func writeHistory() {
	f, err := os.Create(fmt.Sprintf("lds-%d.log", time.Now().UnixNano()))
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	defer f.Close()
	n, err := f.Write([]byte(ow.Text()))
	f.Sync()
	log.Infof("wrote %d bytes to %s", n, f.Name())
}

func setLevel(level log.Level) {
	log.SetLevel(level)
}

var logList l.List

func createOutputForm() {
	logList = l.List{Axis: l.Vertical, ScrollToEnd: true}
}

func outputForm(th *material.Theme) l.FlexChild {

	widgets := []l.FlexChild{
		xmat.RigidSection(th, "Output"),
		l.Rigid(func(gtx l.Context) l.Dimensions {
			return logList.Layout(gtx, len(ow.Lines), func(gtx l.Context, i int) l.Dimensions {
				return material.Body1(th, ow.Lines[i]).Layout(gtx)
			})
		}),
	}

	inset := l.Inset{Left: unit.Dp(30)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
	})
}
