package main

import (
	"fmt"
)

const cpuFreq = 1789773 / 1.5

// 寻址模式（Addressing Modes）
const (
	_                 byte = iota
	amImmediate            // 1  立即
	amZero                 // 2  零页索引
	amZeroX                // 3  零页直接
	amZeroY                // 4  零页直接
	amAbsolute             // 5  绝对
	amAbsoluteX            // 6  绝对X
	amAbsoluteY            // 7  绝对Y
	amIndirect             // 8  间接
	amIndexedIndirect      // 9  先索引后间接
	amIndirectIndexed      // 10 先间接后索引
	amRelative             // 11 相对
	amImplied              // 12 隐含
	amAccumulator          // 13 累加器
)

// 中断模式
const (
	intNone = iota
	intNMI
	intIRQ
)

var opcodeModes = [256]byte{
	12, 9, 0, 0, 0, 2, 2, 0, 12, 1, 13, 0, 0, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
	5, 9, 0, 0, 2, 2, 2, 0, 12, 1, 13, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
	12, 9, 0, 0, 0, 2, 2, 0, 12, 1, 13, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
	12, 9, 0, 0, 0, 2, 2, 0, 12, 1, 13, 0, 8, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
	0, 9, 0, 0, 2, 2, 2, 0, 12, 0, 12, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 3, 3, 4, 0, 12, 7, 12, 0, 0, 6, 0, 0,
	1, 9, 1, 0, 2, 2, 2, 0, 12, 1, 12, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 3, 3, 4, 0, 12, 7, 12, 0, 6, 6, 7, 0,
	1, 9, 0, 0, 2, 2, 2, 0, 12, 1, 12, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
	1, 9, 0, 0, 2, 2, 2, 0, 12, 1, 12, 0, 5, 5, 5, 0,
	11, 10, 0, 0, 0, 3, 3, 0, 12, 7, 0, 0, 0, 6, 6, 0,
}

var opcodeSizes = [256]byte{
	1, 2, 1, 1, 1, 2, 2, 1, 1, 2, 1, 1, 1, 3, 3, 1,
	2, 2, 1, 1, 1, 2, 2, 1, 1, 3, 1, 1, 1, 3, 3, 1,
	3, 2, 1, 1, 2, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 1, 3, 2, 1, 1, 2, 1, 1, 1, 3, 3, 1,
	1, 2, 1, 1, 1, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 1, 2, 2, 1, 1, 3, 1, 1, 1, 3, 3, 1,
	1, 2, 1, 1, 1, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 1, 2, 2, 1, 1, 3, 1, 1, 1, 3, 3, 1,
	1, 2, 1, 1, 2, 2, 2, 1, 1, 1, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 2, 2, 2, 1, 1, 3, 1, 1, 1, 3, 1, 1,
	2, 2, 2, 1, 2, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 2, 2, 2, 1, 1, 3, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 2, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 1, 2, 2, 1, 1, 3, 1, 1, 1, 3, 3, 1,
	2, 2, 1, 1, 2, 2, 2, 1, 1, 2, 1, 1, 3, 3, 3, 1,
	2, 2, 1, 1, 1, 2, 2, 1, 1, 3, 1, 1, 1, 3, 3, 1,
}

var opcodePagedSize = [256]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
}

var opcodeCycles = [256]byte{
	1, 6, 1, 1, 1, 3, 5, 1, 3, 2, 3, 1, 1, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
	6, 6, 1, 1, 3, 3, 5, 1, 4, 2, 2, 1, 4, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
	6, 6, 1, 1, 1, 3, 5, 1, 3, 2, 2, 1, 3, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
	6, 6, 1, 1, 1, 3, 5, 1, 4, 3, 2, 1, 5, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
	1, 6, 1, 1, 3, 3, 3, 1, 2, 1, 2, 1, 4, 4, 4, 1,
	2, 6, 1, 1, 4, 4, 4, 1, 2, 5, 2, 1, 1, 5, 1, 1,
	2, 6, 2, 1, 3, 3, 3, 1, 2, 2, 2, 1, 4, 4, 4, 1,
	2, 5, 1, 1, 4, 4, 4, 1, 2, 4, 2, 1, 4, 4, 4, 1,
	2, 6, 1, 1, 3, 3, 5, 1, 2, 2, 2, 1, 4, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
	2, 6, 1, 1, 3, 3, 5, 1, 2, 2, 2, 1, 4, 4, 6, 1,
	2, 5, 1, 1, 1, 4, 6, 1, 2, 4, 1, 1, 1, 4, 7, 1,
}

var opcodeNames = [256]string{
	"brk", "ora", "bad", "bad", "bad", "ora", "asl", "bad",
	"php", "ora", "asl", "bad", "bad", "ora", "asl", "bad",
	"bpl", "ora", "bad", "bad", "bad", "ora", "asl", "bad",
	"clc", "ora", "bad", "bad", "bad", "ora", "asl", "bad",
	"jsr", "and", "bad", "bad", "bit", "and", "rol", "bad",
	"plp", "and", "rol", "bad", "bit", "and", "rol", "bad",
	"bmi", "and", "bad", "bad", "bad", "and", "rol", "bad",
	"sec", "and", "bad", "bad", "bad", "and", "rol", "bad",
	"rti", "eor", "bad", "bad", "bad", "eor", "lsr", "bad",
	"pha", "eor", "lsr", "bad", "jmp", "eor", "lsr", "bad",
	"bvc", "eor", "bad", "bad", "bad", "eor", "lsr", "bad",
	"cli", "eor", "bad", "bad", "bad", "eor", "lsr", "bad",
	"rts", "adc", "bad", "bad", "bad", "adc", "ror", "bad",
	"pla", "adc", "ror", "bad", "jmp", "adc", "ror", "bad",
	"bvs", "adc", "bad", "bad", "bad", "adc", "ror", "bad",
	"sei", "adc", "bad", "bad", "bad", "adc", "ror", "bad",
	"bad", "sta", "bad", "bad", "sty", "sta", "stx", "bad",
	"dey", "bad", "txa", "bad", "sty", "sta", "stx", "bad",
	"bcc", "sta", "bad", "bad", "sty", "sta", "stx", "bad",
	"tya", "sta", "txs", "bad", "bad", "sta", "bad", "bad",
	"ldy", "lda", "ldx", "bad", "ldy", "lda", "ldx", "bad",
	"tay", "lda", "tax", "bad", "ldy", "lda", "ldx", "bad",
	"bcs", "lda", "bad", "bad", "ldy", "lda", "ldx", "bad",
	"clv", "lda", "tsx", "bad", "ldy", "lda", "ldx", "bad",
	"cpy", "cmp", "bad", "bad", "cpy", "cmp", "dec", "bad",
	"iny", "cmp", "dex", "bad", "cpy", "cmp", "dec", "bad",
	"bne", "cmp", "bad", "bad", "bad", "cmp", "dec", "bad",
	"cld", "cmp", "bad", "bad", "bad", "cmp", "dec", "bad",
	"cpx", "sbc", "bad", "bad", "cpx", "sbc", "inc", "bad",
	"inx", "sbc", "nop", "bad", "cpx", "sbc", "inc", "bad",
	"beq", "sbc", "bad", "bad", "bad", "sbc", "inc", "bad",
	"sed", "sbc", "bad", "bad", "bad", "sbc", "inc", "bad",
}

type stepContext struct {
	a    uint16
	pc   uint16
	mode byte
}

type opCodeFunc func(*stepContext)

// 状态标志位（Status Bits）
type Flag struct {
	C byte // 进位（Carry）
	Z byte // 零（Zero）
	I byte // 中断控制
	D byte // 十进制模式
	B byte // 中断
	U byte // 未使用
	V byte // 溢出
	N byte // 负数
}

func (o *Flag) SetFlags(bits byte) {
	o.C = bits >> 0 & 1
	o.Z = bits >> 1 & 1
	o.I = bits >> 2 & 1
	o.D = bits >> 3 & 1
	o.B = bits >> 4 & 1
	o.U = bits >> 5 & 1
	o.V = bits >> 6 & 1
	o.N = bits >> 7 & 1
}

func (o *Flag) GetFlags() byte {
	var f byte
	f |= o.C << 0
	f |= o.Z << 1
	f |= o.I << 2
	f |= o.D << 3
	f |= o.B << 4
	f |= o.U << 5
	f |= o.V << 6
	f |= o.N << 7
	return f
}

func (o *Flag) SetZ(V byte) {
	if V == 0 {
		o.Z = 1
	} else {
		o.Z = 0
	}
}

func (o *Flag) SetN(V byte) {
	if V&0x80 != 0 {
		o.N = 1
	} else {
		o.N = 0
	}
}

func (o *Flag) SetZN(V byte) {
	o.SetN(V)
	o.SetZ(V)
}

type CPU struct {
	A                byte            // 累加寄存器（Accumulator）
	X                byte            // 索引寄存器（Index Register）
	Y                byte            // 索引寄存器（Index Register）
	SP               byte            // 栈指针（Stack Pointer）
	PC               uint16          // 指令指针
	Flag                             // 状态寄存器（Status Register）
	Cycles           uint64          // 指令周期
	opcodes          [256]opCodeFunc // 指令执行函数
	irq              byte            // 当前中断
	RAM              [2048]byte      // CPU RAM
	MemoryReadWriter                 // 内存读写实现
	suspendCycles    uint32          // 暂时执行的周期数（比如DMA发生时）
}

func NewCPU(console *Console) *CPU {
	cpu := &CPU{}
	cpu.createOpcodeFuncs()
	cpu.MemoryReadWriter = NewCPUMemory(console)
	return cpu
}

func (o *CPU) Reset() {
	o.PC = o.Read16(0xFFFC)
	o.SP = 0xFD
	o.SetFlags(0x24)
}

func (o *CPU) PrintInstruction(opcode byte, pc uint16) {
	return
	bytes := opcodeSizes[opcode]
	name := opcodeNames[opcode]
	w0 := fmt.Sprintf("%02X", o.Read(pc+0))
	w1 := fmt.Sprintf("%02X", o.Read(pc+1))
	w2 := fmt.Sprintf("%02X", o.Read(pc+2))
	if bytes < 2 {
		w1 = "  "
	}
	if bytes < 3 {
		w2 = "  "
	}
	fmt.Printf(
		"%4X  %s %s %s  %s %8s"+
			"A:%02X X:%02X Y:%02X P:%02X SP:%02X\n",
		o.PC, w0, w1, w2, name, "",
		o.A, o.X, o.Y, o.GetFlags(), o.SP)
}

func (o *CPU) Step() int {
	if o.suspendCycles > 0 {
		o.suspendCycles--
		return 1
	}

	cycles := o.Cycles

	switch o.irq {
	case intNMI:
		o.nmiSvc()
	case intIRQ:
		o.irqSvc()
	}

	o.irq = intNone

	if o.PC == 0x90CE {
		o.PC = o.PC
	}

	opcode := o.Read(o.PC)
	mode := opcodeModes[opcode]
	var A uint16
	var paged bool

	switch mode {
	case amAbsolute:
		A = o.Read16(o.PC + 1)
	case amAbsoluteX:
		A = o.Read16(o.PC+1) + uint16(o.X)
		paged = pagesDiffer(A-uint16(o.X), A)
	case amAbsoluteY:
		A = o.Read16(o.PC+1) + uint16(o.Y)
		paged = pagesDiffer(A-uint16(o.Y), A)
	case amAccumulator:
		A = 0
	case amImmediate:
		A = o.PC + 1
	case amImplied:
		A = 0
	case amIndexedIndirect:
		A = o.Read16Bug(uint16(o.Read(o.PC+1) + o.X))
	case amIndirect:
		A = o.Read16Bug(o.Read16(o.PC + 1))
	case amIndirectIndexed:
		A = o.Read16Bug(uint16(o.Read(o.PC+1))) + uint16(o.Y)
		paged = pagesDiffer(A-uint16(o.Y), A)
	case amRelative:
		offset := uint16(o.Read(o.PC + 1))
		if offset < 0x80 {
			A = o.PC + 2 + offset
		} else {
			A = o.PC + 2 + offset - 0x100
		}
	case amZero:
		A = uint16(o.Read(o.PC + 1))
	case amZeroX:
		A = uint16(o.Read(o.PC+1)+o.X) & 0xFF
	case amZeroY:
		A = uint16(o.Read(o.PC+1)+o.Y) & 0xFF
	}

	pcback := o.PC

	o.PC += uint16(opcodeSizes[opcode])
	o.Cycles += uint64(opcodeCycles[opcode])
	if paged {
		o.Cycles += uint64(opcodePagedSize[opcode])
	}

	ctx := &stepContext{A, o.PC, mode}
	o.opcodes[opcode](ctx)

	o.PrintInstruction(opcode, pcback)

	return int(o.Cycles - cycles)
}

func pagesDiffer(a, b uint16) bool {
	return a&0xFF00 != b&0xFF00
}

func (o *CPU) addBranchCycles(ctx *stepContext) {
	o.Cycles++
	if pagesDiffer(ctx.pc, ctx.a) {
		o.Cycles++
	}
}

func (o *CPU) compare(a, b byte) {
	o.SetZN(a - b)
	if a >= b {
		o.C = 1
	} else {
		o.C = 0
	}
}

func (o *CPU) Read16(A uint16) uint16 {
	lo := uint16(o.Read(A))
	hi := uint16(o.Read(A + 1))
	return hi<<8 | lo
}

func (o *CPU) Read16Bug(A uint16) uint16 {
	a := A
	b := (a & 0xFF00) | uint16(byte(a)+1)
	lo := o.Read(a)
	hi := o.Read(b)
	return uint16(hi)<<8 | uint16(lo)
}

func (o *CPU) push(v byte) {
	o.Write(0x100|uint16(o.SP), v)
	o.SP--
}

func (o *CPU) pop() byte {
	o.SP++
	return o.Read(0x100 | uint16(o.SP))
}

func (o *CPU) push16(v uint16) {
	o.push(byte(v >> 8))
	o.push(byte(v & 0xFF))
}

func (o *CPU) pop16() uint16 {
	lo := uint16(o.pop())
	hi := uint16(o.pop())
	return hi<<8 | lo
}

func (o *CPU) triggerNMI() {
	o.irq = intNMI
}

func (o *CPU) triggerIRQ() {
	if o.I == 0 {
		o.irq = intIRQ
	}
}

func (o *CPU) nmiSvc() {
	o.push16(o.PC)
	o.php(nil)
	o.PC = o.Read16(0xFFFA)
	o.Cycles += 7
}

func (o *CPU) irqSvc() {
	o.push16(o.PC)
	o.php(nil)
	o.PC = o.Read16(0xFFFE)
	o.I = 1
	o.Cycles += 7
}

func (o *CPU) createOpcodeFuncs() {
	o.opcodes = [256]opCodeFunc{
		o.brk, o.ora, o.bad, o.bad, o.bad, o.ora, o.asl, o.bad,
		o.php, o.ora, o.asl, o.bad, o.bad, o.ora, o.asl, o.bad,
		o.bpl, o.ora, o.bad, o.bad, o.bad, o.ora, o.asl, o.bad,
		o.clc, o.ora, o.bad, o.bad, o.bad, o.ora, o.asl, o.bad,
		o.jsr, o.and, o.bad, o.bad, o.bit, o.and, o.rol, o.bad,
		o.plp, o.and, o.rol, o.bad, o.bit, o.and, o.rol, o.bad,
		o.bmi, o.and, o.bad, o.bad, o.bad, o.and, o.rol, o.bad,
		o.sec, o.and, o.bad, o.bad, o.bad, o.and, o.rol, o.bad,
		o.rti, o.eor, o.bad, o.bad, o.bad, o.eor, o.lsr, o.bad,
		o.pha, o.eor, o.lsr, o.bad, o.jmp, o.eor, o.lsr, o.bad,
		o.bvc, o.eor, o.bad, o.bad, o.bad, o.eor, o.lsr, o.bad,
		o.cli, o.eor, o.bad, o.bad, o.bad, o.eor, o.lsr, o.bad,
		o.rts, o.adc, o.bad, o.bad, o.bad, o.adc, o.ror, o.bad,
		o.pla, o.adc, o.ror, o.bad, o.jmp, o.adc, o.ror, o.bad,
		o.bvs, o.adc, o.bad, o.bad, o.bad, o.adc, o.ror, o.bad,
		o.sei, o.adc, o.bad, o.bad, o.bad, o.adc, o.ror, o.bad,
		o.bad, o.sta, o.bad, o.bad, o.sty, o.sta, o.stx, o.bad,
		o.dey, o.bad, o.txa, o.bad, o.sty, o.sta, o.stx, o.bad,
		o.bcc, o.sta, o.bad, o.bad, o.sty, o.sta, o.stx, o.bad,
		o.tya, o.sta, o.txs, o.bad, o.bad, o.sta, o.bad, o.bad,
		o.ldy, o.lda, o.ldx, o.bad, o.ldy, o.lda, o.ldx, o.bad,
		o.tay, o.lda, o.tax, o.bad, o.ldy, o.lda, o.ldx, o.bad,
		o.bcs, o.lda, o.bad, o.bad, o.ldy, o.lda, o.ldx, o.bad,
		o.clv, o.lda, o.tsx, o.bad, o.ldy, o.lda, o.ldx, o.bad,
		o.cpy, o.cmp, o.bad, o.bad, o.cpy, o.cmp, o.dec, o.bad,
		o.iny, o.cmp, o.dex, o.bad, o.cpy, o.cmp, o.dec, o.bad,
		o.bne, o.cmp, o.bad, o.bad, o.bad, o.cmp, o.dec, o.bad,
		o.cld, o.cmp, o.bad, o.bad, o.bad, o.cmp, o.dec, o.bad,
		o.cpx, o.sbc, o.bad, o.bad, o.cpx, o.sbc, o.inc, o.bad,
		o.inx, o.sbc, o.nop, o.bad, o.cpx, o.sbc, o.inc, o.bad,
		o.beq, o.sbc, o.bad, o.bad, o.bad, o.sbc, o.inc, o.bad,
		o.sed, o.sbc, o.bad, o.bad, o.bad, o.sbc, o.inc, o.bad,
	}
}

func (o *CPU) bad(c *stepContext) {
	//panic(c)
	c.a = c.pc
}

// A,Z,C,N = A+M+C
func (o *CPU) adc(ctx *stepContext) {
	a := o.A
	b := o.Read(ctx.a)
	c := o.C
	o.A = a + b + c
	o.SetZN(o.A)
	if int(a)+int(b)+int(c) > 0xFF {
		o.C = 1
	} else {
		o.C = 0
	}
	if (a^b)&0x80 == 0 && (a^o.A)&0x80 != 0 {
		o.V = 1
	} else {
		o.V = 0
	}
}

// A,Z,N = A & M
func (o *CPU) and(ctx *stepContext) {
	o.A &= o.Read(ctx.a)
	o.SetZN(o.A)
}

// A,Z,C,N = M*2 or M,Z,C,N = M*2
func (o *CPU) asl(ctx *stepContext) {
	if ctx.mode == amAccumulator {
		o.C = o.A >> 7 & 1
		o.A <<= 1
		o.SetZN(o.A)
	} else {
		V := o.Read(ctx.a)
		o.C = V >> 7 & 1
		V <<= 1
		o.Write(ctx.a, V)
		o.SetZN(V)
	}
}

// branch on C == 0
func (o *CPU) bcc(ctx *stepContext) {
	if o.C == 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

// branch on C == 1
func (o *CPU) bcs(ctx *stepContext) {
	if o.C != 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) beq(ctx *stepContext) {
	if o.Z != 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

// A & M, N = M7, V = M6
func (o *CPU) bit(ctx *stepContext) {
	V := o.Read(ctx.a)
	o.V = V >> 6 & 1
	o.SetZ(V & o.A)
	o.SetN(V)
}

func (o *CPU) bmi(ctx *stepContext) {
	if o.N != 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) bne(ctx *stepContext) {
	if o.Z == 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) bpl(ctx *stepContext) {
	if o.N == 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) brk(ctx *stepContext) {
	o.push16(o.PC)
	o.php(ctx)
	o.sei(ctx)
	o.PC = o.Read16(0xFFFE)
}

func (o *CPU) bvc(ctx *stepContext) {
	if o.V == 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) bvs(ctx *stepContext) {
	if o.V != 0 {
		o.PC = ctx.a
		o.addBranchCycles(ctx)
	}
}

func (o *CPU) clc(ctx *stepContext) {
	o.C = 0
}

func (o *CPU) cld(ctx *stepContext) {
	o.D = 0
}

func (o *CPU) cli(ctx *stepContext) {
	o.I = 0
}

func (o *CPU) clv(ctx *stepContext) {
	o.V = 0
}

func (o *CPU) cmp(ctx *stepContext) {
	v := o.Read(ctx.a)
	o.compare(o.A, v)
}

func (o *CPU) cpx(ctx *stepContext) {
	v := o.Read(ctx.a)
	o.compare(o.X, v)
}

func (o *CPU) cpy(ctx *stepContext) {
	v := o.Read(ctx.a)
	o.compare(o.Y, v)
}

func (o *CPU) dec(ctx *stepContext) {
	v := o.Read(ctx.a) - 1
	o.Write(ctx.a, v)
	o.SetZN(v)
}

func (o *CPU) dex(c *stepContext) {
	o.X--
	o.SetZN(o.X)
}

func (o *CPU) dey(c *stepContext) {
	o.Y--
	o.SetZN(o.Y)
}

func (o *CPU) eor(ctx *stepContext) {
	o.A ^= o.Read(ctx.a)
	o.SetZN(o.A)
}

func (o *CPU) inc(ctx *stepContext) {
	v := o.Read(ctx.a) + 1
	o.Write(ctx.a, v)
	o.SetZN(v)
}

func (o *CPU) inx(c *stepContext) {
	o.X++
	o.SetZN(o.X)
}

func (o *CPU) iny(c *stepContext) {
	o.Y++
	o.SetZN(o.Y)
}

func (o *CPU) jmp(ctx *stepContext) {
	o.PC = ctx.a
}

func (o *CPU) jsr(ctx *stepContext) {
	o.push16(o.PC - 1)
	o.PC = ctx.a
}

func (o *CPU) lda(ctx *stepContext) {
	o.A = o.Read(ctx.a)
	o.SetZN(o.A)
}

func (o *CPU) ldx(ctx *stepContext) {
	o.X = o.Read(ctx.a)
	o.SetZN(o.X)
}

func (o *CPU) ldy(ctx *stepContext) {
	o.Y = o.Read(ctx.a)
	o.SetZN(o.Y)
}

func (o *CPU) lsr(ctx *stepContext) {
	if ctx.mode == amAccumulator {
		o.C = o.A & 1
		o.A >>= 1
		o.SetZN(o.A)
	} else {
		v := o.Read(ctx.a)
		o.C = v & 1
		v >>= 1
		o.Write(ctx.a, v)
		o.SetZN(v)
	}
}

func (o *CPU) nop(c *stepContext) {

}

func (o *CPU) ora(ctx *stepContext) {
	o.A |= o.Read(ctx.a)
	o.SetZN(o.A)
}

func (o *CPU) pha(c *stepContext) {
	o.push(o.A)
}

func (o *CPU) php(c *stepContext) {
	o.push(o.GetFlags() | 0x10)
}

func (o *CPU) pla(c *stepContext) {
	o.A = o.pop()
	o.SetZN(o.A)
}

func (o *CPU) plp(c *stepContext) {
	o.SetFlags(o.pop()&0xEF | 0x20)
}

func (o *CPU) rol(ctx *stepContext) {
	if ctx.mode == amAccumulator {
		c := o.C
		o.C = (o.A >> 7) & 1
		o.A = (o.A << 1) | c
		o.SetZN(o.A)
	} else {
		c := o.C
		v := o.Read(ctx.a)
		o.C = (v >> 7) & 1
		v = (v << 1) | c
		o.Write(ctx.a, v)
		o.SetZN(v)
	}
}

func (o *CPU) ror(ctx *stepContext) {
	if ctx.mode == amAccumulator {
		c := o.C
		o.C = o.A & 1
		o.A = o.A>>1 | c<<7
		o.SetZN(o.A)
	} else {
		c := o.C
		v := o.Read(ctx.a)
		o.C = v & 1
		v = v>>1 | c<<7
		o.Write(ctx.a, v)
		o.SetZN(v)
	}
}

func (o *CPU) rti(ctx *stepContext) {
	o.SetFlags(o.pop()&0xEF | 0x20)
	o.PC = o.pop16()
}

func (o *CPU) rts(ctx *stepContext) {
	o.PC = o.pop16() + 1
}

func (o *CPU) sbc(ctx *stepContext) {
	a := o.A
	b := o.Read(ctx.a)
	c := o.C
	o.A = a - b - (1 - c)
	o.SetZN(o.A)
	if int(a)-int(b)-int(1-c) >= 0 {
		o.C = 1
	} else {
		o.C = 0
	}
	if (a^b)&0x80 != 0 && (a^o.A)&0x80 != 0 {
		o.V = 1
	} else {
		o.V = 0
	}
}

func (o *CPU) sec(ctx *stepContext) {
	o.C = 1
}

func (o *CPU) sed(ctx *stepContext) {
	o.D = 1
}

func (o *CPU) sei(ctx *stepContext) {
	o.I = 1
}

func (o *CPU) sta(ctx *stepContext) {
	o.Write(ctx.a, o.A)
}

func (o *CPU) stx(ctx *stepContext) {
	o.Write(ctx.a, o.X)
}

func (o *CPU) sty(ctx *stepContext) {
	o.Write(ctx.a, o.Y)
}

func (o *CPU) tax(c *stepContext) {
	o.X = o.A
	o.SetZN(o.X)
}

func (o *CPU) tay(c *stepContext) {
	o.Y = o.A
	o.SetZN(o.Y)
}

func (o *CPU) tsx(c *stepContext) {
	o.X = o.SP
	o.SetZN(o.X)
}

func (o *CPU) txa(c *stepContext) {
	o.A = o.X
	o.SetZN(o.A)
}

func (o *CPU) txs(c *stepContext) {
	o.SP = o.X
}

func (o *CPU) tya(c *stepContext) {
	o.A = o.Y
	o.SetZN(o.A)
}
