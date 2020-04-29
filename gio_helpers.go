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

func gioButton(gtx *layout.Context, th *material.Theme, caption string, button *widget.Button) func() {
    return func() {
        th.Button("Connect").Layout(gtx, button)
    }
}

func gioSection(gtx *layout.Context, th * material.Theme, caption string) func() {
    return func() {
        th.H5(caption).Layout(gtx)
    }
}

func gioLabel(gtx *layout.Context, th * material.Theme, caption string) func() {
    return func() {
        th.Label(unit.Px(16), caption).Layout(gtx)
    }
}
