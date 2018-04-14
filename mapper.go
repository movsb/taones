package main

type Mapper interface {
	Read(a uint16) byte
	Write(a uint16, v byte)
	Step()
}

func NewMapper(console *Console, cart *Cartridge) Mapper {
	switch cart.Mapper {
	case 0:
		return NewMapper2(cart)
	default:
		panic("bad cart")
	}
}
