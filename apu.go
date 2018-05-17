package main

/*
 $4000~$4013, $4015, $4017

 5个通道：两个脉冲波、一个三角波、一个噪声、一个DPCM
*/

type APU struct {
}

type Triangle struct {
	enabled     bool
	counterLoad byte
	timer       uint16
}
