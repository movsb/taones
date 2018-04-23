package main

import (
	"flag"

	"github.com/veandco/go-sdl2/sdl"
)

var config struct {
	opcodes bool
	scale   uint
}

func main() {
	flag.BoolVar(&config.opcodes, "opcodes", false, "show opcodes")
	flag.UintVar(&config.scale, "scale", 2, "video scaler")
	flag.Parse()

	var err error
	_ = err

	console := NewConsole()
	cartridge := LoadROM("smb.nes")
	console.Run(cartridge)

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	defer sdl.Quit()

	window, err := sdl.CreateWindow("taones",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		256*int32(config.scale), 240*int32(config.scale), sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}

	defer window.Destroy()

	wid, _ := window.GetID()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	buffer, err := sdl.CreateRGBSurface(0, 256, 240, 32, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}

	bufPixels := buffer.Pixels()
	console.ppu.SetBuffer(bufPixels)

	var keys [8]bool
	var turboA, turboB bool

	kbdCtrl1 := NewKeyboardController(func(frameCounter uint64) [8]bool {
		var keys2 = keys

		if frameCounter&3 == 0 {
			keys2[ButtonA] = keys2[ButtonA] || turboA
			keys2[ButtonB] = keys2[ButtonB] || turboB
		}

		return keys2
	})

	console.SetController1(kbdCtrl1)

	var lastTime uint32

	var originRect = &sdl.Rect{0, 0, 256, 240}
	var scaledRect = &sdl.Rect{0, 0, 256 * int32(config.scale), 240 * int32(config.scale)}

	for run := true; run; {
		switch evt := sdl.PollEvent().(type) {
		case *sdl.KeyboardEvent:
			if evt.WindowID == wid {
				switch evt.Keysym.Sym {
				case sdl.K_w:
					keys[ButtonUp] = evt.Type == sdl.KEYDOWN
				case sdl.K_s:
					keys[ButtonDown] = evt.Type == sdl.KEYDOWN
				case sdl.K_a:
					keys[ButtonLeft] = evt.Type == sdl.KEYDOWN
				case sdl.K_d:
					keys[ButtonRight] = evt.Type == sdl.KEYDOWN
				case sdl.K_t:
					keys[ButtonSelect] = evt.Type == sdl.KEYDOWN
				case sdl.K_y:
					keys[ButtonStart] = evt.Type == sdl.KEYDOWN
				case sdl.K_j:
					keys[ButtonB] = evt.Type == sdl.KEYDOWN
				case sdl.K_k:
					keys[ButtonA] = evt.Type == sdl.KEYDOWN
				case sdl.K_u:
					turboB = evt.Type == sdl.KEYDOWN
				case sdl.K_i:
					turboA = evt.Type == sdl.KEYDOWN
				}
			}
		case *sdl.QuitEvent:
			run = false
		}

		ticks := sdl.GetTicks()
		diff := ticks - lastTime
		if diff > 1000 {
			diff = 0
		}
		lastTime = ticks

		console.StepSeconds(float64(diff) / 1000)

		buffer.BlitScaled(originRect, surface, scaledRect)

		window.UpdateSurface()
	}
}
