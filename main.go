package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	l "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	xmat "github.com/scartill/giox/material"
)

// This holds the "console" visible text, line number and history (so we can dump everything even when console has been cleared).
type outputWriter struct {
	Text    string
	Counter int
	History string
}

// Write just appends to text and history, using the counter as line number for the text.
// It also allows outputWriter to implement the Writer interface so it may be passed to the logger.
func (o *outputWriter) Write(p []byte) (n int, err error) {
	o.Counter++
	o.Text = fmt.Sprintf("%s%05d  %s", o.Text, o.Counter, string(p))
	o.History = fmt.Sprintf("%s%s", o.History, string(p))
	return len(p), nil
}

// The writer instance
var ow = &outputWriter{Text: "", Counter: 0}

// Message sending control and status.
var (
	repeat   bool
	running  bool
	stop     bool
	sendOnce bool
	interval int32
)

func beginMenu() {
	/*!	if imgui.BeginMainMenuBar() {
	    if imgui.BeginMenu("File") {

	        if imgui.MenuItem("Open") {
	            openFile = true
	            var err error
	            files, err = ioutil.ReadDir("./confs/")
	            if err != nil {
	                log.Errorf("couldn't list files: %s", err)
	            }
	        }

	        if imgui.MenuItem("Save") {
	            saveFile = true
	        }

	        if imgui.MenuItem("Provision") {
	            openProvisioner = true
	        }

	        imgui.EndMenu()
	    }
	    if imgui.BeginMenu("Console") {
	        if imgui.MenuItem("Clear") {
	            ow.Text = ""
	            ow.Counter = 0
	        }

	        if imgui.MenuItem("Copy") {
	            err := clipboard.WriteAll(ow.Text)
	            if err != nil {
	                log.Errorf("copy error: %s", err)
	            }
	        }

	        if imgui.MenuItem("Dump history") {
	            writeHistory()
	        }

	        imgui.EndMenu()
	    }
	    if imgui.BeginMenu("Log level") {
	        if imgui.MenuItem("Debug") {
	            setLevel(log.DebugLevel)
	        }
	        if imgui.MenuItem("Info") {
	            setLevel(log.InfoLevel)
	        }
	        if imgui.MenuItem("Warning") {
	            setLevel(log.WarnLevel)
	        }
	        if imgui.MenuItem("Error") {
	            setLevel(log.ErrorLevel)
	        }

	        imgui.EndMenu()
	    }
	    imgui.EndMainMenuBar()
	}*/
}

func beginOpenFile() {
	/*!	if openFile {
	          imgui.OpenPopup("Select file")
	          openFile = false
	      }
	      imgui.SetNextWindowPos(imgui.Vec2{X: float32(config.Window.Width-190) / 2, Y: float32(config.Window.Height-90) / 2})
	      imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 180})
	      imgui.PushItemWidth(250.0)
	      if imgui.BeginPopupModal("Select file") {
	          if imgui.BeginComboV("Select", *confFile, 0) {
	              for _, f := range files {
	                  filename := fmt.Sprintf("confs/%s", f.Name())
	                  if !strings.Contains(filename, ".toml") {
	                      continue
	                  }
	                  if imgui.SelectableV(filename, *confFile == filename, 0, imgui.Vec2{}) {
	                      *confFile = filename
	                  }
	              }
	              imgui.EndCombo()
	          }
	          imgui.Separator()
	          if imgui.Button("Cancel") {
	              imgui.CloseCurrentPopup()
	          }
	          imgui.SameLine()
	          if imgui.Button("Import") {
	              //Import file.
	  			importConf()
	  			resetGuiValues()
	  			setDevice()
	              imgui.CloseCurrentPopup()
	              //Close popup.
	          }
	          imgui.EndPopup()
	      }*/
}

func beginSaveFile() {
	/*!	if saveFile {
	      imgui.OpenPopup("Save file")
	      saveFile = false
	  }
	  imgui.SetNextWindowPos(imgui.Vec2{X: float32(config.Window.Width-190) / 2, Y: float32(config.Window.Height-90) / 2})
	  imgui.SetNextWindowSize(imgui.Vec2{X: 380, Y: 180})
	  imgui.PushItemWidth(250.0)
	  if imgui.BeginPopupModal("Save file") {

	      imgui.InputText("Name", &saveFilename)
	      imgui.Separator()
	      if imgui.Button("Cancel") {
	          imgui.CloseCurrentPopup()
	      }
	      imgui.SameLine()
	      if imgui.Button("Save") {
	          //Import file.
	          exportConf(fmt.Sprintf("confs/%s", saveFilename))
	          imgui.CloseCurrentPopup()
	          //Close popup.
	      }
	      imgui.EndPopup()
	  }*/
}

func resetGuiValues() {
	mqttResetGuiValue()
	forwarderResetGuiValues()
	loraResetGuiValues()
	deviceResetGuiValues()
	macResetGuiValues()
	dataResetGuiValues()
}

var (
	mqttButton      widget.Clickable
	forwarderButton widget.Clickable
	deviceButton    widget.Clickable
	loraButton      widget.Clickable
	controlButton   widget.Clickable
	dataButton      widget.Clickable
	outputButton    widget.Clickable

	tabIndex uint
)

func mainWindow(gtx l.Context, th *material.Theme) {
	/*!
	  beginMenu()
	  beginProvisioner()
	*/

	for mqttButton.Clicked() {
		tabIndex = 0
	}

	for forwarderButton.Clicked() {
		tabIndex = 1
	}

	for deviceButton.Clicked() {
		tabIndex = 2
	}

	for loraButton.Clicked() {
		tabIndex = 3
	}

	for controlButton.Clicked() {
		tabIndex = 4
	}

	for dataButton.Clicked() {
		tabIndex = 5
	}

	for outputButton.Clicked() {
		tabIndex = 6
	}

	tabsWidget := l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			xmat.RigidButton(th, "MQTT", &mqttButton),
			xmat.RigidButton(th, "Forwarder", &forwarderButton),
			xmat.RigidButton(th, "Device", &deviceButton),
			xmat.RigidButton(th, "LoRa", &loraButton),
			xmat.RigidButton(th, "Control", &controlButton),
			xmat.RigidButton(th, "Data", &dataButton),
			xmat.RigidButton(th, "Output", &outputButton),
		)
	})

	wMqttForm := mqttForm(th)
	wForwarderForm := forwarderForm(th)
	wDeviceForm := deviceForm(th)
	wLoraForm := loRaForm(th)
	wControlForm := controlForm(th)
	wDataForm := dataForm(th)
	wOutputForm := outputForm(th)

	var selectedWidget l.FlexChild
	switch tabIndex {
	case 0:
		selectedWidget = wMqttForm
	case 1:
		selectedWidget = wForwarderForm
	case 2:
		selectedWidget = wDeviceForm
	case 3:
		selectedWidget = wLoraForm
	case 4:
		selectedWidget = wControlForm
	case 5:
		selectedWidget = wDataForm
	case 6:
		selectedWidget = wOutputForm
	}

	l.NW.Layout(gtx, func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Horizontal}.Layout(gtx,
			tabsWidget,
			selectedWidget,
		)
	})
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())

	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := l.NewContext(&ops, e)
			mainWindow(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

func main() {
	/*!	runtime.LockOSThread() */

	mw := io.MultiWriter(ow, os.Stderr)
	log.SetOutput(mw)

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()

	createLoRaForm()
	createDeviceForm()
	createDataForm()
	tabIndex = 0

	importConf()
	resetGuiValues()
	setDevice()

	go func() {
		w := app.NewWindow(app.Size(unit.Px(1600), unit.Px(1024)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}
