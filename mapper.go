// https://wiki.nesdev.com/w/index.php/Mapper

package main

import (
	"log"
)

type Mapper interface {
	Read(a uint16) byte
	Write(a uint16, v byte)
	Step()
}

func NewMapper(console *Console, cart *Cartridge) Mapper {
	switch cart.Mapper {
	case 0:
		return NewMapper0(console, cart)
	case 2:
		return NewMapper2(console, cart)
	}
	log.Fatalf("unsupported mapper: %d\n", cart.Mapper)
	return nil
}

// NROM (Mapper 0)
type xMapper0 struct {
	console *Console
	*Cartridge
}

func NewMapper0(console *Console, cart *Cartridge) Mapper {
	return &xMapper0{
		console,
		cart,
	}
}

func (o *xMapper0) Read(a uint16) byte {
	switch {
	case a < 0x2000:
		return o.CHR[a]
	case a >= 0x8000:
		return o.PRG[a-0x8000]
	}
	return 0
}

func (o *xMapper0) Write(a uint16, v byte) {

}

func (o *xMapper0) Step() {

}

// UxROM (Mapper 2)
type xMapper2 struct {
	console *Console
	cart    *Cartridge
	banks   [][]byte // $8000-$BFFF
	last    []byte   // $C000-$FFFF
	nprg    int
	bank    byte
}

func NewMapper2(console *Console, cart *Cartridge) Mapper {
	m := &xMapper2{}
	m.console = console
	m.cart = cart

	m.nprg = len(cart.PRG) / 16384
	m.last = cart.PRG[(m.nprg-1)*16384:]
	m.banks = make([][]byte, m.nprg)
	for i := 0; i < m.nprg; i++ {
		m.banks[i] = cart.PRG[i*16384 : (i+1)*16384]
	}

	return m
}

func (o *xMapper2) Read(a uint16) byte {
	switch {
	case a < 0x2000:
		return o.cart.CHR[a]
	case a >= 0x8000 && a <= 0xBFFF:
		return o.banks[o.bank][a-0x8000]
	case a >= 0xC000:
		return o.last[a-0xC000]
	default:
		log.Fatalf("未知读内存：%04X", a)
		return 0
	}
}

func (o *xMapper2) Write(a uint16, v byte) {
	switch {
	case a < 0x2000:
		o.cart.CHR[a] = v
	case a >= 0x8000 && a <= 0xFFFF:
		o.bank = v % byte(o.nprg)
	default:
		log.Fatalf("未知写内存：%04X = %d\n", a, v)
	}
}

func (o *xMapper2) Step() {

}
