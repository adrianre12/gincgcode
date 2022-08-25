// gincgcode project main.go
package main

import (
	"github.com/adrianre12/logl"
	"github.com/alecthomas/kong"
)

type CliType struct {
	Debug      bool    `help:"Enable debug mode."`
	Pretty     bool    `short:"p" help:"Enable pretty print, this makes the output much larger"`
	Increment  float32 `optional:"" short:"i" default:"-3.0" help:"Increment in depth of cut in each pass"`
	Feed       float32 `optional:"" short:"f" help: "Feed rate override for incremental passes"`
	MinCut     float32 `optional:"" short:"m" default:"0.5" help:"Minimum thickness to leave for Finish cut"`
	SkipHeight float32 `optional:"" short:"s" default:"1.0" help:"Skip height for rapid movement, should be as low as possible to clear materarial"`
	Infile     string  `arg:"" help:"Input filename"`
	Outfile    string  `arg:"" optional:"" help:"Output filename"`
	Align      string  `short:"a" enum:"none,corner,center" default:none help:"Realign output Gcode"`
}

func main() {
	var cli CliType
	kctx := kong.Parse(&cli)
	if cli.Debug {
		logl.SetLevel(logl.DEBUG)
	}
	logl.Debug("Logging at DEBUG")
	err := Run(&cli)
	kctx.FatalIfErrorf(err)
	logl.Close()
}
