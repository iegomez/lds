package main

import (
/*!	"strconv"*/

)
/*!
func maxLength(input string, limit int) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if len(input) >= limit {
			return 1
		}
		return 0
	}
}

func handleInt(input string, limit int, uValue *int) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
			if len(input) > limit || data.EventChar() == rune('.') {
				return 1
			}
			return 0
		}
		v, err := strconv.Atoi(input)
		if err == nil {
			*uValue = v
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleUint8(input string, limit int, uValue *uint8) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
			if len(input) > limit || data.EventChar() == rune('.') {
				return 1
			}
			return 0
		}
		v, err := strconv.ParseInt(input, 10, 8)
		if err == nil {
			*uValue = uint8(v)
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleInt8(input string, limit int, uValue *int8) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
			if len(input) > limit || data.EventChar() == rune('.') {
				return 1
			}
			return 0
		}
		v, err := strconv.ParseInt(input, 10, 8)
		if err == nil {
			*uValue = int8(v)
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleFloat32(input string, uValue *float32) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		v, err := strconv.ParseFloat(input, 32)
		if err == nil {
			*uValue = float32(v)
		} else {
			*uValue = 0
		}
		return 0
	}
}

func handleFloat64(input string, uValue *float64) func(data imgui.InputTextCallbackData) int32 {
	return func(data imgui.InputTextCallbackData) int32 {
		v, err := strconv.ParseFloat(input, 64)
		if err == nil {
			*uValue = v
		} else {
			*uValue = 0
		}
		return 0
	}
}
*/