package main

type Cartridge struct {
	PRG    []byte
	CHR    []byte
	Mapper byte
	Mirror byte
}

func NewCartridge(prg []byte, chr []byte, mapper byte, mirror byte) *Cartridge {
	return &Cartridge{
		PRG:    prg,
		CHR:    chr,
		Mapper: mapper,
		Mirror: mirror,
	}
}
