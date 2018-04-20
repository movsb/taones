package main

import (
	"log"
)

type MemoryReadWriter interface {
	Read(a uint16) byte
	Write(a uint16, v byte)
}

type CPUMemory struct {
	console *Console
}

func NewCPUMemory(console *Console) MemoryReadWriter {
	return &CPUMemory{console: console}
}

func (o *CPUMemory) Read(a uint16) byte {
	switch {
	case a < 0x2000:
		return o.console.cpu.RAM[a&0x0800]
	case a < 0x4000:
		return o.console.ppu.readRegister(0x2000 + a&7)
	case a == 0x4014:
		return o.console.ppu.readRegister(a)
	case a >= 0x6000:
		return o.console.mapper.Read(a)
	default:
		log.Fatalf("unhandled cpu memory read at 0x%04X\n", a)
	}
	return 0
}

func (o *CPUMemory) Write(a uint16, v byte) {
	switch {
	case a < 0x2000:
		o.console.cpu.RAM[a&0x0800] = v
	case a < 0x4000:
		o.console.ppu.writeRegister(0x2000+a&7, v)
	case a == 0x4014:
		o.console.ppu.writeRegister(a, v)
	case a >= 0x6000:
		o.console.mapper.Write(a, v)
	default:
		log.Fatalf("unhandled cpu memory write at 0x%04X\n", a)
	}
}

type PPUMemory struct {
	console *Console
}

func NewPPUMemory(console *Console) MemoryReadWriter {
	return &PPUMemory{console: console}
}

func (o *PPUMemory) Read(a uint16) byte {
	a &= 0x3FFF
	switch {
	// 图案表
	case a < 0x2000:
		return o.console.mapper.Read(a)
	// 命名表 & 属性表
	case a < 0x3F00:
		mirror := o.console.cart.Mirror
		return o.console.ppu.nameTable[mirrorAddress(mirror, a)%2048]
	// 调色板
	case a < 0x4000:
		return o.console.ppu.readPalette(a % 32)
	default:
		log.Fatalf("unhandled ppu memory read at: 0x%04X\n", a)
	}
	return 0
}

func (o *PPUMemory) Write(a uint16, v byte) {
	a = a % 0x4000
	switch {
	case a < 0x2000:
		o.console.mapper.Write(a, v)
	case a < 0x3F00:
		mode := o.console.cart.Mirror
		o.console.ppu.nameTable[mirrorAddress(mode, a)%2048] = v
	case a < 0x4000:
		o.console.ppu.writePalette(a%32, v)
	default:
		log.Fatalf("unhandled ppu memory write at address: 0x%04X", a)
	}
}

const (
	mirrorHorizontal = 0
	mirrorVertical   = 1
	mirrorSingle0    = 2
	mirrorSingle1    = 3
	mirrorFour       = 4
)

var mirrorLookup = [...][4]uint16{
	{0, 0, 1, 1},
	{0, 1, 0, 1},
	{0, 0, 0, 0},
	{1, 1, 1, 1},
	{0, 1, 2, 3},
}

func mirrorAddress(mode byte, a uint16) uint16 {
	a = (a - 0x2000) % 0x1000
	table := a / 0x0400
	offset := a % 0x0400
	return 0x2000 + mirrorLookup[mode][table]*0x0400 + offset
}
