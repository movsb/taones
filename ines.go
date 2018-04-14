package main

import (
	"encoding/binary"
	"io"
	"os"
)

const iNESMagic = 0x1a53454e

type iNESHeader struct {
	Magic    uint32
	NumPRG   byte
	NumCHR   byte
	Control1 byte
	Control2 byte
	_        [8]byte
}

func LoadROM(path string) *Cartridge {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer fp.Close()

	header := iNESHeader{}
	if err := binary.Read(fp, binary.LittleEndian, &header); err != nil {
		panic(err)
	}

	if header.Magic != iNESMagic {
		panic("bad rom")
	}

	mapper1 := header.Control1 >> 4
	mapper2 := header.Control2 >> 4
	mapper := mapper1 | mapper2<<4

	mirror1 := header.Control1 & 1
	mirror2 := header.Control2 >> 3 & 1
	mirror := mirror1 | mirror2<<1

	if header.Control1&4 == 4 {
		trainer := make([]byte, 512)
		if _, err := io.ReadFull(fp, trainer); err != nil {
			panic(err)
		}
	}

	prg := make([]byte, int(header.NumPRG)*16384)
	if _, err := io.ReadFull(fp, prg); err != nil {
		panic(err)
	}

	chr := make([]byte, int(header.NumCHR)*8192)
	if _, err := io.ReadFull(fp, chr); err != nil {
		panic(err)
	}

	if header.NumCHR == 0 {
		chr = make([]byte, 8192)
	}

	return NewCartridge(prg, chr, mapper, mirror)
}
