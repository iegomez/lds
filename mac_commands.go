package main

import (
	"github.com/brocaar/lorawan"
	"github.com/inkyblackness/imgui-go"
)

//List of available mac commands.
var macCommands = []lorawan.CID{
	lorawan.ResetInd,
	lorawan.LinkCheckReq,
	lorawan.LinkADRAns,
	lorawan.DutyCycleAns,
	lorawan.RXParamSetupAns,
	lorawan.DevStatusAns,
	lorawan.NewChannelAns,
	lorawan.RXTimingSetupAns,
	lorawan.TXParamSetupAns,
	lorawan.DLChannelAns,
	lorawan.RekeyInd,
	lorawan.ADRParamSetupAns,
	lorawan.DeviceTimeReq,
	lorawan.RejoinParamSetupAns,
}

//List of mac command payloads for each command.
var mcPayloads = []lorawan.MACCommandPayload{
	&lorawan.ResetIndPayload{},
	&lorawan.LinkCheckAnsPayload{},
	&lorawan.LinkADRAnsPayload{},
	nil,
	&lorawan.RXParamSetupAnsPayload{},
	&lorawan.DevStatusAnsPayload{},
	&lorawan.NewChannelAnsPayload{},
	nil,
	nil,
	&lorawan.DLChannelAnsPayload{},
	&lorawan.RekeyIndPayload{},
	nil,
	nil,
	&lorawan.RejoinParamSetupAnsPayload{},
}

var mcs = []lorawan.MACCommand{
	lorawan.MACCommand{
		CID:     lorawan.ResetInd,
		Payload: &lorawan.ResetIndPayload{},
	},
	lorawan.MACCommand{
		CID:     lorawan.LinkCheckReq,
		Payload: nil,
	},
}

func beginMACCommands() {
	if imgui.BeginCombo("MACCommand", macCommand.String()) {

		for _, mc := range macCommands {
			if imgui.SelectableV(mc.String(), mc == macCommand, 0, imgui.Vec2{}) {
				macCommand = mc
			}
		}
		imgui.EndCombo()
	}
	//Create a mac command form depending on the selected mac command.
	if macCommand != lorawan.CID(0) {
		switch macCommand {
		case lorawan.ResetInd:
			//lorawan.ResetIndPayload{}
		}
	}
}
