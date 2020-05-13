package main

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/scartill/giox"
	xmat "github.com/scartill/giox/material"
)

func labelCombo(gtx *layout.Context, th *material.Theme, label string, combo *giox.Combo) layout.FlexChild {
	inset := layout.Inset{Top: unit.Px(10), Right: unit.Px(10)}
	return layout.Rigid(func() {
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func() {
				inset.Layout(gtx, func() {
					material.Label(th, unit.Px(16), label).Layout(gtx)
				})
			}),
			layout.Rigid(func() {
				xmat.Combo(th).Layout(gtx, combo)
			}))
	})
}

func extractInt(edit *widget.Editor, value *int, onError int) {
	parsed, err := strconv.Atoi(edit.Text())

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}

func extractUInt8(edit *widget.Editor, value *uint8, onError uint8) {
	parsed, err := strconv.ParseUint(edit.Text(), 10, 8)

	if err == nil {
		*value = uint8(parsed)
	} else {
		*value = onError
	}
}

func extractFloat(edit *widget.Editor, value *float64, onError float64) {
	parsed, err := strconv.ParseFloat(edit.Text(), 64)

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}

func extractIntCombo(combo *giox.Combo, value *int, onError int) {
	if !combo.HasSelected() {
		*value = onError
		return
	}

	parsed, err := strconv.Atoi(combo.SelectedText())

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}
