package main

import (
	"image"
	"log"
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

func (o *PPUSTAT) Get(nmiOccurred bool) byte {
	var v byte
	v |= o.statSpriteOverflow << 5
	v |= o.statSpriteHit << 6
	if nmiOccurred {
		v |= o.statVBlank << 7
	}
	return v
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

	palette   [32]byte
	nameTable [2048]byte
	oam       [256]byte
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

	// 内部奇偶帧标志位
	// 每完成一帧，切换一次
	// 不论是否开启渲染，都会切换
	// 奇帧比偶帧少一个周期
	oddFrame bool

	nmiOccurred bool // 中断标志：进入 VBlank
	// 功能同 ctrlEnableNMI
	// nmiOutput bool // 开启中断输出

	// 当前 VRAM 地址，15位
	// yyy NN YYYYY XXXXX
	// ||| || ||||| +++++-- 粗糙 X 滚动（Tile 列）
	// ||| || +++++-------- 粗糙 Y 滚动（Tile 行）
	// ||| ++-------------- 命名表选择（2位，4个表）
	// +++----------------- 精确 Y 滚动（Tile 级别）
	v uint16
	// 临时 VRAM 地址，15位
	// 即：当前可见屏幕区域最左上角的Tile的地址
	t uint16
	// 精确 X 滚动，3位
	// [0,7] Tile 内的像素点列
	x byte
	// 第1次/第2次写切换
	w bool

	// 扫描一条扫描线时的扫描周期数
	// NTSC 一条扫描线是 341 个周期
	Cycle int // [0,340]

	// 一帧的扫描线数
	// NTSC 是 262 条
	// 所以总扫描周期是：341*262 = 89342
	Scanline int // [0,261]

	// 帧计数器
	FrameCount uint64

	// 渲染临时数据
	// 两个 Tiles 的位图数据
	nameTableByte      byte
	attributeTableByte byte
	tileByteLo         byte
	tileByteHi         byte

	// 渲染图块所需数据（当前32位+下一个32位）
	// 一个图块一行8个像素，一个像素4个位，总共32位
	tileData uint64

	bufferedData byte
	register     byte

	spriteCount     int
	spritePatterns  [8]uint32
	spritePositions [8]byte
	spritePriorites [8]byte
	spriteIndexes   [8]byte
}

func NewPPU(console *Console) *PPU {
	ppu := PPU{MemoryReadWriter: NewPPUMemory(console), console: console}
	ppu.front = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.back = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.Power()
	return &ppu
}

func (ppu *PPU) readRegister(address uint16) byte {
	switch address {
	case 0x2002:
		return ppu.readStatus()
	case 0x2007:
		return ppu.readData()
	default:
		log.Fatalln("err reading ppu register", address)
	}
	return 0
}

func (ppu *PPU) writeRegister(address uint16, value byte) {
	ppu.register = value
	switch address {
	case 0x2000:
		ppu.writeControl(value)
	case 0x2001:
		ppu.PPUMASK.Set(value)
	case 0x2003:
		//ppu.writeOAMAddress(value)
	case 0x2004:
		//ppu.writeOAMData(value)
		break
	case 0x2005:
		ppu.writeScroll(value)
	case 0x2006:
		ppu.writeAddress(value)
	case 0x2007:
		ppu.writeData(value)
	case 0x4014:
		//ppu.writeDMA(value)
	default:
		log.Fatalf("error write ppu register: %x=%x\n", address, value)
	}
}

func (o *PPU) Power() {
	o.PPUCTRL.Set(0x00)
	o.PPUMASK.Set(0x00)
	o.PPUSTAT.Set(0xA0)

	o.Cycle = 340
	o.Scanline = 240
	o.FrameCount = 0
}

// $2007: PPUDATA (read)
func (ppu *PPU) readData() byte {
	value := ppu.Read(ppu.v)
	// emulate buffered reads
	if ppu.v%0x4000 < 0x3F00 {
		buffered := ppu.bufferedData
		ppu.bufferedData = value
		value = buffered
	} else {
		ppu.bufferedData = ppu.Read(ppu.v - 0x1000)
	}
	// increment address
	if ppu.ctrlIncrement == 0 {
		ppu.v += 1
	} else {
		ppu.v += 32
	}
	return value
}

func (o *PPU) writeData(v byte) {
	o.Write(o.v, v)
	if o.ctrlIncrement == 0 {
		o.v++
	} else {
		o.v += 32
	}
}

func (o *PPU) setVBlank() {
	o.front, o.back = o.back, o.front
	o.nmiOccurred = true
}

func (o *PPU) clrVBlank() {
	o.nmiOccurred = false
}

// 写 $2000
func (o *PPU) writeControl(v byte) {
	o.PPUCTRL.Set(v)
	// PPU_scrolling
	// t: ...BA.. ........ = d: ......BA
	o.t = o.t&0x73FF | uint16(v)&0x0003<<10
}

// 读 $2002
func (o *PPU) readStatus() byte {
	v := o.register & 0x1F
	v |= o.PPUSTAT.Get(o.nmiOccurred)

	// PPU_scrolling
	// w:                  = 0
	o.w = false

	o.nmiOccurred = false

	return v
}

// 写 $2005 PPU_scrolling
// 第1次写 (w is 0)
// t: ....... ...HGFED = d: HGFED...
// x:              CBA = d: .....CBA
// w:                  = 1
// 第2次写 (w is 1)
// t: CBA..HG FED..... = d: HGFEDCBA
// w:                  = 0
func (o *PPU) writeScroll(d byte) {
	dd := uint16(d)

	if !o.w {
		o.t = o.t&0xFFE0 | dd>>3
		o.x = d & 7
	} else {
		o.t = o.t&0x8C1F | (dd&7)<<12 | (dd&0xF8)<<2
	}
	o.w = !o.w
}

// 写 $2006 PPU_scrolling
// 第1次写 (w is 0)
// t: .FEDCBA ........ = d: ..FEDCBA
// t: X...... ........ = 0
// w:                  = 1
// 第2次写(w is 1)
// t: ....... HGFEDCBA = d: HGFEDCBA
// v                   = t
// w:                  = 0
func (o *PPU) writeAddress(d byte) {
	dd := uint16(d)

	if !o.w {
		o.t = o.t&0x00FF | (dd&0x3F)<<8
	} else {
		o.t = o.t&0x7F00 | dd
		o.v = o.t
	}
	o.w = !o.w
}

// 读调色板
func (o *PPU) readPalette(a uint16) byte {
	a &= 0x1F
	return o.palette[a]
}

// 写调色板
func (o *PPU) writePalette(a uint16, v byte) {
	a &= 0x1F
	o.palette[a] = v & 0x3F
}

// PPU 显存 X 坐标写入增加
// 绘制下一个图块时增加
func (o *PPU) incX() {
	// 如果到达了命名表的边界
	if o.v&0x1F == 31 {
		// Tile 列从 0 开始
		o.v &^= 0x1F
		// 切换水平命名表
		o.v ^= 0x0400
	} else {
		o.v++
	}
}

// PPU 显存 Y 坐标写入增加
// 绘制下一条扫描线的时候增加
func (o *PPU) incY() {
	// 如果没有跨越图块
	if o.v&0x7000 != 0x7000 {
		// 下移一列
		o.v += 0x1000
	} else {
		// 跨越图块，从0开始
		o.v &^= 0x7000
		// 得到当前扫描号
		// 判断是否需要切垂直命名表
		r := o.v >> 5 & 0x1F
		if r == 29 { // [0,29]，29是最后一行显示行
			r = 0         // 回到第0行
			o.v ^= 0x0800 // 切换垂直命名表
		} else {
			r++ // TODO 没有处理 r > 29 的情况
		}
		// 把新的扫描行号写回vram地址
		o.v = o.v&^0x03E0 | r<<5
	}
}

/* 开始抓数据逻辑 */

// 从当前地址抓取命名表：一个字节
func (o *PPU) fetchNameTable() {
	// $2000 - $2FFF
	a := 0x2000 | o.v&0x0FFF
	o.nameTableByte = o.Read(a)
}

// 为当前命名表爬取属性表：一个字节
/*
属性表与命名表对应关系：

20 | 00 01 02 03 | 04 05 06 07 | ... | 1C 1D 1E 1F
20 | 20 21 22 23 |
20 | 40 41 42 43 |
20 | 60 61 62 63 |-> 23C0 + 0
 ... ... ... ...
23 | 80 81 82 83
23 | A0 A1 A2 A3 | ..................| BC BD BE BF
23 | C0 C1 C2 C3

属性表一个字节，代表命名表 4x4 个 图块
每4行增加8个属性表字节

          7
2000 0 0  0 +8个字节
2080 0 0  1 +8
2100 0 1  0 +8
2180 0 1  1 +8
2200 1 0  0 +8
2280 1 0  1 +8
2300 1 1  0 +8
2380 1 1  1 +8
 ...        +8
2C00 0 0  0 +8
 ...        +8
2F80 1 1  1 +8

属性起始地址：由第11、10位决定 0x23C0 | v & 0x0C00
通过精确滚动Y坐标的高3位可以得到属性表行起始地址：(v >> 7 & 7) * 8 == (v >> 4 & 0x38)
v / 4 & 7 得到 [0,7] 范围内的字节

一个属性表字节控制4x4个图块，那么：X的第1位，Y的第1位共同决定图块使用此字节的的哪两位
start = ((v >> 5 & 2) | (v >> 1 & 1)) << 1
       -------------   -------------
       决定图块的奇偶行    决定图块的列

	可简化为：start = (v >> 4 & 4) | (v & 2)，此即图块的高两位颜色在属性字节中的开始位
	start 可能为：0，2，4，6
	attr >> start 得到两位数据
	保留低两位并左移两位放到正确位置（低两位是后面图块的）
*/
func (o *PPU) fetchAttributeTableByte() {
	var (
		offset = (0x23C0 | o.v&0x0C00) + (o.v >> 4 & 0x38) + (o.v >> 2 & 7)
		start  = (o.v >> 4 & 4) | (o.v & 2)
		attr   = o.Read(offset)
	)

	o.attributeTableByte = attr >> start & 3 << 2
}

// 抓取图块第0面数据
func (o *PPU) fetchTileByteLo() {
	fineY := o.v >> 12 & 7
	patternTable := uint16(o.ctrlBackgroundTable) * 0x1000
	tile := o.nameTableByte
	a := patternTable + uint16(tile)*16 + fineY
	o.tileByteLo = o.Read(a)
}

// 抓取图块第1面数据
func (o *PPU) fetchTileByteHi() {
	fineY := o.v >> 12 & 7
	patternTable := uint16(o.ctrlBackgroundTable) * 0x1000
	tile := o.nameTableByte
	a := patternTable + uint16(tile)*16 + fineY + 8
	o.tileByteHi = o.Read(a)
}

// 根据当前 fetch 到的数据
// 生成一个 tile 的数据
func (o *PPU) makeTileData() {
	p3p2 := o.attributeTableByte

	var d uint32

	for i := 0; i < 8; i++ {
		p0 := o.tileByteLo >> 7 << 0 // 第0位
		p1 := o.tileByteHi >> 7 << 1 // 第1位
		o.tileByteLo <<= 1
		o.tileByteHi <<= 1
		d <<= 4
		d |= uint32(p3p2 | p1 | p0)
	}

	o.tileData |= uint64(d)
}

func (o *PPU) backgroundPixel() byte {
	if o.maskShowBackground == 0 {
		return 0
	}

	d := o.tileData >> 32 >> ((7 - o.x) * 4)
	return byte(d & 0x0F)
}

func (o *PPU) spritePixel() (byte, byte) {
	if o.maskShowSprites == 0 {
		return 0, 0
	}

	for i := 0; i < o.spriteCount; i++ {
		offset := o.Cycle - 1 - int(o.spritePositions[i])
		if offset < 0 || offset > 7 {
			continue
		}
		offset = 7 - offset
		color := byte(o.spritePatterns[i] >> byte(offset*4) & 0x0F)
		if color&3 == 0 {
			continue
		}
		return byte(i), color
	}

	return 0, 0
}

func (o *PPU) renderPixel() {
	x := o.Cycle - 1
	y := o.Scanline

	background := o.backgroundPixel()
	i, sprite := o.spritePixel()

	if x < 8 && o.maskShowLeftBackground == 0 {
		background = 0
	}

	if x < 8 && o.maskShowLeftSprites == 0 {
		sprite = 0
	}

	b := background&3 != 0
	s := sprite&3 != 0

	var color byte

	switch {
	case b && s:
		if o.spriteIndexes[i] == 0 && x < 255 {
			o.statSpriteHit = 1
		}
		if o.spritePriorites[i] == 0 {
			color = sprite | 0x10
		} else {
			color = background
		}
	case b:
		color = background
	case s:
		color = sprite | 0x10
	default:
		color = 0
	}

	c := Palette[o.readPalette(uint16(color))&0x3F]
	o.back.SetRGBA(x, y, c)
}

func (o *PPU) copyX() {
	// hori(v) = hori(t)
	// v: .....F.. ...EDCBA = t: .....F.. ...EDCBA
	o.v = (o.v & 0xFBE0) | (o.t & 0x041F)
}

func (o *PPU) copyY() {
	// vert(v) = vert(t)
	// v: .IHGF.ED CBA..... = t: .IHGF.ED CBA.....
	o.v = (o.v & 0x841F) | (o.t & 0x7BE0)
}

func (o *PPU) tick() {
	if o.nmiOccurred && o.ctrlEnableNMI == 1 {
		o.console.cpu.triggerNMI()
	}

	if o.maskShowBackground != 0 || o.maskShowSprites != 0 {
		if o.oddFrame && o.Scanline == 261 && o.Cycle == 339 {
			o.Cycle = 0
			o.Scanline = 0
			o.FrameCount++
			o.oddFrame = !o.oddFrame
			return
		}
	}

	if o.Cycle++; o.Cycle > 340 {
		o.Cycle = 0
		if o.Scanline++; o.Scanline > 261 {
			o.Scanline = 0
			o.FrameCount++
			o.oddFrame = !o.oddFrame
		}
	}
}

// 精灵评估与精灵渲染是两个独立的过程
// 所以此处一次性搞定精灵评估操作
// 而没有按照周期一步一步来
func (o *PPU) evaluateSprites() {
	var h = 8
	if o.ctrlSpriteSize == 1 {
		h = 16
	}

	count := 0

	for i := 0; i < 64; i++ {
		y := o.oam[i*4+0]
		a := o.oam[i*4+2]
		x := o.oam[i*4+3]

		r := o.Scanline - int(y)
		if r < 0 || r >= h {
			continue
		}

		if count < 8 {
			o.spritePatterns[count] = o.fetchSpritePattern(i, r)
			o.spritePositions[count] = x
			o.spritePriorites[count] = a >> 5 & 1
			o.spriteIndexes[count] = byte(i)
		}
		count++
	}
	if count > 8 {
		count = 8
		o.statSpriteOverflow = 1
	}
	o.spriteCount = count
}

// 获取第i个精灵第row条扫描线的渲染数据
func (o *PPU) fetchSpritePattern(i int, row int) uint32 {
	tile := o.oam[i*4+1]
	attr := o.oam[i*4+2]

	var addr uint16

	if o.ctrlSpriteSize == 0 { // 8x8
		// 如果垂直翻转
		if attr&0x80 == 0x80 {
			row = 7 - row
		}
		table := 0x1000 * uint16(o.ctrlSpriteTable)
		addr = table + uint16(tile)*16 + uint16(row)
	} else { // 8x16
		// 如果垂直翻转翻转
		if attr&0x80 == 0x80 {
			row = 15 - row
		}
		// 第0位决定使用的图案表
		table := 0x1000 * uint16(tile&1)

		tile &= 0xFE
		if row > 7 {
			tile++
			row -= 8
		}

		addr = table + uint16(tile)*16 + uint16(row)
	}

	lo := o.Read(addr)
	hi := o.Read(addr + 8)
	p3p2 := attr & 3 << 2
	var data uint32

	for i := 0; i < 8; i++ {
		var p1, p0 byte
		if attr&0x40 == 0x40 { // 水平翻转
			p0 = lo & 1 << 0
			p1 = hi & 1 << 1
			lo >>= 1
			hi >>= 1
		} else {
			p0 = lo & 0x80 >> 7 << 0
			p1 = lo & 0x80 >> 7 << 1
			lo <<= 1
			hi <<= 1
		}

		data <<= 4
		data |= uint32(p3p2 | p1 | p0)
	}

	return data
}

// PPU 步进一个周期
func (o *PPU) Step() {
	o.tick()

	renderEnabled := o.maskShowBackground != 0 || o.maskShowSprites != 0
	visibleCycle := o.Cycle >= 1 && o.Cycle <= 256
	prefetchCycle := o.Cycle >= 321 && o.Cycle <= 336
	visibleLine := o.Scanline < 240
	fetchCycle := prefetchCycle || visibleCycle
	preLine := o.Scanline == 261

	if renderEnabled {
		if visibleLine && visibleCycle {
			o.renderPixel()
		}

		if (visibleLine || preLine) && fetchCycle {
			// 每个周期渲染一个像素，消耗掉4位
			o.tileData <<= 4
			switch o.Cycle & 7 {
			case 1:
				o.fetchNameTable()
			case 3:
				o.fetchAttributeTableByte()
			case 5:
				o.fetchTileByteLo()
			case 7:
				o.fetchTileByteHi()
			case 0:
				o.makeTileData()
			}
		}

		if preLine && o.Cycle >= 280 && o.Cycle <= 304 {
			o.copyY()
		}

		if visibleLine || preLine {
			// 水平移动
			if fetchCycle && o.Cycle&7 == 0 {
				o.incX()
			}

			// 换行
			if o.Cycle == 256 {
				o.incY()
			}

			if o.Cycle == 257 {
				o.copyX()
			}
		}
	}

	if renderEnabled && o.Cycle == 257 {
		if visibleLine {
			o.evaluateSprites()
		} else {
			o.spriteCount = 0
		}
	}

	if o.Cycle == 1 {
		switch o.Scanline {
		case 241:
			o.setVBlank()
		case 261:
			o.clrVBlank()
			o.statSpriteHit = 0
			o.statSpriteOverflow = 0
		}
	}
}
