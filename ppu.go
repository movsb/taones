package main

import (
	"image"
)

type PPUCTRL struct {
	flagNameTable       byte // 0: $2000, 1: $2400, 2: $2800, 3: $2C00
	flagIncrement       byte // 0: +1, 1: +32
	flagSpriteTable     byte // 0: $0000, 1: $1000
	flagBackgroundTable byte // 0: $0000, 1: $1000
	flagSpriteSize      byte // 0: 8x8, 1: 8x16
	flagVBlank          byte // 1: 在 vblank 时中断
}

type PPUMASK struct {
	flagGrayscale          byte // 0: color, 1: grayscale
	flagShowLeftBackground byte // 0: hide, 1: show
	flagShowLeftSprites    byte // 0: hide, 1: show
	flagShowBackground     byte // 0:hide, 1: show
	flagShowSprites        byte // 0: hide, 1: show
}

type PPU struct {
	MemoryReadWriter
	console *Console

	Cycle      int // 0-340
	Scanline   int //0-261
	FrameCount uint64

	palette   [32]byte
	nameTable [2048]byte
	sprite    [256]byte
	front     *image.RGBA
	back      *image.RGBA

	// NMI flags
	nmiOccurred bool
	nmiOutput   bool
	nmiPrevious bool
	nmiDelay    byte

	PPUCTRL

	PPUMASK

	// $2003
	spriteAddr byte // dma sprites

	// $2007 PPUDATA
}

func NewPPU(console *Console) *PPU {
	ppu := PPU{}
	ppu.front = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.back = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.Reset()
	return &ppu
}
