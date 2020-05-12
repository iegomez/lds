package main

import (
    "flag"
    "fmt"
    "io"
    "os"

    log "github.com/sirupsen/logrus"

    "gioui.org/app"
    "gioui.org/io/system"
    "gioui.org/layout"
    "gioui.org/unit"
    "gioui.org/widget/material"
    "gioui.org/font/gofont"
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
}

func mainWindow(gtx *layout.Context, th *material.Theme) {
    /*!
        beginMenu()
        beginProvisioner()
    */

    wMqttForm := mqttForm(gtx, th)
    wForwarderForm := forwarderForm(gtx, th)
    wDeviceForm := deviceForm(gtx, th)
    wLoraForm := loRaForm(gtx, th)
    wControlForm := controlForm(gtx, th)
    wDataForm := dataForm(gtx, th)
    wOutputForm := outputForm(gtx, th)

    inset := layout.UniformInset(unit.Px(10))
    wLeft := func() {
        inset.Layout(gtx, func() {
            layout.Flex{Axis: layout.Vertical}.Layout(gtx, wMqttForm, wForwarderForm, wLoraForm)
        })
    }

    wMid := func() {
        inset.Layout(gtx, func() {
            layout.Flex{Axis: layout.Vertical}.Layout(gtx, wDeviceForm)
        })
    }

    wRight := func() {
        inset.Layout(gtx, func() {
            layout.Flex{Axis: layout.Vertical}.Layout(gtx, wControlForm, wDataForm, wOutputForm)
        })
    }

    inset.Layout(gtx, func() {
        layout.W.Layout(gtx, func() {
            layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
                layout.Rigid(wLeft),
                layout.Rigid(wMid),
                layout.Rigid(wRight),
            )
        })
    })
}

func loop(w *app.Window) error {
	gofont.Register()
	th := material.NewTheme()
	gtx := new(layout.Context)

	for e := range w.Events() {
		if e, ok := e.(system.FrameEvent); ok {
			gtx.Reset(e.Queue, e.Config, e.Size)
			mainWindow(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
	
	return nil
}

func main() {
/*!	runtime.LockOSThread() */

    mw := io.MultiWriter(ow, os.Stderr)
    log.SetOutput(mw)

    confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
    flag.Parse()

    createLoRaForm()
    createDeviceForm()

	importConf()
	resetGuiValues()
    setDevice()
        
    go func() {
        p1024 := unit.Px(1024)
        w := app.NewWindow(app.Size(p1024, p1024))
        if err := loop(w); err != nil {
            log.Fatal(err)
        }
    }()
    app.Main()
}
