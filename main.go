package main

func main() {
	console := NewConsole()
	cartridge := LoadROM("smb.nes")
	console.Run(cartridge)
}
