package main

import (
    "gioui.org/layout"
    "gioui.org/unit"
    "gioui.org/widget"
    "gioui.org/widget/material"
)

func gioEditor(gtx *layout.Context, th *material.Theme, caption string, hint string, editor *widget.Editor) func() {
    return func() {
        layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
            layout.Rigid( func() {
                th.Label(unit.Px(16), caption).Layout(gtx)
			}),
            layout.Rigid( func() {
                th.Editor(hint).Layout(gtx, editor)
            }))
    }
}