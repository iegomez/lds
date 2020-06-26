package main

import (
	"strconv"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/scartill/giox"
	xmat "github.com/scartill/giox/material"
)

func labelCombo(th *material.Theme, label string, combo *giox.Combo) l.FlexChild {
	inset := l.Inset{Top: unit.Px(10), Right: unit.Px(10)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Horizontal}.Layout(gtx,
			l.Rigid(func(gtx l.Context) l.Dimensions {
				return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
					return material.Label(th, unit.Px(16), label).Layout(gtx)
				})
			}),
			l.Rigid(func(gtx l.Context) l.Dimensions {
				return xmat.Combo(th, combo).Layout(gtx)
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

func extractInt8(edit *widget.Editor, value *int8, onError int8) {
	parsed, err := strconv.ParseInt(edit.Text(), 10, 8)

	if err == nil {
		*value = int8(parsed)
	} else {
		*value = onError
	}
}

func extractInt32(edit *widget.Editor, value *int32, onError int32) {
	parsed, err := strconv.ParseInt(edit.Text(), 10, 32)

	if err == nil {
		*value = int32(parsed)
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
