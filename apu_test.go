package main

import "testing"

func TestAPUStatusEnablesAndClearsLengthCounters(t *testing.T) {
	apu := NewAPU(nil)

	apu.WriteRegister(0x4015, 0x01)
	apu.WriteRegister(0x4003, 0x18)

	if apu.ReadRegister(0x4015)&0x01 == 0 {
		t.Fatal("pulse 1 length counter was not reported active")
	}

	apu.WriteRegister(0x4015, 0x00)

	if apu.pulse1.lengthCounter != 0 {
		t.Fatalf("pulse 1 length counter = %d, want 0", apu.pulse1.lengthCounter)
	}
	if apu.ReadRegister(0x4015)&0x01 != 0 {
		t.Fatal("pulse 1 was reported active after being disabled")
	}
}

func TestPulseProducesAlternatingOutput(t *testing.T) {
	apu := NewAPU(nil)
	apu.WriteRegister(0x4015, 0x01)
	apu.WriteRegister(0x4000, 0x9F) // 50% duty, constant volume 15
	apu.WriteRegister(0x4002, 0x08)
	apu.WriteRegister(0x4003, 0x00)

	seenSilent := false
	seenAudible := false
	for i := 0; i < 200; i++ {
		apu.pulse1.stepTimer()
		if apu.pulse1.output() == 0 {
			seenSilent = true
		} else {
			seenAudible = true
		}
	}

	if !seenSilent || !seenAudible {
		t.Fatalf("pulse output did not alternate: silent=%v audible=%v", seenSilent, seenAudible)
	}
}

func TestTriangleRequiresLinearAndLengthCounters(t *testing.T) {
	apu := NewAPU(nil)
	apu.WriteRegister(0x4015, 0x04)
	apu.WriteRegister(0x4008, 0x83)
	apu.WriteRegister(0x400A, 0x02)
	apu.WriteRegister(0x400B, 0x00)

	if apu.triangle.output() != 0 {
		t.Fatal("triangle produced output before the linear counter was clocked")
	}

	apu.clockQuarterFrame()
	for i := 0; i < 8; i++ {
		apu.triangle.stepTimer()
	}

	if apu.triangle.output() == 0 {
		t.Fatal("triangle produced no output after linear and length counters were active")
	}
}

func TestNoiseShiftRegisterAdvances(t *testing.T) {
	apu := NewAPU(nil)
	before := apu.noise.shift

	apu.WriteRegister(0x4015, 0x08)
	apu.WriteRegister(0x400C, 0x1F)
	apu.WriteRegister(0x400E, 0x00)
	apu.WriteRegister(0x400F, 0x00)
	apu.clockQuarterFrame()
	for i := 0; i < 16; i++ {
		apu.noise.stepTimer()
	}

	if apu.noise.shift == before {
		t.Fatal("noise shift register did not advance")
	}
}

func TestFrameCounterClocksLengthCounters(t *testing.T) {
	apu := NewAPU(nil)
	apu.WriteRegister(0x4015, 0x01)
	apu.WriteRegister(0x4000, 0x10)
	apu.WriteRegister(0x4003, 0x08)

	before := apu.pulse1.lengthCounter
	for i := 0; i < 14913; i++ {
		apu.Step()
	}

	if apu.pulse1.lengthCounter != before-1 {
		t.Fatalf("pulse 1 length counter = %d, want %d", apu.pulse1.lengthCounter, before-1)
	}
}

func TestStepGeneratesPCMSamples(t *testing.T) {
	apu := NewAPU(nil)
	apu.WriteRegister(0x4015, 0x01)
	apu.WriteRegister(0x4000, 0x9F)
	apu.WriteRegister(0x4002, 0x20)
	apu.WriteRegister(0x4003, 0x00)

	for i := 0; i < cpuFreq/audioSampleRate+2; i++ {
		apu.Step()
	}

	samples := apu.TakeSamples()
	if len(samples) == 0 {
		t.Fatal("APU generated no PCM bytes")
	}
	if len(samples)%2 != 0 {
		t.Fatalf("PCM byte length = %d, want an even int16 byte count", len(samples))
	}
	if len(apu.TakeSamples()) != 0 {
		t.Fatal("TakeSamples did not drain the pending buffer")
	}
}
