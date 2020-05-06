package material

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/iegomez/lds/giox"
)

// ComboStyle holds combobox rendering parameters
type ComboStyle struct {
	theme *material.Theme
}

// Combo constructs c ComboStyle
func Combo(th *material.Theme) ComboStyle {
	return ComboStyle{theme: th}
}

// Layout a combobox
func (c ComboStyle) Layout(gtx *layout.Context, widget *giox.Combo) {
	c.theme.Label(unit.Px(16), widget.Selected()).Layout(gtx)
}
