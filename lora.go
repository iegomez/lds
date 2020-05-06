package main

import (
    /*!	"strconv" */

    lwband "github.com/brocaar/lorawan/band"

    "gioui.org/layout"
    "gioui.org/widget/material"
    "github.com/scartill/giox"
    xmat "github.com/scartill/giox/material"
)

// Bands and data rate options.
var (
    bandwidths    = []int{50, 125, 250, 500}
    spreadFactors = []int{7, 8, 9, 10, 11, 12}
    bands         = []lwband.Name{
        lwband.AS_923,
        lwband.AU_915_928,
        lwband.CN_470_510,
        lwband.CN_779_787,
        lwband.EU_433,
        lwband.EU_863_870,
        lwband.IN_865_867,
        lwband.KR_920_923,
        lwband.US_902_928,
        lwband.RU_864_870,
    }
)

type band struct {
    Name lwband.Name `toml:"name"`
}

type dataRate struct {
    Bandwidth    int    `toml:"bandwith"`
    SpreadFactor int    `toml:"spread_factor"`
    BitRate      int    `toml:"bit_rate"`
    BitRateS     string `toml:"-"`
}

type rxInfo struct {
    Channel   int     `toml:"channel"`
    CodeRate  string  `toml:"code_rate"`
    CrcStatus int     `toml:"crc_status"`
    Frequency int     `toml:"frequency"`
    LoRaSNR   float64 `toml:"lora_snr"`
    RfChain   int     `toml:"rf_chain"`
    Rssi      int     `toml:"rssi"`
    //String representations for numeric values so that we can manage them with input texts.
    ChannelS   string `toml:"-"`
    CrcStatusS string `toml:"-"`
    FrequencyS string `toml:"-"`
    LoRASNRS   string `toml:"-"`
    RfChainS   string `toml:"-"`
    RssiS      string `toml:"-"`
}

var (
    loraBandCombo = giox.MakeCombo(
        []string {
            string(lwband.AS_923),
            string(lwband.AU_915_928),
            string(lwband.CN_470_510),
            string(lwband.CN_779_787),
            string(lwband.EU_433),
            string(lwband.EU_863_870),
            string(lwband.IN_865_867),
            string(lwband.KR_920_923),
            string(lwband.US_902_928),
            string(lwband.RU_864_870),
        },
        "Select a band",
    )
)

func loRaForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
    widgets := []layout.FlexChild{
        xmat.RigidSection(gtx, th, "LoRa Configuration"),
   		layout.Rigid(func() {
			xmat.Combo(th).Layout(gtx, &loraBandCombo)
		}),
    }

    return layout.Rigid(func() {
        layout.Flex{Axis: layout.Vertical}.Layout(gtx, widgets...)
    })
}

func beginLoRaForm() {
    /*!	//imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: 650})
    //imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 265})
    imgui.Begin("LoRa Configuration")
    imgui.PushItemWidth(250.0)
    if imgui.BeginCombo("Band", string(config.Band.Name)) {
        for _, band := range bands {
            if imgui.SelectableV(string(band), band == config.Band.Name, 0, imgui.Vec2{}) {
                config.Band.Name = band
            }
        }
        imgui.EndCombo()
    }

    if imgui.BeginCombo("Bandwidth", strconv.Itoa(config.DR.Bandwidth)) {
        for _, bandwidth := range bandwidths {
            if imgui.SelectableV(strconv.Itoa(bandwidth), bandwidth == config.DR.Bandwidth, 0, imgui.Vec2{}) {
                config.DR.Bandwidth = bandwidth
            }
        }
        imgui.EndCombo()
    }

    if imgui.BeginCombo("SpreadFactor", strconv.Itoa(config.DR.SpreadFactor)) {
        for _, sf := range spreadFactors {
            if imgui.SelectableV(strconv.Itoa(sf), sf == config.DR.SpreadFactor, 0, imgui.Vec2{}) {
                config.DR.SpreadFactor = sf
            }
        }
        imgui.EndCombo()
    }

    imgui.InputTextV("Bit rate", &config.DR.BitRateS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.DR.BitRateS, 6, &config.DR.BitRate))

    imgui.InputTextV("Channel", &config.RXInfo.ChannelS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.ChannelS, 10, &config.RXInfo.Channel))

    imgui.InputTextV("CrcStatus", &config.RXInfo.CrcStatusS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.CrcStatusS, 10, &config.RXInfo.CrcStatus))

    imgui.InputTextV("Frequency", &config.RXInfo.FrequencyS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.FrequencyS, 14, &config.RXInfo.Frequency))

    imgui.InputTextV("LoRaSNR", &config.RXInfo.LoRASNRS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.RXInfo.LoRASNRS, &config.RXInfo.LoRaSNR))

    imgui.InputTextV("RfChain", &config.RXInfo.RfChainS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(config.RXInfo.RfChainS, 10, &config.RXInfo.RfChain))

    imgui.InputTextV("Rssi", &config.RXInfo.RssiS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways|imgui.InputTextFlagsCallbackCharFilter, handleInt(config.RXInfo.RssiS, 10, &config.RXInfo.Rssi))

    imgui.End()*/
}
