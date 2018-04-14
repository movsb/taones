package main

import (
	"image"
)

// PPU 控制寄存器 $2000
type PPUCTRL struct {
	ctrlNameTable       byte // 命名表基地址 0: $2000, 1: $2400, 2: $2800, 3: $2C00
	ctrlIncrement       byte // 在 CPU 读写 PPUDATA 后的 VRAM 地址增量： 0: +1, 1: +32
	ctrlSpriteTable     byte // 8x8 精灵图案表地址 0: $0000, 1: $1000；8x16 时忽略
	ctrlBackgroundTable byte // 背景图案表地址 0: $0000, 1: $1000
	ctrlSpriteSize      byte // 精灵大小 0: 8x8, 1: 8x16
	ctrlMasterSlave     byte // 主/从模式选择
	ctrlEnableNMI       byte // 使能 VBlank NMI 中断 0: 关闭，1: 打开
}

func (o *PPUCTRL) Set(v byte) {
	o.ctrlNameTable = v >> 0 & 3
	o.ctrlIncrement = v >> 2 & 1
	o.ctrlSpriteTable = v >> 3 & 1
	o.ctrlBackgroundTable = v >> 4 & 1
	o.ctrlSpriteSize = v >> 5 & 1
	o.ctrlMasterSlave = v >> 6 & 1
	o.ctrlEnableNMI = v >> 7 & 1
}

// PPU 掩码寄存器 $2001
type PPUMASK struct {
	maskGrayscale          byte // 灰阶图案 0: 彩色，1: 灰阶
	maskShowLeftBackground byte // 是否显示最左边的8个像素：1: 显示，0: 隐藏
	maskShowLeftSprites    byte // 是否显示最左边的8个像素内的精灵：1: 显示，0: 隐藏
	maskShowBackground     byte // 是否显示背景 1: 显示
	maskShowSprites        byte // 是否显示精灵 1: 显示
	maskEmphasizeRed       byte
	maskEmphasizeGreen     byte
	maskEmphasizeBlue      byte
}

func (o *PPUMASK) Set(v byte) {
	o.maskGrayscale = v >> 0 & 1
	o.maskShowLeftBackground = v >> 1 & 1
	o.maskShowLeftSprites = v >> 2 & 1
	o.maskShowBackground = v >> 3 & 1
	o.maskShowSprites = v >> 4 & 1
	o.maskEmphasizeRed = v >> 5 & 1
	o.maskEmphasizeGreen = v >> 6 & 1
	o.maskEmphasizeBlue = v >> 7 & 1
}

// PPU 状态寄存器 $2002
type PPUSTAT struct {
	statSpriteOverflow byte // 精灵数量溢出
	statSpriteHit      byte // 精灵碰撞
	statVBlank         byte // VBlank 已开始
}

func (o *PPUSTAT) Set(v byte) {
	o.statSpriteOverflow = v >> 5 & 1
	o.statSpriteHit = v >> 6 & 1
	o.statVBlank = v >> 7 & 1
}

// 精灵地址 $2003
type OAMADDR byte

// 精灵数据 $2004
type OAMDATA byte

// 滚动寄存器 $2005
type PPUSCRL byte

// PPU 写地址 $2006
type PPUADDR byte

// PPU 读写数据 $2007
type PPUDATA byte

// 精灵 DMA 寄存器
type OAMDMAR byte

type PPU struct {
	MemoryReadWriter
	console *Console

	Cycle      int // [0,340]
	Scanline   int // [0,261]
	FrameCount uint64

	palette   [32]byte
	nameTable [2048]byte
	sprite    [256]byte
	front     *image.RGBA
	back      *image.RGBA

	// 寄存器
	PPUCTRL
	PPUMASK
	PPUSTAT
	OAMADDR
	OAMDATA
	PPUSCRL
	PPUADDR
	PPUDATA
	OAMDMAR

	oddFrame bool // 奇偶帧标志

	nmiOccurred bool // 中断标志：进入 VBlank
	// 功能同 ctrlEnableNMI
	// nmiOutput bool // 开启中断输出
}

func NewPPU(console *Console) *PPU {
	ppu := PPU{}
	ppu.front = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.back = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.Power()
	return &ppu
}

func (o *PPU) Power() {
	o.PPUCTRL.Set(0x00)
	o.PPUMASK.Set(0x00)
	o.PPUSTAT.Set(0xA0)
}
