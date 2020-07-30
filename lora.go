package main

import (
	"strconv"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	lwband "github.com/brocaar/lorawan/band"
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
	loraBandCombo     giox.Combo
	bandwidthCombo    giox.Combo
	spreadFactorCombo giox.Combo
	bitrateEdit       widget.Editor
	channelEdit       widget.Editor
	crcEdit           widget.Editor
	frequencyEdit     widget.Editor
	snrEdit           widget.Editor
	rfChainEdit       widget.Editor
	rssiEdit          widget.Editor
)

func createLoRaForm() {
	bandItems := make([]string, len(bands))
	for i, v := range bands {
		bandItems[i] = string(v)
	}
	loraBandCombo = giox.MakeCombo(bandItems, "<select band>")

	bandwidthItems := make([]string, len(bandwidths))
	for i, v := range bandwidths {
		bandwidthItems[i] = strconv.Itoa(v)
	}
	bandwidthCombo = giox.MakeCombo(bandwidthItems, "<select bandwidth>")

	spreadFactorItems := make([]string, len(spreadFactors))
	for i, v := range spreadFactors {
		spreadFactorItems[i] = strconv.Itoa(v)
	}
	spreadFactorCombo = giox.MakeCombo(spreadFactorItems, "<select SF>")
}

func loraResetGuiValues() {
	loraBandCombo.SelectItem(string(config.Band.Name))
	bandwidthCombo.SelectItem(strconv.Itoa(config.DR.Bandwidth))
	spreadFactorCombo.SelectItem(strconv.Itoa(config.DR.SpreadFactor))
	bitrateEdit.SetText(config.DR.BitRateS)
	channelEdit.SetText(config.RXInfo.ChannelS)
	crcEdit.SetText(config.RXInfo.CrcStatusS)
	frequencyEdit.SetText(config.RXInfo.FrequencyS)
	snrEdit.SetText(config.RXInfo.LoRASNRS)
	rfChainEdit.SetText(config.RXInfo.RfChainS)
	rssiEdit.SetText(config.RXInfo.RssiS)
}

func loRaForm(th *material.Theme) l.FlexChild {
	config.Band.Name = bands[0]
	if loraBandCombo.HasSelected() {
		for _, v := range bands {
			if loraBandCombo.SelectedText() == string(v) {
				config.Band.Name = v
			}
		}
	}

	extractIntCombo(&bandwidthCombo, &config.DR.Bandwidth, 125)
	extractIntCombo(&spreadFactorCombo, &config.DR.SpreadFactor, 10)

	extractInt(&bitrateEdit, &config.DR.BitRate, 0)
	extractInt(&channelEdit, &config.RXInfo.Channel, 0)
	extractInt(&crcEdit, &config.RXInfo.CrcStatus, 1)
	extractInt(&frequencyEdit, &config.RXInfo.Frequency, 916800000)
	extractFloat(&snrEdit, &config.RXInfo.LoRaSNR, 7.0)
	extractInt(&rfChainEdit, &config.RXInfo.RfChain, 1)
	extractInt(&rssiEdit, &config.RXInfo.Rssi, -57)

	widgets := []l.FlexChild{
		xmat.RigidSection(th, "LoRa Configuration"),
	}

	comboOpen := loraBandCombo.IsExpanded() || bandwidthCombo.IsExpanded() || spreadFactorCombo.IsExpanded()
	if !comboOpen || loraBandCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(th, "Band", &loraBandCombo))
	}

	if !comboOpen || bandwidthCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(th, "Bandwidth", &bandwidthCombo))
	}

	if !comboOpen || spreadFactorCombo.IsExpanded() {
		widgets = append(widgets, labelCombo(th, "SpreadFactor", &spreadFactorCombo))
	}

	if !comboOpen {
		widgets = append(widgets, []l.FlexChild{
			xmat.RigidEditor(th, "Bitrate", "<bitrate>", &bitrateEdit),
			xmat.RigidEditor(th, "Channel", "<channel>", &channelEdit),
			xmat.RigidEditor(th, "CRC", "<checksum>", &crcEdit),
			xmat.RigidEditor(th, "Frequency", "<frequency>", &frequencyEdit),
			xmat.RigidEditor(th, "Lora SNR", "<snr>", &snrEdit),
			xmat.RigidEditor(th, "RF Chain", "<rfchain>", &rfChainEdit),
			xmat.RigidEditor(th, "RSSI", "<RSSI>", &rssiEdit),
		}...)
	}

	inset := l.Inset{Left: unit.Dp(30)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
	})
}
