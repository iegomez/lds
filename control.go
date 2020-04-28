package main

import (
	"github.com/brocaar/lorawan"

    "gioui.org/layout"
    "gioui.org/widget/material"
)

type macCommandItem struct {
	MACCommand lorawan.MACCommand
	Use        bool
}

//We need some vars to store string representations of payload fields.
var ResetIndMinorS string
var RekeyIndMinorS string
var BatteryS string
var MarginS string

var fCtrl lorawan.FCtrl

//List of all available mac commands and their payloads.
var macCommands = []*macCommandItem{
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID: lorawan.ResetInd,
			Payload: &lorawan.ResetIndPayload{
				DevLoRaWANVersion: lorawan.Version{
					Minor: 0,
				},
			},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.LinkCheckReq,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.LinkADRAns,
			Payload: &lorawan.LinkADRAnsPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DutyCycleAns,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RXParamSetupAns,
			Payload: &lorawan.RXParamSetupAnsPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DevStatusAns,
			Payload: &lorawan.DevStatusAnsPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.NewChannelAns,
			Payload: &lorawan.NewChannelAnsPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RXTimingSetupAns,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.TXParamSetupAns,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DLChannelAns,
			Payload: &lorawan.DLChannelAnsPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RekeyInd,
			Payload: &lorawan.RekeyIndPayload{},
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.ADRParamSetupAns,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.DeviceTimeReq,
			Payload: nil,
		},
	},
	&macCommandItem{
		MACCommand: lorawan.MACCommand{
			CID:     lorawan.RejoinParamSetupAns,
			Payload: &lorawan.RejoinParamSetupAnsPayload{},
		},
	},
}

func controlForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
	return layout.Rigid( func() {
		th.Caption("control").Layout(gtx)
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
				imgui.InputTextV("DevLoRaWANVersion##ResetIndPayload-Minor", &ResetIndMinorS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(ResetIndMinorS, 1, &payload.DevLoRaWANVersion.Minor))
			case lorawan.LinkADRAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.LinkADRAnsPayload)
				imgui.Checkbox("ChannelMaskACK##LinkADRAns-ChannelMaskACK", &payload.ChannelMaskACK)
				imgui.SameLine()
				imgui.Checkbox("DateRateACK##LinkADRAns-DateRateACK", &payload.DataRateACK)
				imgui.SameLine()
				imgui.Checkbox("PowerACK##LinkADRAns-PowerACK", &payload.PowerACK)
			case lorawan.RXParamSetupAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RXParamSetupAnsPayload)
				imgui.Checkbox("ChannelACK##RXParamSetupAns-ChannelMaskACK", &payload.ChannelACK)
				imgui.SameLine()
				imgui.Checkbox("RX2DateRateACK##RXParamSetupAns-RX2DateRateACK", &payload.RX2DataRateACK)
				imgui.SameLine()
				imgui.Checkbox("RX1DROffsetACK##RXParamSetupAns-RX1DROffsetACK", &payload.RX1DROffsetACK)
			case lorawan.DevStatusAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.DevStatusAnsPayload)
				imgui.InputTextV("Battery##devStatusAnsPayload-Battery", &BatteryS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(BatteryS, 2, &payload.Battery))
				imgui.SameLine()
				imgui.InputTextV("Margin##devStatusAnsPayload-Margin", &MarginS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt8(MarginS, 2, &payload.Margin))
			case lorawan.NewChannelAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.NewChannelAnsPayload)
				imgui.Checkbox("ChannelFrequencyOK##NewChannelAns-ChannelFrequencyOK", &payload.ChannelFrequencyOK)
				imgui.SameLine()
				imgui.Checkbox("DataRateRangeOK##NewChannelAns-DataRangeOK", &payload.DataRateRangeOK)
			case lorawan.DLChannelAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.DLChannelAnsPayload)
				imgui.Checkbox("ChannelFrequencyOK##DLChannelAns-ChannelFrequencyOK", &payload.ChannelFrequencyOK)
				imgui.SameLine()
				imgui.Checkbox("UplinkFrequencyExists##DLChannelAns-UplinkFrequencyExists", &payload.UplinkFrequencyExists)
			case lorawan.RekeyInd:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RekeyIndPayload)
				imgui.InputTextV("DevLoRaWANVersion##RekeyIndPayload-Minor", &RekeyIndMinorS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleUint8(RekeyIndMinorS, 1, &payload.DevLoRaWANVersion.Minor))
			case lorawan.RejoinParamSetupAns:
				imgui.SameLine()
				payload := macCommand.MACCommand.Payload.(*lorawan.RejoinParamSetupAnsPayload)
				imgui.Checkbox("TimeOK##RejoinParamSetupAns-TimeOK", &payload.TimeOK)
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
