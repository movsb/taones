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
	Flush()
	Read() byte
}

type EmptyController struct {
}

func (o *EmptyController) Flush() {

}

func (o *EmptyController) Read() byte {
	return 0
}

type KeyboardController struct {
	buttons [8]bool
	index   byte
	flusher func() [8]bool
}

func NewKeyboardController(flusher func() [8]bool) ControllerProvider {
	return &KeyboardController{
		flusher: flusher,
	}
}

func (o *KeyboardController) Flush() {
	if o.flusher != nil {
		o.buttons = o.flusher()
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
