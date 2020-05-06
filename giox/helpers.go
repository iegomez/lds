package giox

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// RigidEditor returns layout function for labeled edit field
func RigidEditor(gtx *layout.Context, th *material.Theme, caption string, hint string, editor *widget.Editor) layout.FlexChild {
	return layout.Rigid(func() {
		inset := layout.UniformInset(unit.Dp(3))
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func() {
				inset.Layout(gtx, func() {
					th.Label(unit.Px(16), caption).Layout(gtx)
				})
			}),
			layout.Rigid(func() {
				inset.Layout(gtx, func() {
					th.Editor(hint).Layout(gtx, editor)
				})
			}))
	})
}

// RigidButton returns layout function for a button with inset
func RigidButton(gtx *layout.Context, th *material.Theme, caption string, button *widget.Button) layout.FlexChild {
	inset := layout.UniformInset(unit.Dp(3))
	return layout.Rigid(func() {
		inset.Layout(gtx, func() {
			th.Button("Connect").Layout(gtx, button)
		})
	})
}

// RigidSection returns layout function for a form heading
func RigidSection(gtx *layout.Context, th *material.Theme, caption string) layout.FlexChild {
	inset := layout.UniformInset(unit.Dp(3))
	return layout.Rigid(func() {
		inset.Layout(gtx, func() {
			th.H5(caption).Layout(gtx)
		})
	})
}

// RigidLabel returns layout function for a regular label
func RigidLabel(gtx *layout.Context, th *material.Theme, caption string) layout.FlexChild {
	return layout.Rigid(func() {
		th.Label(unit.Px(16), caption).Layout(gtx)
	})
}
