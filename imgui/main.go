package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
)

func main() {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)

	window, err := glfw.CreateWindow(1280, 720, "ImGui-Go GLFW+OpenGL3 example", nil, nil)
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

	/*
		imgui.CurrentStyle().ScaleAllSizes(2.0)
		imgui.CurrentIO().SetFontGlobalScale(2.0)
	*/

	impl := imguiGlfw3Init(window)
	defer impl.Shutdown()

	showDemoWindow := false
	showAnotherWindow := false
	counter := 0
	var clearColor imgui.Vec4

	var str1 = "str1"
	var str2 = "str2"
	var str3 = "str3"
	var str4 = "str4"

	for !window.ShouldClose() {
		glfw.PollEvents()
		impl.NewFrame()

		// 1. Show a simple window.
		// Tip: if we don't call ImGui::Begin()/ImGui::End() the widgets automatically appears in a window called "Debug".
		{
			imgui.Text("Hello, world!")

			imgui.Checkbox("Demo Window", &showDemoWindow)
			imgui.Checkbox("Another Window", &showAnotherWindow)

			if imgui.Button("Button") {
				counter++
			}
			imgui.SameLine()
			imgui.Text(fmt.Sprintf("counter = %d", counter))
			imgui.InputTextMultilineV("line 1", &str1, imgui.Vec2{X: 0, Y: 0}, imgui.InputTextFlagsCallbackCharFilter, func(data imgui.InputTextCallbackData) int32 {
				fmt.Println(data.Buffer())
				if len(str1) >= 6 {
					return 1
				}
				return 0
			})
			imgui.Text(str1)
			imgui.InputText("line 2", &str2)
			imgui.Text(str2)
			imgui.InputTextV("line 3", &str3, imgui.InputTextFlagsCallbackAlways, func(data imgui.InputTextCallbackData) int32 {
				buf := data.Buffer()
				fmt.Printf("buff len: %d - char: %s - key: %d - flags: %d\n", len(string(data.Buffer())), string(data.EventChar()), data.EventKey(), data.EventFlag())
				if len(string(buf)) > 8 {
					data.DeleteBytes(4, len(data.Buffer())-4)
					data.InsertBytes(4, []byte(" m"))
					//data.MarkBufferModified()
					return 1
				}
				return 0
			})
			imgui.Text(str3)
			imgui.InputTextMultiline("line 4", &str4)
			imgui.Text(str4)
			//label string, buf *string, bufSize int64, flags int

		}

		// 2. Show another simple window. In most cases you will use an explicit Begin/End pair to name your windows.
		if showAnotherWindow {
			imgui.BeginV("Another Window", &showAnotherWindow, 0)
			imgui.Text("Hello from another window!")
			if imgui.Button("Close Me") {
				showAnotherWindow = false
			}

			imgui.End()

		}

		// 3. Show the ImGui demo window. Most of the sample code is in imgui.ShowDemoWindow().
		// Read its code to learn more about Dear ImGui!
		if showDemoWindow {
			imgui.ShowDemoWindow(&showDemoWindow)
		}

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

func cb(data imgui.InputTextCallbackData) int32 {
	fmt.Printf("buff len: %d - char: %s - key: %d - flags: %d\n", len(string(data.Buffer())), string(data.EventChar()), data.EventKey(), data.EventFlag())
	if len(string(data.Buffer())) > 8 {
		data.SetEventChar(0)
		data.MarkBufferModified()
		return 1
	}
	return 0
}
