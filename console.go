package main

type Console struct {
	cpu    *CPU
	ppu    *PPU
	cart   *Cartridge
	mapper Mapper
}

func NewConsole() *Console {
	console := &Console{}
	console.cpu = NewCPU(console)
	console.ppu = NewPPU(console)
	return console
}

func (o *Console) Step() int {
	cpuCycles := o.cpu.Step()
	ppuCycles := cpuCycles * 3
	for ; ppuCycles > 0; ppuCycles-- {
		o.ppu.Step()
	}
	return cpuCycles
}

func (o *Console) StepSeconds(s float64) {
	cycles := int(cpuFreq * s)
	for cycles > 0 {
		cycles -= o.Step()
	}
}

func (o *Console) Reset() {
	o.cpu.Reset()
}

func (o *Console) Run(cart *Cartridge) {
	o.cart = cart
	o.mapper = NewMapper(o, cart)

	o.Reset()
}
