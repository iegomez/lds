package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/atotto/clipboard"
	"github.com/brocaar/lorawan"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/iegomez/lds/lds"
	"github.com/inkyblackness/imgui-go"

	log "github.com/sirupsen/logrus"
)

type redisConf struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type windowConf struct {
	Width  int `toml:"width"`
	Height int `toml:"height"`
}

type tomlConfig struct {
	MQTT        mqtt           `toml:"mqtt"`
	Forwarder   forwarder      `toml:"forwarder"`
	Band        band           `toml:"band"`
	Device      device         `toml:"device"`
	GW          gateway        `toml:"gateway"`
	DR          dataRate       `toml:"data_rate"`
	RXInfo      rxInfo         `toml:"rx_info"`
	RawPayload  rawPayload     `toml:"raw_payload"`
	EncodedType []*encodedType `toml:"encoded_type"`
	LogLevel    string         `toml:"log_level"`
	RedisConf   redisConf      `toml:"redis"`
	Window      windowConf     `toml:"window"`
	Provisioner provisioner    `toml:"provisioner"`
}

// Configuration holders.
var (
	confFile *string
	config   *tomlConfig
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

// Configuration files loading and saving.
var (
	openFile     bool
	files        []os.FileInfo
	saveFile     bool
	saveFilename string
	mwOpen       = true
)

func importConf() {

	//When config hasn't been initialized we need to provide fresh zero instances with some defaults.
	//Decoding the conf file will override any present option.
	if config == nil {
		cMqtt := mqtt{}

		cForwarder := forwarder{}

		cGw := gateway{}

		cDev := device{
			MType: lorawan.UnconfirmedDataUp,
		}

		cBand := band{}

		cDr := dataRate{}

		cRx := rxInfo{}

		cPl := rawPayload{
			MaxExecTime: 100,
		}

		et := []*encodedType{}

		w := windowConf{
			Width:  1200,
			Height: 1000,
		}

		p := provisioner{}

		config = &tomlConfig{
			MQTT:        cMqtt,
			Forwarder:   cForwarder,
			Band:        cBand,
			Device:      cDev,
			GW:          cGw,
			DR:          cDr,
			RXInfo:      cRx,
			RawPayload:  cPl,
			EncodedType: et,
			Window:      w,
			Provisioner: p,
		}
	}

	if _, err := toml.DecodeFile(*confFile, &config); err != nil {
		log.Println(err)
		return
	}

	l, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(l)
	}

	//Try to set redis.
	lds.StartRedis(config.RedisConf.Addr, config.RedisConf.Password, config.RedisConf.DB)

	for i := 0; i < len(config.EncodedType); i++ {
		config.EncodedType[i].ValueS = strconv.FormatFloat(config.EncodedType[i].Value, 'f', -1, 64)
		config.EncodedType[i].MaxValueS = strconv.FormatFloat(config.EncodedType[i].MaxValue, 'f', -1, 64)
		config.EncodedType[i].MinValueS = strconv.FormatFloat(config.EncodedType[i].MinValue, 'f', -1, 64)
		config.EncodedType[i].NumBytesS = strconv.Itoa(config.EncodedType[i].NumBytes)
	}

	//Fill string representations of numeric values.
	config.DR.BitRateS = strconv.Itoa(config.DR.BitRate)
	config.RXInfo.ChannelS = strconv.Itoa(config.RXInfo.Channel)
	config.RXInfo.CrcStatusS = strconv.Itoa(config.RXInfo.CrcStatus)
	config.RXInfo.FrequencyS = strconv.Itoa(config.RXInfo.Frequency)
	config.RXInfo.LoRASNRS = strconv.FormatFloat(config.RXInfo.LoRaSNR, 'f', -1, 64)
	config.RXInfo.RfChainS = strconv.Itoa(config.RXInfo.RfChain)
	config.RXInfo.RssiS = strconv.Itoa(config.RXInfo.Rssi)

	//Set default script when it's not present.
	if config.RawPayload.Script == "" {
		config.RawPayload.Script = defaultScript
	}
	config.RawPayload.FPortS = strconv.Itoa(config.RawPayload.FPort)

	//Set the device with the given options.
	setDevice()
}

func exportConf(filename string) {
	if !strings.Contains(filename, ".toml") {
		filename = fmt.Sprintf("%s.toml", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	encoder := toml.NewEncoder(f)
	err = encoder.Encode(config)
	if err != nil {
		log.Errorf("export error: %s", err)
		return
	}
	log.Infof("exported conf file %s", f.Name())
	*confFile = f.Name()

}

func beginMenu() {
	if imgui.BeginMainMenuBar() {
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
	}
}

func beginOpenFile() {
	if openFile {
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
			imgui.CloseCurrentPopup()
			//Close popup.
		}
		imgui.EndPopup()
	}
}

func beginSaveFile() {
	if saveFile {
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
	}
}

func main() {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	log.SetOutput(ow)

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)

	confFile = flag.String("conf", "conf.toml", "path to toml configuration file")
	flag.Parse()

	importConf()

	window, err := glfw.CreateWindow(config.Window.Width, config.Window.Height, "LoRaServer device simulator", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	err = gl.Init()
	if err != nil {
		panic(err)
	}

	context := imgui.CreateContext(nil)
	defer context.Destroy()

	impl := imguiGlfw3Init(window)
	defer impl.Shutdown()

	var clearColor imgui.Vec4

	for !window.ShouldClose() {
		glfw.PollEvents()
		impl.NewFrame()

		beginMQTTForm()
		beginForwarderForm()
		beginDeviceForm()
		beginLoRaForm()
		beginControl()
		beginDataForm()
		beginOutput()
		beginMenu()
		beginProvisioner()

		displayWidth, displayHeight := window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(displayWidth), int32(displayHeight))
		gl.ClearColor(clearColor.X, clearColor.Y, clearColor.Z, clearColor.W)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		imgui.Render()
		impl.Render(imgui.RenderedDrawData())

		window.SwapBuffers()
		<-time.After(time.Millisecond * 25)
	}
}
