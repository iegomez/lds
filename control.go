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
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DevStatusAns,
			Payload: &lorawan.DevStatusAnsPayload{},
		},
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.NewChannelAns,
			Payload: &lorawan.NewChannelAnsPayload{},
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
	},
	{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RekeyInd,
			Payload: &lorawan.RekeyIndPayload{},
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
	},
}

func macResetGuiValues() {
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
	for _, command := range macCommands {
		if command.MACCommand.Payload == nil {
			continue
		}

		for _, setting := range command.Settings {
			setting.Getter(&setting.Widget, command.MACCommand.Payload)
		}
	}

	widgets := []l.FlexChild{
		xmat.RigidSection(th, "Control"),
	}

	for c := 0; c < len(macCommands); c++ {
		command := macCommands[c]
		if command.MACCommand.Payload == nil {
			continue
		}

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

		label := command.MACCommand.CID.String()
		checkbox := &command.Use
		subsection := l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Horizontal}.Layout(gtx,
				xmat.RigidCheckBox(th, label, checkbox),
				settingsWidget,
			)
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

func beginMACCommands() {
	/*!
	for i := 0; i < len(macCommands); i++ {
		macCommand := macCommands[i]
		imgui.PushItemWidth(200.0)
		imgui.Checkbox(macCommand.MACCommand.CID.String(), &macCommand.Use)
		//Create a mac command form depending on the selected mac command.
		if macCommand.MACCommand.CID != lorawan.CID(0) {
			switch macCommand.MACCommand.CID {
			case lorawan.ResetInd:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.ResetIndPayload)
				imgui.InputTextV("DevLoRaWANVersion", &ResetIndMinorS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(ResetIndMinorS, 1, &payload.DevLoRaWANVersion.Minor))
			case lorawan.LinkADRAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.LinkADRAnsPayload)
				imgui.Checkbox("ChannelMaskACK", &payload.ChannelMaskACK)
				imgui.SameLine()
				imgui.Checkbox("DateRateACK", &payload.DataRateACK)
				imgui.SameLine()
				imgui.Checkbox("PowerACK", &payload.PowerACK)
			case lorawan.RXParamSetupAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RXParamSetupAnsPayload)
				imgui.Checkbox("ChannelACK", &payload.ChannelACK)
				imgui.SameLine()
				imgui.Checkbox("RX2DateRateACK", &payload.RX2DataRateACK)
				imgui.SameLine()
				imgui.Checkbox("RX1DROffsetACK", &payload.RX1DROffsetACK)
			case lorawan.DevStatusAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.DevStatusAnsPayload)
				imgui.InputTextV("Battery", &BatteryS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(BatteryS, 2, &payload.Battery))
				imgui.SameLine()
				imgui.InputTextV("Margin", &MarginS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt8(MarginS, 2, &payload.Margin))
			case lorawan.NewChannelAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.NewChannelAnsPayload)
				imgui.Checkbox("ChannelFrequencyOK", &payload.ChannelFrequencyOK)
				imgui.SameLine()
				imgui.Checkbox("DataRateRangeOK", &payload.DataRateRangeOK)
			case lorawan.DLChannelAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.DLChannelAnsPayload)
				imgui.Checkbox("ChannelFrequencyOK", &payload.ChannelFrequencyOK)
				imgui.SameLine()
				imgui.Checkbox("UplinkFrequencyExists", &payload.UplinkFrequencyExists)
			case lorawan.RekeyInd:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RekeyIndPayload)
				imgui.InputTextV("DevLoRaWANVersion", &RekeyIndMinorS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(RekeyIndMinorS, 1, &payload.DevLoRaWANVersion.Minor))
			case lorawan.RejoinParamSetupAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RejoinParamSetupAnsPayload)
				imgui.Checkbox("TimeOK", &payload.TimeOK)
			default:
				//Nothing to do for nil payload commands.
			}
		}
	}*/
}

func beginFCtrl() {
	/*!	imgui.Checkbox("ACK##FCtrl-ACK", &fCtrl.ACK)
	imgui.SameLine()
	imgui.Checkbox("ADR##FCtrl-ADR", &fCtrl.ADR)
	imgui.SameLine()
	imgui.Checkbox("ADRACKReq##FCtrl-ADRACKReq", &fCtrl.ADRACKReq)
	imgui.SameLine()
	imgui.Checkbox("ClassB##FCtrl-ClassB", &fCtrl.ClassB)
	imgui.SameLine()
	imgui.Checkbox("FPending##FCtrl-FPending", &fCtrl.FPending)*/
}

func beginControl() {
	/*!	//imgui.SetNextWindowPos(imgui.Vec2{X: 400, Y: 25})
	//imgui.SetNextWindowSize(imgui.Vec2{X: 780, Y: 250})
	imgui.Begin("Control")
	imgui.Text("FCtrl")
	imgui.Separator()
	beginFCtrl()
	imgui.Text("MAC Commands")
	beginMACCommands()
	imgui.Separator()
	imgui.End()*/
}
