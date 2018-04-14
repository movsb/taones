package main

type Console struct {
	cpu    *CPU
	cart   *Cartridge
	mapper Mapper
}

func NewConsole() *Console {
	console := &Console{}
	console.cpu = NewCPU(console)
	return console
}

func (o *Console) Step() int {
	cpuCycles := o.cpu.Step()
	return cpuCycles
}

func (o *Console) Reset() {
	o.cpu.Reset()
}

func (o *Console) Run(cart *Cartridge) {
	o.cart = cart
	o.mapper = NewMapper(o, cart)

	o.Reset()

	for {
		o.Step()
	}
}
