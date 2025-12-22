package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
)

func main() {
	var gym bool
	var cpu bool

	rootCmd := &cobra.Command{
		Use: "ants-again",
		RunE: func(cmd *cobra.Command, args []string) error {
			ebiten.SetTPS(TPS)

			if cpu {
				// Start CPU profiling
				f, err := os.Create("cpu.prof")
				if err != nil {
					log.Fatal(err)
				}
				pprof.StartCPUProfile(f)
				defer func() {
					pprof.StopCPUProfile()
					f.Close()
				}()
			}

			if gym {
				return runGym()
			}

			var params *Params
			game := NewGame(params)
			ebiten.SetWindowSize(800, 800)
			ebiten.SetWindowTitle("Hello, World!")
			if err := ebiten.RunGame(game); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVar(&gym, "gym", false, "Enable gym mode")
	rootCmd.Flags().BoolVar(&cpu, "cpu", false, "Enable CPU mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
