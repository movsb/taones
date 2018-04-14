package main

// 状态标志位（Status Bits）
const (
	_C byte = 1 << iota // 进位（Carry）
	_Z                  // 零（Zero）
	_I                  // 中断控制
	_D                  // 十进制模式
	_B                  // 中断
	_                   // 未使用
	_V                  // 溢出
	_N                  // 负数
)

// 寻址模式（Addressing Modes）
const (
	_                 byte = iota
	amImmediate            // 1  立即寻址
	amZero                 // 2  零页索引寻址
	amZeroX                // 3  零页直接寻址
	amZeroY                // 4  零页直接寻址
	amAbsolute             // 5  绝对寻址
	amAbsoluteX            // 6  绝对X寻址
	amAbsoluteY            // 7  绝对Y寻址
	amIndirect             // 8  间接寻址
	amIndexedIndirect      // 9  先索引后间接
	amIndirectIndexed      // 10 先间接后索引
	amRelative             // 11 相对寻址
	amImplied              // 12 隐含寻址
	amAccumulator          // 13 累加器寻址
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
	1, 2, 0, 0, 0, 2, 2, 0, 1, 2, 1, 0, 0, 3, 3, 0,
	2, 2, 0, 0, 0, 2, 2, 0, 1, 3, 0, 0, 0, 3, 3, 0,
	3, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 0, 3, 2, 0, 1, 2, 0, 0, 0, 3, 3, 0,
	1, 2, 0, 0, 0, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 0, 2, 2, 0, 1, 3, 0, 0, 0, 3, 3, 0,
	1, 2, 0, 0, 0, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 0, 2, 2, 0, 1, 3, 0, 0, 0, 3, 3, 0,
	0, 3, 0, 0, 2, 3, 2, 0, 1, 0, 1, 0, 3, 3, 3, 0,
	2, 3, 0, 0, 2, 3, 2, 0, 1, 3, 1, 0, 0, 3, 0, 0,
	2, 2, 2, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 0, 2, 2, 0, 1, 3, 0, 0, 0, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 0, 2, 2, 0, 1, 3, 0, 0, 0, 3, 3, 0,
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
	7, 6, 0, 0, 0, 3, 5, 0, 3, 2, 3, 0, 0, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
	6, 6, 0, 0, 3, 3, 5, 0, 4, 2, 2, 0, 4, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
	6, 6, 0, 0, 0, 3, 5, 0, 3, 2, 2, 0, 3, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
	6, 6, 0, 0, 0, 3, 5, 0, 4, 3, 2, 0, 5, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
	0, 5, 0, 0, 3, 5, 3, 0, 2, 0, 2, 0, 4, 5, 4, 0,
	2, 5, 0, 0, 4, 5, 4, 0, 2, 5, 2, 0, 0, 5, 0, 0,
	2, 6, 2, 0, 3, 3, 3, 0, 2, 2, 2, 0, 4, 4, 4, 0,
	2, 5, 0, 0, 4, 4, 4, 0, 2, 4, 2, 0, 4, 4, 4, 0,
	2, 6, 0, 0, 3, 3, 5, 0, 2, 2, 2, 0, 4, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
	2, 6, 0, 0, 3, 3, 5, 0, 2, 2, 2, 0, 4, 4, 6, 0,
	2, 5, 0, 0, 0, 4, 6, 0, 2, 4, 0, 0, 0, 4, 7, 0,
}

type stepContext struct {
	a    uint16
	pc   uint16
	mode byte
}

type opCodeFunc func(*stepContext)

var opcodeModes = [256]byte{}

type CPU struct {
	A  byte // 累加寄存器（Accumulator）
	X  byte // 索引寄存器（Index Register）
	Y  byte // 索引寄存器（Index Register）
	SP byte // 栈指针（Stack Pointer）
	F  byte // 状态寄存器（Status Register）
}

func (o *CPU) bad(c *stepContext) {

}

func (o *CPU) adc(c *stepContext) {

}

func (o *CPU) and(c *stepContext) {

}

func (o *CPU) asl(c *stepContext) {

}

func (o *CPU) bcc(c *stepContext) {

}

func (o *CPU) bcs(c *stepContext) {

}

func (o *CPU) beq(c *stepContext) {

}

func (o *CPU) bit(c *stepContext) {

}

func (o *CPU) bmi(c *stepContext) {

}

func (o *CPU) bne(c *stepContext) {

}

func (o *CPU) bpl(c *stepContext) {

}

func (o *CPU) brk(c *stepContext) {

}

func (o *CPU) bvc(c *stepContext) {

}

func (o *CPU) bvs(c *stepContext) {

}

func (o *CPU) clc(c *stepContext) {

}

func (o *CPU) cld(c *stepContext) {

}

func (o *CPU) cli(c *stepContext) {

}

func (o *CPU) clv(c *stepContext) {

}

func (o *CPU) cmp(c *stepContext) {

}

func (o *CPU) cpx(c *stepContext) {

}

func (o *CPU) cpy(c *stepContext) {

}

func (o *CPU) dec(c *stepContext) {

}

func (o *CPU) dex(c *stepContext) {

}

func (o *CPU) dey(c *stepContext) {

}

func (o *CPU) eor(c *stepContext) {

}

func (o *CPU) inc(c *stepContext) {

}

func (o *CPU) int(c *stepContext) {

}

func (o *CPU) inx(c *stepContext) {

}

func (o *CPU) iny(c *stepContext) {

}

func (o *CPU) jmp(c *stepContext) {

}

func (o *CPU) jsr(c *stepContext) {

}

func (o *CPU) lda(c *stepContext) {

}

func (o *CPU) ldx(c *stepContext) {

}

func (o *CPU) ldy(c *stepContext) {

}

func (o *CPU) lsr(c *stepContext) {

}

func (o *CPU) nop(c *stepContext) {

}

func (o *CPU) ora(c *stepContext) {

}

func (o *CPU) pha(c *stepContext) {

}

func (o *CPU) php(c *stepContext) {

}

func (o *CPU) pla(c *stepContext) {

}

func (o *CPU) plp(c *stepContext) {

}

func (o *CPU) rol(c *stepContext) {

}

func (o *CPU) ror(c *stepContext) {

}

func (o *CPU) rti(c *stepContext) {

}

func (o *CPU) rts(c *stepContext) {

}

func (o *CPU) sbc(c *stepContext) {

}

func (o *CPU) sec(c *stepContext) {

}

func (o *CPU) sed(c *stepContext) {

}

func (o *CPU) sei(c *stepContext) {

}

func (o *CPU) sta(c *stepContext) {

}

func (o *CPU) stx(c *stepContext) {

}

func (o *CPU) sty(c *stepContext) {

}

func (o *CPU) tax(c *stepContext) {

}

func (o *CPU) tay(c *stepContext) {

}

func (o *CPU) tsx(c *stepContext) {

}

func (o *CPU) txa(c *stepContext) {

}

func (o *CPU) txs(c *stepContext) {

}

func (o *CPU) tya(c *stepContext) {

}
