package main

const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

type ControllerProvider interface {
	Flush(frameCounter uint64)
	Read() byte
}

type EmptyController struct {
}

func (o *EmptyController) Flush(frameCounter uint64) {

}

func (o *EmptyController) Read() byte {
	return 0
}

type KeyboardController struct {
	buttons [8]bool
	index   byte
	flusher func(frameCounter uint64) [8]bool
}

func NewKeyboardController(flusher func(frameCounter uint64) [8]bool) ControllerProvider {
	return &KeyboardController{
		flusher: flusher,
	}
}

func (o *KeyboardController) Flush(frameCounter uint64) {
	if o.flusher != nil {
		o.buttons = o.flusher(frameCounter)
	}
	o.index = 0
}

func (o *KeyboardController) Read() byte {
	var d byte
	if o.index < 8 && o.buttons[o.index] {
		d = 1
	}
	o.index++
	return d
}
