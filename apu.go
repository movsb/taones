package main

import "math"

const (
	audioSampleRate = 44100
	audioAmplitude  = 24000
)

var lengthTable = [32]byte{
	10, 254, 20, 2, 40, 4, 80, 6,
	160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22,
	192, 24, 72, 26, 16, 28, 32, 30,
}

var pulseDutyTable = [4][8]byte{
	{0, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0},
	{1, 0, 0, 1, 1, 1, 1, 1},
}

var triangleTable = [32]byte{
	15, 14, 13, 12, 11, 10, 9, 8,
	7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15,
}

var noisePeriodTable = [16]uint16{
	4, 8, 16, 32, 64, 96, 128, 160,
	202, 254, 380, 508, 762, 1016, 2034, 4068,
}

type APU struct {
	console *Console

	pulse1   pulse
	pulse2   pulse
	triangle triangle
	noise    noise

	cycle           uint64
	frameCycle      int
	frameMode5      bool
	sampleRemainder int
	samples         []byte
}

func NewAPU(console *Console) *APU {
	apu := &APU{console: console}
	apu.Reset()
	return apu
}

func (o *APU) Reset() {
	o.pulse1 = pulse{channel: 1}
	o.pulse2 = pulse{channel: 2}
	o.triangle = triangle{}
	o.noise = noise{shift: 1}
	o.cycle = 0
	o.frameCycle = 0
	o.frameMode5 = false
	o.sampleRemainder = 0
	o.samples = o.samples[:0]
}

func (o *APU) Step() {
	o.cycle++
	o.clockFrameCounter()

	o.triangle.stepTimer()
	if o.cycle&1 == 0 {
		o.pulse1.stepTimer()
		o.pulse2.stepTimer()
		o.noise.stepTimer()
	}

	o.sampleRemainder += audioSampleRate
	for o.sampleRemainder >= cpuFreq {
		o.sampleRemainder -= cpuFreq
		o.appendSample(o.output())
	}
}

func (o *APU) ReadRegister(a uint16) byte {
	switch a {
	case 0x4015:
		var v byte
		if o.pulse1.lengthCounter > 0 {
			v |= 1 << 0
		}
		if o.pulse2.lengthCounter > 0 {
			v |= 1 << 1
		}
		if o.triangle.lengthCounter > 0 {
			v |= 1 << 2
		}
		if o.noise.lengthCounter > 0 {
			v |= 1 << 3
		}
		return v
	default:
		return 0
	}
}

func (o *APU) WriteRegister(a uint16, v byte) {
	switch a {
	case 0x4000:
		o.pulse1.writeControl(v)
	case 0x4001:
		o.pulse1.writeSweep(v)
	case 0x4002:
		o.pulse1.writeTimerLow(v)
	case 0x4003:
		o.pulse1.writeTimerHigh(v)
	case 0x4004:
		o.pulse2.writeControl(v)
	case 0x4005:
		o.pulse2.writeSweep(v)
	case 0x4006:
		o.pulse2.writeTimerLow(v)
	case 0x4007:
		o.pulse2.writeTimerHigh(v)
	case 0x4008:
		o.triangle.writeControl(v)
	case 0x400A:
		o.triangle.writeTimerLow(v)
	case 0x400B:
		o.triangle.writeTimerHigh(v)
	case 0x400C:
		o.noise.writeControl(v)
	case 0x400E:
		o.noise.writePeriod(v)
	case 0x400F:
		o.noise.writeLength(v)
	case 0x4015:
		o.writeStatus(v)
	case 0x4017:
		o.writeFrameCounter(v)
	}
}

func (o *APU) TakeSamples() []byte {
	if len(o.samples) == 0 {
		return nil
	}
	samples := o.samples
	o.samples = nil
	return samples
}

func (o *APU) clockFrameCounter() {
	o.frameCycle++

	if o.frameMode5 {
		switch o.frameCycle {
		case 7457, 22371:
			o.clockQuarterFrame()
		case 14913, 37281:
			o.clockQuarterFrame()
			o.clockHalfFrame()
		case 37282:
			o.frameCycle = 0
		}
		return
	}

	switch o.frameCycle {
	case 7457, 22371:
		o.clockQuarterFrame()
	case 14913, 29829:
		o.clockQuarterFrame()
		o.clockHalfFrame()
	case 29830:
		o.frameCycle = 0
	}
}

func (o *APU) clockQuarterFrame() {
	o.pulse1.envelope.clock()
	o.pulse2.envelope.clock()
	o.triangle.clockLinearCounter()
	o.noise.envelope.clock()
}

func (o *APU) clockHalfFrame() {
	o.pulse1.clockLength()
	o.pulse2.clockLength()
	o.triangle.clockLength()
	o.noise.clockLength()
	o.pulse1.clockSweep()
	o.pulse2.clockSweep()
}

func (o *APU) writeStatus(v byte) {
	o.pulse1.enabled = v&0x01 != 0
	if !o.pulse1.enabled {
		o.pulse1.lengthCounter = 0
	}
	o.pulse2.enabled = v&0x02 != 0
	if !o.pulse2.enabled {
		o.pulse2.lengthCounter = 0
	}
	o.triangle.enabled = v&0x04 != 0
	if !o.triangle.enabled {
		o.triangle.lengthCounter = 0
	}
	o.noise.enabled = v&0x08 != 0
	if !o.noise.enabled {
		o.noise.lengthCounter = 0
	}
}

func (o *APU) writeFrameCounter(v byte) {
	o.frameMode5 = v&0x80 != 0
	o.frameCycle = 0
	if o.frameMode5 {
		o.clockQuarterFrame()
		o.clockHalfFrame()
	}
}

func (o *APU) output() float64 {
	p1 := float64(o.pulse1.output())
	p2 := float64(o.pulse2.output())
	t := float64(o.triangle.output())
	n := float64(o.noise.output())

	var pulseOut float64
	if p1+p2 > 0 {
		pulseOut = 95.88 / ((8128.0 / (p1 + p2)) + 100)
	}

	var tndOut float64
	if t+n > 0 {
		tndOut = 159.79 / (1/(t/8227.0+n/12241.0) + 100)
	}

	return pulseOut + tndOut
}

func (o *APU) appendSample(v float64) {
	s := int16(math.Round(v * audioAmplitude))
	o.samples = append(o.samples, byte(s), byte(s>>8))
}

type envelope struct {
	loop     bool
	constant bool
	period   byte

	start   bool
	divider byte
	decay   byte
}

func (o *envelope) clock() {
	if o.start {
		o.start = false
		o.decay = 15
		o.divider = o.period
		return
	}

	if o.divider > 0 {
		o.divider--
		return
	}

	o.divider = o.period
	if o.decay > 0 {
		o.decay--
	} else if o.loop {
		o.decay = 15
	}
}

func (o *envelope) output() byte {
	if o.constant {
		return o.period
	}
	return o.decay
}

type pulse struct {
	channel int
	enabled bool

	duty          byte
	sequence      byte
	timer         uint16
	timerCounter  uint16
	lengthHalt    bool
	lengthCounter byte

	envelope envelope

	sweepEnabled bool
	sweepPeriod  byte
	sweepNegate  bool
	sweepShift   byte
	sweepReload  bool
	sweepDivider byte
}

func (o *pulse) writeControl(v byte) {
	o.duty = v >> 6
	o.lengthHalt = v&0x20 != 0
	o.envelope.loop = o.lengthHalt
	o.envelope.constant = v&0x10 != 0
	o.envelope.period = v & 0x0F
}

func (o *pulse) writeSweep(v byte) {
	o.sweepEnabled = v&0x80 != 0
	o.sweepPeriod = (v >> 4) & 7
	o.sweepNegate = v&0x08 != 0
	o.sweepShift = v & 7
	o.sweepReload = true
}

func (o *pulse) writeTimerLow(v byte) {
	o.timer = o.timer&0xFF00 | uint16(v)
}

func (o *pulse) writeTimerHigh(v byte) {
	o.timer = o.timer&0x00FF | uint16(v&7)<<8
	o.sequence = 0
	o.envelope.start = true
	if o.enabled {
		o.lengthCounter = lengthTable[v>>3]
	}
}

func (o *pulse) stepTimer() {
	if o.timerCounter == 0 {
		o.timerCounter = o.timer
		o.sequence = (o.sequence + 1) & 7
	} else {
		o.timerCounter--
	}
}

func (o *pulse) clockLength() {
	if !o.lengthHalt && o.lengthCounter > 0 {
		o.lengthCounter--
	}
}

func (o *pulse) clockSweep() {
	muted := o.muted()
	if o.sweepDivider == 0 {
		if o.sweepEnabled && o.sweepShift > 0 && !muted {
			o.timer = o.sweepTarget()
		}
		o.sweepDivider = o.sweepPeriod
		o.sweepReload = false
	} else if o.sweepReload {
		o.sweepDivider = o.sweepPeriod
		o.sweepReload = false
	} else {
		o.sweepDivider--
	}
}

func (o *pulse) sweepTarget() uint16 {
	change := o.timer >> o.sweepShift
	if !o.sweepNegate {
		return o.timer + change
	}
	if o.channel == 1 {
		return o.timer - change - 1
	}
	return o.timer - change
}

func (o *pulse) muted() bool {
	return o.timer < 8 || (o.sweepShift > 0 && o.sweepTarget() > 0x07FF)
}

func (o *pulse) output() byte {
	if !o.enabled || o.lengthCounter == 0 || o.muted() {
		return 0
	}
	if pulseDutyTable[o.duty][o.sequence] == 0 {
		return 0
	}
	return o.envelope.output()
}

type triangle struct {
	enabled bool

	control           bool
	linearReloadValue byte
	linearReload      bool
	linearCounter     byte

	timer         uint16
	timerCounter  uint16
	sequence      byte
	lengthCounter byte
}

func (o *triangle) writeControl(v byte) {
	o.control = v&0x80 != 0
	o.linearReloadValue = v & 0x7F
}

func (o *triangle) writeTimerLow(v byte) {
	o.timer = o.timer&0xFF00 | uint16(v)
}

func (o *triangle) writeTimerHigh(v byte) {
	o.timer = o.timer&0x00FF | uint16(v&7)<<8
	o.linearReload = true
	if o.enabled {
		o.lengthCounter = lengthTable[v>>3]
	}
}

func (o *triangle) stepTimer() {
	if o.timerCounter == 0 {
		o.timerCounter = o.timer
		if o.lengthCounter > 0 && o.linearCounter > 0 {
			o.sequence = (o.sequence + 1) & 31
		}
	} else {
		o.timerCounter--
	}
}

func (o *triangle) clockLinearCounter() {
	if o.linearReload {
		o.linearCounter = o.linearReloadValue
	} else if o.linearCounter > 0 {
		o.linearCounter--
	}

	if !o.control {
		o.linearReload = false
	}
}

func (o *triangle) clockLength() {
	if !o.control && o.lengthCounter > 0 {
		o.lengthCounter--
	}
}

func (o *triangle) output() byte {
	if !o.enabled || o.lengthCounter == 0 || o.linearCounter == 0 {
		return 0
	}
	return triangleTable[o.sequence]
}

type noise struct {
	enabled bool

	mode          bool
	timer         uint16
	timerCounter  uint16
	shift         uint16
	lengthHalt    bool
	lengthCounter byte

	envelope envelope
}

func (o *noise) writeControl(v byte) {
	o.lengthHalt = v&0x20 != 0
	o.envelope.loop = o.lengthHalt
	o.envelope.constant = v&0x10 != 0
	o.envelope.period = v & 0x0F
}

func (o *noise) writePeriod(v byte) {
	o.mode = v&0x80 != 0
	o.timer = noisePeriodTable[v&0x0F]
}

func (o *noise) writeLength(v byte) {
	o.envelope.start = true
	if o.enabled {
		o.lengthCounter = lengthTable[v>>3]
	}
}

func (o *noise) stepTimer() {
	if o.timerCounter == 0 {
		o.timerCounter = o.timer
		o.shiftRegister()
	} else {
		o.timerCounter--
	}
}

func (o *noise) shiftRegister() {
	tap := uint16(1)
	if o.mode {
		tap = 6
	}
	feedback := (o.shift & 1) ^ ((o.shift >> tap) & 1)
	o.shift >>= 1
	o.shift |= feedback << 14
}

func (o *noise) clockLength() {
	if !o.lengthHalt && o.lengthCounter > 0 {
		o.lengthCounter--
	}
}

func (o *noise) output() byte {
	if !o.enabled || o.lengthCounter == 0 || o.shift&1 != 0 {
		return 0
	}
	return o.envelope.output()
}
