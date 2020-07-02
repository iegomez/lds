package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	log "github.com/sirupsen/logrus"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	l "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/scartill/giox"
	xmat "github.com/scartill/giox/material"
)

// This holds the "console" visible text, line number and history (so we can dump everything even when console has been cleared).
type outputWriter struct {
	Lines []string
}

// Write just appends to text and history, using the counter as line number for the text.
// It also allows outputWriter to implement the Writer interface so it may be passed to the logger.
func (o *outputWriter) Write(p []byte) (n int, err error) {
	o.Lines = append(o.Lines, string(p))
	return len(p), nil
}

func (o *outputWriter) Text() string {
	text := ""
	for i := 0; i < len(o.Lines); i++ {
		text = fmt.Sprintf("%s%05d  %s", text, i, o.Lines[i])
	}
	return text
}

// The writer instance
var ow = &outputWriter{Lines: []string{}}

// Message sending control and status.
var (
	repeat   bool
	running  bool
	stop     bool
	sendOnce bool
	interval int32
)

// Menu variables
var (
	fileMI           bool
	fileMIBtn        widget.Clickable
	fileOpenBtn      widget.Clickable
	fileSaveBtn      widget.Clickable
	fileProvisionBtn widget.Clickable

	consoleMI        bool
	consoleMIBtn     widget.Clickable
	consoleClearBtn  widget.Clickable
	consoleCopyBtn   widget.Clickable
	consoleDumpBtn   widget.Clickable
	consoleCancelBtn widget.Clickable

	logMI            bool
	logMIBtn         widget.Clickable
	logLvlDebugBtn   widget.Clickable
	logLvlInfoBtn    widget.Clickable
	logLvlWarningBtn widget.Clickable
	logLvlErrorBtn   widget.Clickable
)

var (
	openFileCombo     giox.Combo
	openFileCancelBtn widget.Clickable
	openFileImportBtn widget.Clickable
)

var (
	saveFileEditor    widget.Editor
	saveFileCancelBtn widget.Clickable
	saveFileSaveBtn   widget.Clickable
)

func buildMenu(th *material.Theme) (l.FlexChild, bool) {
	if openFile {
		return buildOpenFile(th)
	}

	if openProvisioner {
		return buildProvisioner(th)
	}

	if saveFile {
		return buildSaveFile(th)
	}

	for fileMIBtn.Clicked() {
		fileMI = true
	}

	for consoleMIBtn.Clicked() {
		consoleMI = true
	}

	for logMIBtn.Clicked() {
		logMI = true
	}

	for fileOpenBtn.Clicked() {
		openFile = true
		var err error
		files, err = ioutil.ReadDir("./confs/")
		if err != nil {
			log.Errorf("couldn't list files: %s", err)
		}

		names := []string{}
		for _, info := range files {
			filename := fmt.Sprintf("confs/%s", info.Name())
			if !strings.Contains(filename, ".toml") {
				continue
			}

			names = append(names, filename)
		}

		openFileCombo = giox.MakeCombo(names, "<filename>")
		fileMI = false
	}

	for fileSaveBtn.Clicked() {
		saveFile = true
		fileMI = false
	}

	for fileProvisionBtn.Clicked() {
		openProvisioner = true
		fileMI = false
	}

	for consoleClearBtn.Clicked() {
		ow.Lines = []string{}
		consoleMI = false
	}

	for consoleClearBtn.Clicked() {
		err := clipboard.WriteAll(ow.Text())
		if err != nil {
			log.Errorf("copy error: %s", err)
		}
		consoleMI = false
	}

	for consoleDumpBtn.Clicked() {
		writeHistory()
		consoleMI = false
	}

	for consoleCancelBtn.Clicked() {
		consoleMI = false
	}

	for logLvlDebugBtn.Clicked() {
		setLevel(log.DebugLevel)
		logMI = false
	}

	for logLvlInfoBtn.Clicked() {
		setLevel(log.InfoLevel)
		logMI = false
	}

	for logLvlWarningBtn.Clicked() {
		setLevel(log.WarnLevel)
		logMI = false
	}

	for logLvlErrorBtn.Clicked() {
		setLevel(log.ErrorLevel)
		logMI = false
	}

	if fileMI {
		widget := l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx,
				xmat.RigidButton(th, "Open", &fileOpenBtn),
				xmat.RigidButton(th, "Save", &fileSaveBtn),
				xmat.RigidButton(th, "Provision", &fileProvisionBtn),
			)
		})

		return widget, true
	}

	if consoleMI {
		widget := l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx,
				xmat.RigidButton(th, "Clear", &consoleClearBtn),
				xmat.RigidButton(th, "Copy", &consoleClearBtn),
				xmat.RigidButton(th, "Dump history", &consoleDumpBtn),
				xmat.RigidButton(th, "Cancel", &consoleCancelBtn),
			)
		})

		return widget, true
	}

	if logMI {
		widget := l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx,
				xmat.RigidButton(th, "Debug", &logLvlDebugBtn),
				xmat.RigidButton(th, "Info", &logLvlInfoBtn),
				xmat.RigidButton(th, "Warning", &logLvlWarningBtn),
				xmat.RigidButton(th, "Error", &logLvlErrorBtn),
			)
		})

		return widget, true
	}

	widget := l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Horizontal}.Layout(gtx,
			xmat.RigidButton(th, "File", &fileMIBtn),
			xmat.RigidButton(th, "Console", &consoleMIBtn),
			xmat.RigidButton(th, "Log", &logMIBtn),
		)
	})

	return widget, false
}

func buildOpenFile(th *material.Theme) (l.FlexChild, bool) {

	*confFile = openFileCombo.SelectedText()

	for openFileCancelBtn.Clicked() {
		openFile = false
	}

	for openFileImportBtn.Clicked() {
		importConf()
		resetGuiValues()
		setDevice()
		openFile = false
	}

	widgets := l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			xmat.RigidSection(th, "Select file"),
			labelCombo(th, "Filename", &openFileCombo),
			xmat.RigidButton(th, "Cancel", &openFileCancelBtn),
			xmat.RigidButton(th, "Import", &openFileImportBtn),
		)
	})

	return widgets, true
}

func buildSaveFile(th *material.Theme) (l.FlexChild, bool) {

	saveFilename = saveFileEditor.Text()

	for saveFileCancelBtn.Clicked() {
		saveFile = false
	}

	for saveFileSaveBtn.Clicked() {
		exportConf(fmt.Sprintf("confs/%s", saveFilename))
		saveFile = false
	}

	widgets := l.Rigid(func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			xmat.RigidSection(th, "Save file"),
			xmat.RigidEditor(th, "Name", "<filename>", &saveFileEditor),
			xmat.RigidButton(th, "Cancel", &saveFileCancelBtn),
			xmat.RigidButton(th, "Save", &saveFileSaveBtn),
		)
	})

	return widgets, true
}

func resetGuiValues() {
	mqttResetGuiValue()
	forwarderResetGuiValues()
	loraResetGuiValues()
	deviceResetGuiValues()
	macResetGuiValues()
	dataResetGuiValues()
	provResetGuiValues()
}

var (
	serversButton widget.Clickable
	deviceButton  widget.Clickable
	loraButton    widget.Clickable
	controlButton widget.Clickable
	dataButton    widget.Clickable

	tabIndex uint
)

func mainWindow(gtx l.Context, th *material.Theme) {
	wOutputForm := outputForm(th)

	wMenu, isMenuOpen := buildMenu(th)

	if isMenuOpen {
		l.NW.Layout(gtx, func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx,
				wMenu,
				xmat.RigidSeparator(th, &giox.Separator{}),
				wOutputForm,
			)
		})
		return
	}

	for serversButton.Clicked() {
		tabIndex = 0
	}

	for deviceButton.Clicked() {
		tabIndex = 1
	}

	for loraButton.Clicked() {
		tabIndex = 2
	}

	for controlButton.Clicked() {
		tabIndex = 3
	}

	for dataButton.Clicked() {
		tabIndex = 4
	}

	tabsWidget := l.Rigid(func(gtx l.Context) l.Dimensions {
		gtx.Constraints = l.Exact(image.Point{X: 100, Y: 500})
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			xmat.RigidButton(th, "Connect", &serversButton),
			xmat.RigidButton(th, "Device", &deviceButton),
			xmat.RigidButton(th, "LoRa", &loraButton),
			xmat.RigidButton(th, "Control", &controlButton),
			xmat.RigidButton(th, "Data", &dataButton),
		)
	})

	wMqttForm := mqttForm(th)
	wForwarderForm := forwarderForm(th)
	wDeviceForm := deviceForm(th)
	wLoraForm := loRaForm(th)
	wControlForm := controlForm(th)
	wDataForm := dataForm(th)

	var selectedWidget l.FlexChild
	switch tabIndex {
	case 0:
		selectedWidget = l.Rigid(func(gtx l.Context) l.Dimensions {
			return l.Flex{Axis: l.Vertical}.Layout(gtx,
				wMqttForm,
				xmat.RigidSeparator(th, &giox.Separator{}),
				wForwarderForm,
			)
		})
	case 1:
		selectedWidget = wDeviceForm
	case 2:
		selectedWidget = wLoraForm
	case 3:
		selectedWidget = wControlForm
	case 4:
		selectedWidget = wDataForm
	}

	l.NW.Layout(gtx, func(gtx l.Context) l.Dimensions {
		return l.Flex{Axis: l.Vertical}.Layout(gtx,
			wMenu,
			xmat.RigidSeparator(th, &giox.Separator{}),
			l.Rigid(func(gtx l.Context) l.Dimensions {
				return l.Flex{Axis: l.Horizontal}.Layout(gtx,
					tabsWidget,
					selectedWidget,
				)
			}),
			xmat.RigidSeparator(th, &giox.Separator{}),
			wOutputForm,
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
	/* runtime.LockOSThread() */

	mw := io.MultiWriter(ow, os.Stderr)
	log.SetOutput(mw)

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()

	createLoRaForm()
	createDeviceForm()
	createDataForm()
	createOutputForm()
	tabIndex = 0

	importConf()
	resetGuiValues()
	setDevice()

	go func() {
		defer os.Exit(0)
		w := app.NewWindow(app.Size(unit.Px(1024), unit.Px(1024)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}
