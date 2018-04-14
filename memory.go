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
	case a >= 0x6000:
		o.console.mapper.Write(a, v)
	default:
		log.Fatalf("unhandled cpu memory write at 0x%04X\n", a)
	}
}
