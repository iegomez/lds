package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	c "strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"

	l "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/scartill/giox"
	xmat "github.com/scartill/giox/material"
)

type encodedType struct {
	Name     string  `toml:"name"`
	Value    float64 `toml:"value"`
	MaxValue float64 `toml:"max_value"`
	MinValue float64 `toml:"min_value"`
	IsFloat  bool    `toml:"is_float"`
	NumBytes int     `toml:"num_bytes"`
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

type encodedTypeWidgets struct {
	Name         widget.Editor
	NumBytes     widget.Editor
	IsFloat      widget.Bool
	DeleteButton widget.Clickable
	Value        widget.Editor
	MaxValue     widget.Editor
	MinValue     widget.Editor
}

// MaxEncodedTypes defines max number of simulated data types to send
const MaxEncodedTypes = 10

var (
	rawBytesEditor     widget.Editor
	sendRawCheckbox    widget.Bool
	useEncoderCheckBox widget.Bool
	openEncoderButton  widget.Clickable
	fPortEditor        widget.Editor
	intervalEditor     widget.Editor
	repeatCheckbox     widget.Bool
	sendDataButton     widget.Clickable
	stopDataButton     widget.Clickable
	addEncodedType     widget.Clickable

	encodedWidgets []encodedTypeWidgets

	funcEditor        widget.Editor
	objEditor         widget.Editor
	clearScriptEditor widget.Clickable
	closeScriptEditor widget.Clickable
)

func createDataForm() {
	encodedWidgets = make([]encodedTypeWidgets, MaxEncodedTypes)
	openScript = false
}

func dataResetGuiValues() {
	rawBytesEditor.SetText(config.RawPayload.Payload)
	sendRawCheckbox.Value = config.RawPayload.UseRaw
	useEncoderCheckBox.Value = config.RawPayload.UseEncoder
	fPortEditor.SetText(c.Itoa(config.RawPayload.FPort))
	intervalEditor.SetText(fmt.Sprintf("%d", interval))
	repeatCheckbox.Value = repeat

	for i := 0; i < len(config.EncodedType); i++ {
		encodedWidgets[i].Name.SetText(config.EncodedType[i].Name)
		encodedWidgets[i].NumBytes.SetText(c.Itoa(config.EncodedType[i].NumBytes))
		encodedWidgets[i].IsFloat.Value = config.EncodedType[i].IsFloat
		encodedWidgets[i].Value.SetText(
			fmt.Sprintf("%f", config.EncodedType[i].Value))
		encodedWidgets[i].MaxValue.SetText(
			fmt.Sprintf("%f", config.EncodedType[i].MaxValue))
		encodedWidgets[i].MinValue.SetText(
			fmt.Sprintf("%f", config.EncodedType[i].MinValue))

	}

	funcEditor.SetText(config.RawPayload.Script)
	objEditor.SetText(config.RawPayload.Obj)
}

func dataForm(th *material.Theme) l.FlexChild {
	config.RawPayload.Payload = rawBytesEditor.Text()
	config.RawPayload.UseRaw = sendRawCheckbox.Value
	config.RawPayload.UseEncoder = useEncoderCheckBox.Value
	extractInt(&fPortEditor, &config.RawPayload.FPort, 0)
	extractInt32(&intervalEditor, &interval, 1)
	repeat = repeatCheckbox.Value

	for i := 0; i < len(config.EncodedType); i++ {
		config.EncodedType[i].Name = encodedWidgets[i].Name.Text()
		extractInt(&encodedWidgets[i].NumBytes, &config.EncodedType[i].NumBytes, 0)
		config.EncodedType[i].IsFloat = encodedWidgets[i].IsFloat.Value
		extractFloat(&encodedWidgets[i].Value, &config.EncodedType[i].Value, 0)
		extractFloat(&encodedWidgets[i].MaxValue, &config.EncodedType[i].MaxValue, 0)
		extractFloat(&encodedWidgets[i].MinValue, &config.EncodedType[i].MinValue, 0)
	}

	config.RawPayload.Script = funcEditor.Text()
	config.RawPayload.Obj = objEditor.Text()

	for openEncoderButton.Clicked() {
		openScript = true
	}

	if !running {
		for sendDataButton.Clicked() {
			go run()
		}
	}

	if repeat && running {
		for stopDataButton.Clicked() {
			running = false
		}
	}

	for addEncodedType.Clicked() {
		et := &encodedType{
			Name:     "New type",
			Value:    0,
			MaxValue: 0,
			MinValue: 0,
			NumBytes: 0,
		}
		config.EncodedType = append(config.EncodedType, et)
		log.Println("added new type")
	}

	for i := 0; i < len(config.EncodedType); i++ {
		for encodedWidgets[i].DeleteButton.Clicked() {
			if len(config.EncodedType) == 1 {
				config.EncodedType = make([]*encodedType, 0)
			} else {
				copy(config.EncodedType[i:], config.EncodedType[i+1:])
				config.EncodedType[len(config.EncodedType)-1] = &encodedType{}
				config.EncodedType = config.EncodedType[:len(config.EncodedType)-1]
			}
		}
	}

	for clearScriptEditor.Clicked() {
		config.RawPayload.Script = defaultScript
		funcEditor.SetText(config.RawPayload.Script)
	}

	for closeScriptEditor.Clicked() {
		openScript = false
	}

	widgets := make([]l.FlexChild, 0)
	if !openScript {
		widgets = append(widgets,
			xmat.RigidSection(th, "Raw Data"),
			xmat.RigidEditor(th, "Raw bytes in hex", "DEADBEEF", &rawBytesEditor),
			xmat.RigidCheckBox(th, "Send raw", &sendRawCheckbox),
			xmat.RigidCheckBox(th, "Use encoder", &useEncoderCheckBox),
			xmat.RigidButton(th, "Open encoder", &openEncoderButton),
			xmat.RigidEditor(th, "fPort", "<fport>", &fPortEditor),
			l.Rigid(func(gtx l.Context) l.Dimensions {
				return l.Flex{Axis: l.Horizontal}.Layout(gtx,
					xmat.RigidEditor(th, "Interval", "<inteval>", &intervalEditor),
					xmat.RigidCheckBox(th, "Send every X seconds", &repeatCheckbox),
				)
			}),
		)

		if !running {
			widgets = append(widgets, xmat.RigidButton(th, "Send data", &sendDataButton))
		}

		if repeat && running {
			widgets = append(widgets, xmat.RigidButton(th, "Stop", &stopDataButton))
		}

		widgets = append(widgets,
			xmat.RigidSection(th, "Encoded data"),
			xmat.RigidButton(th, "Add encoded type", &addEncodedType),
		)

		for i := 0; i < len(config.EncodedType); i++ {
			etw := &encodedWidgets[i]
			widgets = append(widgets,
				xmat.RigidSeparator(th, &giox.Separator{}),
				l.Rigid(func(gtx l.Context) l.Dimensions {
					return l.Flex{Axis: l.Horizontal}.Layout(gtx,
						xmat.RigidEditor(th, "Name", "<name>", &etw.Name),
						xmat.RigidEditor(th, "Bytes", "<bytes>", &etw.NumBytes),
						xmat.RigidCheckBox(th, "Float", &etw.IsFloat),
						xmat.RigidButton(th, "Delete", &etw.DeleteButton),
					)
				}),
				l.Rigid(func(gtx l.Context) l.Dimensions {
					return l.Flex{Axis: l.Horizontal}.Layout(gtx,
						xmat.RigidEditor(th, "Value", "0", &etw.Value),
						xmat.RigidEditor(th, "Max", "0", &etw.MaxValue),
						xmat.RigidEditor(th, "Min", "0", &etw.MinValue),
					)
				}),
			)
		}
	} else {
		widgets = append(widgets,
			xmat.RigidSection(th, "JS Encoder"),
			xmat.RigidLabel(th, `If "Use encoder" is checked, you may write a function that accepts a JS object`),
			xmat.RigidLabel(th, `and returns a byte array that'll be used as the raw bytes when sending data.`),
			xmat.RigidLabel(th, `The function must be named Encode and accept a port and JS object.`),
			xmat.RigidEditor(th, "Encoder Function", "JS", &funcEditor),
			xmat.RigidLabel(th, `JS Object:`),
			xmat.RigidEditor(th, "Encoder object", "JS", &objEditor),
			xmat.RigidButton(th, "Clear", &clearScriptEditor),
			xmat.RigidButton(th, "Close", &closeScriptEditor),
		)
	}

	inset := l.Inset{Left: unit.Dp(30)}
	return l.Rigid(func(gtx l.Context) l.Dimensions {
		return inset.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx, widgets...)
		})
	})
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
