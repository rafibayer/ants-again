package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Start CPU profiling
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer func() {
	// 	pprof.StopCPUProfile()
	// 	f.Close()
	// }()

	game := NewGame(nil)

	ebiten.SetTPS(TPS)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
