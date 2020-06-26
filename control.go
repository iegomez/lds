package main

import (
	"strconv"

	"github.com/brocaar/lorawan"
	xmat "github.com/scartill/giox/material"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type settingWidget struct {
	Editor   widget.Editor
	CheckBox widget.Bool
}

type commandSetting struct {
	IsBool bool
	Widget settingWidget
	Label  string
	Hint   string
	Setter func(*settingWidget, lorawan.MACCommandPayload)
	Getter func(*settingWidget, lorawan.MACCommandPayload)
}

type macCommandItem struct {
	MACCommand lorawan.MACCommand
	Use        widget.Bool
	Settings   []commandSetting
}

var fCtrl lorawan.FCtrl

type fCtrlWidgetsType struct {
	ACK       widget.Bool
	ADR       widget.Bool
	ADRACKReq widget.Bool
	ClassB    widget.Bool
	fPending  widget.Bool
}

var fCtrlWidgets fCtrlWidgetsType

//List of all available mac commands and their payloads.
var macCommands = []*macCommandItem{
	{
		MACCommand: lorawan.MACCommand{
			CID: lorawan.ResetInd,
			Payload: &lorawan.ResetIndPayload{
				DevLoRaWANVersion: lorawan.Version{
					Minor: 0,
				},
			},
		},
		Settings: []commandSetting{
			{
				Label: "DevLoRaWANVersion",
				Hint:  "<DevLWVersion-Minor>",
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.ResetIndPayload)
					widget.Editor.SetText(strconv.Itoa(int(casted.DevLoRaWANVersion.Minor)))
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.ResetIndPayload)
					extractUInt8(&widget.Editor, &casted.DevLoRaWANVersion.Minor, 0)
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.LinkCheckReq,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.LinkADRAns,
			Payload: &lorawan.LinkADRAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label:  "ChannelMaskACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					widget.CheckBox.Value = casted.ChannelMaskACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					casted.ChannelMaskACK = widget.CheckBox.Value
				},
			},
			{
				Label:  "DateRateACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					widget.CheckBox.Value = casted.DataRateACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					casted.DataRateACK = widget.CheckBox.Value
				},
			},
			{
				Label:  "PowerACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					widget.CheckBox.Value = casted.PowerACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.LinkADRAnsPayload)
					casted.PowerACK = widget.CheckBox.Value
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DutyCycleAns,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RXParamSetupAns,
			Payload: &lorawan.RXParamSetupAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label:  "ChannelACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					widget.CheckBox.Value = casted.ChannelACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					casted.ChannelACK = widget.CheckBox.Value
				},
			},
			{
				Label:  "RX2DateRateACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					widget.CheckBox.Value = casted.RX2DataRateACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					casted.RX2DataRateACK = widget.CheckBox.Value
				},
			},
			{
				Label:  "RX1DROffsetACK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					widget.CheckBox.Value = casted.RX1DROffsetACK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RXParamSetupAnsPayload)
					casted.RX1DROffsetACK = widget.CheckBox.Value
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DevStatusAns,
			Payload: &lorawan.DevStatusAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label: "Battery",
				Hint:  "<battery>",
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DevStatusAnsPayload)
					widget.Editor.SetText(strconv.Itoa(int(casted.Battery)))
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DevStatusAnsPayload)
					extractUInt8(&widget.Editor, &casted.Battery, 0)
				},
			},
			{
				Label: "Margin",
				Hint:  "<margin>",
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DevStatusAnsPayload)
					widget.Editor.SetText(strconv.Itoa(int(casted.Margin)))
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DevStatusAnsPayload)
					extractInt8(&widget.Editor, &casted.Margin, 0)
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.NewChannelAns,
			Payload: &lorawan.NewChannelAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label:  "ChannelFrequencyOK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.NewChannelAnsPayload)
					widget.CheckBox.Value = casted.ChannelFrequencyOK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.NewChannelAnsPayload)
					casted.ChannelFrequencyOK = widget.CheckBox.Value
				},
			},
			{
				Label:  "DataRateRangeOK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.NewChannelAnsPayload)
					widget.CheckBox.Value = casted.DataRateRangeOK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.NewChannelAnsPayload)
					casted.DataRateRangeOK = widget.CheckBox.Value
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RXTimingSetupAns,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.TXParamSetupAns,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DLChannelAns,
			Payload: &lorawan.DLChannelAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label:  "ChannelFrequencyOK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DLChannelAnsPayload)
					widget.CheckBox.Value = casted.ChannelFrequencyOK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DLChannelAnsPayload)
					casted.ChannelFrequencyOK = widget.CheckBox.Value
				},
			},
			{
				Label:  "UplinkFrequencyExists",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DLChannelAnsPayload)
					widget.CheckBox.Value = casted.UplinkFrequencyExists
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.DLChannelAnsPayload)
					casted.UplinkFrequencyExists = widget.CheckBox.Value
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RekeyInd,
			Payload: &lorawan.RekeyIndPayload{},
		},
		Settings: []commandSetting{
			{
				Label: "DevLoRaWANVersion",
				Hint:  "<DevLoRaWANVersion>",
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RekeyIndPayload)
					widget.Editor.SetText(strconv.Itoa(int(casted.DevLoRaWANVersion.Minor)))
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RekeyIndPayload)
					extractUInt8(&widget.Editor, &casted.DevLoRaWANVersion.Minor, 0)
				},
			},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.ADRParamSetupAns,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DeviceTimeReq,
			Payload: nil,
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RejoinParamSetupAns,
			Payload: &lorawan.RejoinParamSetupAnsPayload{},
		},
		Settings: []commandSetting{
			{
				Label:  "TimeOK",
				IsBool: true,
				Setter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RejoinParamSetupAnsPayload)
					widget.CheckBox.Value = casted.TimeOK
				},
				Getter: func(widget *settingWidget, payload lorawan.MACCommandPayload) {
					casted := payload.(*lorawan.RejoinParamSetupAnsPayload)
					casted.TimeOK = widget.CheckBox.Value
				},
			},
		},
	},
}

func macResetGuiValues() {
	fCtrlWidgets.ACK.Value = fCtrl.ACK
	fCtrlWidgets.ADR.Value = fCtrl.ADR
	fCtrlWidgets.ClassB.Value = fCtrl.ClassB
	fCtrlWidgets.fPending.Value = fCtrl.FPending

	for _, command := range macCommands {
		if command.MACCommand.Payload == nil {
			continue
		}

		for _, setting := range command.Settings {
			setting.Setter(&setting.Widget, command.MACCommand.Payload)
		}
	}
}

func controlForm(th *material.Theme) l.FlexChild {
	fCtrl.ACK = fCtrlWidgets.ACK.Value
	fCtrl.ADR = fCtrlWidgets.ADR.Value
	fCtrl.ClassB = fCtrlWidgets.ClassB.Value
	fCtrl.FPending = fCtrlWidgets.fPending.Value

	for _, command := range macCommands {
		if command.MACCommand.Payload != nil {
			for _, setting := range command.Settings {
				setting.Getter(&setting.Widget, command.MACCommand.Payload)
			}
		}
	}

	widgets := []l.FlexChild{
		xmat.RigidSection(th, "Control"),
		xmat.RigidLabel(th, "FCtrl"),
		l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Horizontal}.Layout(gtx,
				xmat.RigidCheckBox(th, "ACK", &fCtrlWidgets.ACK),
				xmat.RigidCheckBox(th, "ARD", &fCtrlWidgets.ADR),
				xmat.RigidCheckBox(th, "ClassB", &fCtrlWidgets.ClassB),
				xmat.RigidCheckBox(th, "FPending", &fCtrlWidgets.fPending),
			)
		}),
		xmat.RigidLabel(th, "MAC Commands"),
	}

	for c := 0; c < len(macCommands); c++ {
		command := macCommands[c]

		label := command.MACCommand.CID.String()
		checkbox := &command.Use

		commandWidgets := []l.FlexChild{
			xmat.RigidCheckBox(th, label, checkbox),
		}

		if command.MACCommand.Payload != nil {
			subwidgets := make([]l.FlexChild, 0)
			for s := 0; s < len(command.Settings); s++ {
				setting := &macCommands[c].Settings[s]

				var widget l.FlexChild
				if setting.IsBool {
					widget = xmat.RigidCheckBox(th, setting.Label, &setting.Widget.CheckBox)
				} else {
					widget = xmat.RigidEditor(th, setting.Label, setting.Hint, &setting.Widget.Editor)
				}
				subwidgets = append(subwidgets, widget)
			}

			subinset := l.Inset{Left: unit.Px(20)}
			settingsWidget := l.Rigid(func(gtx l.Context) l.Dimensions {
				return subinset.Layout(gtx, func(gtx l.Context) l.Dimensions {
					return l.Flex{Axis: l.Horizontal}.Layout(gtx,
						subwidgets...,
					)
				})
			})
			commandWidgets = append(commandWidgets, settingsWidget)
		}

		subsection := l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Horizontal}.Layout(gtx, commandWidgets...)
		})
		widgets = append(widgets, subsection)
	}

	inset := l.Inset{Top: unit.Px(20)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
	})
}
