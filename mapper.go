package main

type Mapper interface {
	Read(a uint16) byte
	Write(a uint16, v byte)
	Step()
}

func NewMapper(console *Console, cart *Cartridge) Mapper {
	switch cart.Mapper {
	case 0:
		return &xMapper0{
			console,
			cart,
		}
	}
	return nil
}

// Mapper 0
type xMapper0 struct {
	console *Console
	*Cartridge
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
