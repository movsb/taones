package main

type Console struct {
	cpu    *CPU
	ppu    *PPU
	apu    *APU
	cart   *Cartridge
	mapper Mapper
	ctrl1  ControllerProvider
}

func NewConsole() *Console {
	console := &Console{}
	console.cpu = NewCPU(console)
	console.ppu = NewPPU(console)
	console.apu = NewAPU(console)
	console.ctrl1 = &EmptyController{}
	return console
}

func (o *Console) SetController1(controller ControllerProvider) {
	o.ctrl1 = controller
}

func (o *Console) Step() int {
	cpuCycles := o.cpu.Step()
	ppuCycles := cpuCycles * 3
	for ; ppuCycles > 0; ppuCycles-- {
		o.ppu.Step()
	}
	for apuCycles := cpuCycles; apuCycles > 0; apuCycles-- {
		o.apu.Step()
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
	o.apu.Reset()
}

func (o *Console) Run(cart *Cartridge) {
	o.cart = cart
	o.mapper = NewMapper(o, cart)

	o.Reset()
}
