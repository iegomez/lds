package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"

    "gioui.org/layout"
    "gioui.org/widget/material"
)

type encodedType struct {
	Name     string  `toml:"name"`
	Value    float64 `toml:"value"`
	MaxValue float64 `toml:"max_value"`
	MinValue float64 `toml:"min_value"`
	IsFloat  bool    `toml:"is_float"`
	NumBytes int     `toml:"num_bytes"`
	//String representations.
	ValueS    string `toml:"-"`
	MinValueS string `toml:"-"`
	MaxValueS string `toml:"-"`
	NumBytesS string `toml:"-"`
}

//rawPayload holds optional raw bytes payload (hex encoded).
type rawPayload struct {
	Payload     string `toml:"payload"`
	UseRaw      bool   `toml:"use_raw"`
	Script      string `toml:"script"`
	UseEncoder  bool   `toml:"use_encoder"`
	MaxExecTime int    `toml:"max_exec_time"`
	Obj         string `toml:"js_object"`
	FPort       int    `toml:"fport"`
	FPortS      string `toml:"-"`
}

var openScript bool
var defaultScript = `
// Encode encodes the given object into an array of bytes.
//  - fPort contains the LoRaWAN fPort number
//  - obj is an object, e.g. {"temperature": 22.5}
// The function must return an array of bytes, e.g. [225, 230, 255, 0]
function Encode(fPort, obj) {
	return [];
}
`
func dataForm(gtx *layout.Context, th *material.Theme) layout.FlexChild {
	return layout.Rigid( func() {
		material.Caption(th, "data").Layout(gtx)
	})
}

func beginDataForm() {
/*! //imgui.SetNextWindowPos(imgui.Vec2{X: 400, Y: 285})
	//imgui.SetNextWindowSize(imgui.Vec2{X: 780, Y: 355})
	imgui.Begin("Data")
	imgui.Text("Raw data")
	imgui.PushItemWidth(150.0)
	imgui.InputTextV("Raw bytes in hex", &config.RawPayload.Payload, imgui.InputTextFlagsCharsHexadecimal, nil)
	imgui.SameLine()
	imgui.Checkbox("Send raw", &config.RawPayload.UseRaw)
	imgui.SameLine()
	imgui.Checkbox("Use encoder", &config.RawPayload.UseEncoder)
	imgui.SameLine()
	if imgui.Button("Open encoder") {
		openScript = true
	}
	imgui.InputTextV(fmt.Sprintf("fPort    ##fport"), &config.RawPayload.FPortS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(config.RawPayload.FPortS, 10, &config.RawPayload.FPort))
	imgui.SliderInt("X", &interval, 1, 60)
	imgui.SameLine()
	imgui.Checkbox("Send every X seconds", &repeat)
	if !running {
		if imgui.Button("Send data") {
			go run()
		}
	}
	if repeat && running {
		if imgui.Button("Stop") {
			running = false
		}
	}

	imgui.Separator()

	imgui.Text("Encoded data")
	if imgui.Button("Add encoded type") {
		et := &encodedType{
			Name:      "New type",
			ValueS:    "0",
			MaxValueS: "0",
			MinValueS: "0",
			NumBytesS: "0",
		}
		config.EncodedType = append(config.EncodedType, et)
		log.Println("added new type")
	}

	for i := 0; i < len(config.EncodedType); i++ {
		delete := false
		imgui.Separator()
		imgui.InputText(fmt.Sprintf("Name     ##%d", i), &config.EncodedType[i].Name)
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Bytes    ##%d", i), &config.EncodedType[i].NumBytesS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleInt(config.EncodedType[i].NumBytesS, 10, &config.EncodedType[i].NumBytes))
		imgui.SameLine()
		imgui.Checkbox(fmt.Sprintf("Float##%d", i), &config.EncodedType[i].IsFloat)
		imgui.SameLine()
		if imgui.Button(fmt.Sprintf("Delete##%d", i)) {
			delete = true
		}
		imgui.InputTextV(fmt.Sprintf("Value    ##%d", i), &config.EncodedType[i].ValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].ValueS, &config.EncodedType[i].Value))
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Max value##%d", i), &config.EncodedType[i].MaxValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].MaxValueS, &config.EncodedType[i].MaxValue))
		imgui.SameLine()
		imgui.InputTextV(fmt.Sprintf("Min value##%d", i), &config.EncodedType[i].MinValueS, imgui.InputTextFlagsCharsDecimal|imgui.InputTextFlagsCallbackAlways, handleFloat64(config.EncodedType[i].MinValueS, &config.EncodedType[i].MinValue))
		if delete {
			if len(config.EncodedType) == 1 {
				config.EncodedType = make([]*encodedType, 0)
			} else {
				copy(config.EncodedType[i:], config.EncodedType[i+1:])
				config.EncodedType[len(config.EncodedType)-1] = &encodedType{}
				config.EncodedType = config.EncodedType[:len(config.EncodedType)-1]
			}
		}
	}
	imgui.Separator()
	beginScript()
	imgui.End()*/
}

func beginScript() {
/*!	if openScript {
		imgui.OpenPopup("JS encoder")
		openScript = false
	}
	imgui.SetNextWindowPos(imgui.Vec2{X: (float32(config.Window.Width) / 2) - 370.0, Y: (float32(config.Window.Height) / 2) - 200.0})
	imgui.SetNextWindowSize(imgui.Vec2{X: 740, Y: 600})
	if imgui.BeginPopupModal("JS encoder") {
		imgui.Text(`If "Use encoder" is checked, you may write a function that accepts a JS object`)
		imgui.Text(`and returns a byte array that'll be used as the raw bytes when sending data.`)
		imgui.Text(`The function must be named Encode and accept a port and JS object.`)
		imgui.InputTextMultilineV("##encoder-function", &config.RawPayload.Script, imgui.Vec2{X: 710, Y: 300}, imgui.InputTextFlagsAllowTabInput, nil)
		imgui.Separator()
		imgui.Text("JS object:")
		imgui.InputTextMultilineV("##encoder-object", &config.RawPayload.Obj, imgui.Vec2{X: 710, Y: 140}, 0, nil)
		if imgui.Button("Clear##encoder-cancel") {
			config.RawPayload.Script = defaultScript
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Close##encoder-close") {
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}*/
}

// EncodeToBytes encodes the payload to a slice of bytes.
// Taken from github.com/brocaar/lora-app-server.
func EncodeToBytes() (b []byte, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%s", caught)
		}
	}()

	script := config.RawPayload.Script + "\n\nEncode(fPort, obj);\n"

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.SetStackDepthLimit(32)
	var jsonData interface{}
	err = json.Unmarshal([]byte(config.RawPayload.Obj), &jsonData)
	if err != nil {
		log.Errorf("couldn't unmarshal object: %s", err)
		return nil, err
	}
	log.Debugf("JS object: %v", jsonData)
	vm.Set("obj", jsonData)
	vm.Set("fPort", config.RawPayload.FPort)

	go func() {
		time.Sleep(time.Duration(config.RawPayload.MaxExecTime) * time.Millisecond)
		vm.Interrupt <- func() {
			panic(errors.New("execution timeout"))
		}
	}()

	var val otto.Value
	val, err = vm.Run(script)
	if err != nil {
		return nil, errors.Wrap(err, "js vm error")
	}
	if !val.IsObject() {
		return nil, errors.New("function must return an array")
	}

	var out interface{}
	out, err = val.Export()
	if err != nil {
		return nil, errors.Wrap(err, "export error")
	}

	return interfaceToByteSlice(out)
}

// Taken from github.com/brocaar/lora-app-server.
func interfaceToByteSlice(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("value must not be nil")
	}

	if reflect.TypeOf(obj).Kind() != reflect.Slice {
		return nil, errors.New("value must be an array")
	}

	s := reflect.ValueOf(obj)
	l := s.Len()

	var out []byte
	for i := 0; i < l; i++ {
		var b int64

		el := s.Index(i).Interface()
		switch v := el.(type) {
		case int:
			b = int64(v)
		case uint:
			b = int64(v)
		case uint8:
			b = int64(v)
		case int8:
			b = int64(v)
		case uint16:
			b = int64(v)
		case int16:
			b = int64(v)
		case uint32:
			b = int64(v)
		case int32:
			b = int64(v)
		case uint64:
			b = int64(v)
			if uint64(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %d", v)
			}
		case int64:
			b = int64(v)
		case float32:
			b = int64(v)
			if float32(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %f", v)
			}
		case float64:
			b = int64(v)
			if float64(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %f", v)
			}
		default:
			return nil, fmt.Errorf("array value must be an array of ints or floats, got: %T", el)
		}

		if b < 0 || b > 255 {
			return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %d", b)
		}

		out = append(out, byte(b))
	}

	return out, nil
}
