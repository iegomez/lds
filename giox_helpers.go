package main

import (
  	"strconv"

    "gioui.org/widget"
    "github.com/scartill/giox"
)

func extractInt(edit *widget.Editor, value *int, onError int) {
	parsed, err := strconv.Atoi(edit.Text())

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}

func extractFloat(edit *widget.Editor, value *float64, onError float64) {
	parsed, err := strconv.ParseFloat(edit.Text(), 64)

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}

func extractIntCombo(combo *giox.Combo, value *int, onError int) {
	if !combo.HasSelected() {
		*value = onError
		return
	}

	parsed, err := strconv.Atoi(combo.SelectedText())

	if err == nil {
		*value = parsed
	} else {
		*value = onError
	}
}
