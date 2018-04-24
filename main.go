package main

import (
	"flag"
	"image"
	"log"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var config struct {
	opcodes bool
}

func init() {
	// we need a parallel OS thread to avoid audio stuttering
	runtime.GOMAXPROCS(2)

	// we need to keep OpenGL calls on a single thread
	runtime.LockOSThread()
}

func createTexture() uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return texture
}

func setTexture(im *image.RGBA) {
	size := im.Rect.Size()
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA, int32(size.X), int32(size.Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / 256
	s2 := float32(h) / 240
	f := float32(1 - 0)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func main() {
	flag.BoolVar(&config.opcodes, "opcodes", false, "show opcodes")
	flag.Parse()

	var err error
	if err = glfw.Init(); err != nil {
		panic(err)
	}

	defer glfw.Terminate()

	window, err := glfw.CreateWindow(256, 240, "taones", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	var focus bool

	window.SetFocusCallback(func(w *glfw.Window, f bool) {
		focus = f
	})

	// initialize gl
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	gl.Enable(gl.TEXTURE_2D)

	var lastTime float64

	console := NewConsole()
	cartridge := LoadROM("smb.nes")
	console.Run(cartridge)

	kbdCtrl1 := NewKeyboardController(func(frameCounter uint64) [8]bool {
		var keys [8]bool

		read := func(index byte, key glfw.Key) {
			keys[index] = keys[index] || window.GetKey(key) == glfw.Press
		}

		read(ButtonA, glfw.KeyK)
		read(ButtonB, glfw.KeyJ)
		if frameCounter&3 == 0 {
			read(ButtonA, glfw.KeyI)
			read(ButtonB, glfw.KeyU)
		}
		read(ButtonSelect, glfw.KeyT)
		read(ButtonStart, glfw.KeyY)
		read(ButtonUp, glfw.KeyW)
		read(ButtonDown, glfw.KeyS)
		read(ButtonLeft, glfw.KeyA)
		read(ButtonRight, glfw.KeyD)

		return keys
	})

	console.SetController1(kbdCtrl1)

	texture := createTexture()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		currTime := glfw.GetTime()
		dt := currTime - lastTime
		lastTime = currTime
		if dt > 1 {
			dt = 0
		}

		if focus {
			console.StepSeconds(dt)
		}

		gl.BindTexture(gl.TEXTURE_2D, texture)

		setTexture(console.ppu.front)
		drawBuffer(window)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		window.SwapBuffers()
		glfw.PollEvents()
	}

}
