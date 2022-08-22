// gincgcode project main.go
package main

import (
	"github.com/adrianre12/logl"
	"github.com/alecthomas/kong"
)

type CliType struct {
	Debug     bool    `help:"Enable debug mode."`
	Increment float32 `optional:"" short:"i" default:"-3.0" help:"Increment in depth of cut in each pass"`
	MinCut    float32 `optional:"" short:"m" default:"0.5" help:"Minimum thickness to leave for Finish cut"`
	Safe      float32 `optional:"" short:"s" default:"5.0" help:"Safe height for rapid movement"`
	Infile    string  `arg:"" help:"Input filename"`
	Outfile   string  `arg:"" optional:"" help:"Output filename"`
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
