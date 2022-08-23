// core
package main

import (
	"bufio"
	"fmt"
	"gincgcode/gcode"
	"math"
	"os"
	"strings"

	"github.com/adrianre12/logl"
)

func ReadFile(fileName string) *gcode.Blocks {
	fileIn, err := os.Open(fileName)
	if err != nil {
		logl.Fatalf("Failed to open file: %s", err)
	}
	defer fileIn.Close()

	scanner := bufio.NewScanner(fileIn)
	blocks := make(gcode.Blocks, 0)

	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if len(txt) == 0 { // ignore blank lines
			continue
		}

		block, err := gcode.ParseLine(txt)
		if err != nil {
			logl.Fatalf("Failed to parse '%s' %s", txt, err)
		}
		blocks = append(blocks, block)
	}

	if scanner.Err() != nil {
		logl.Fatalf("Failed to read file: %s", scanner.Err())
	}
	logl.Debugf("loaded %d blocks", len(blocks))
	return &blocks
}

func OutputBlock(writer *bufio.Writer, block *gcode.Block) {
	writer.WriteString(block.String(true, " "))
}

func OutputBlocks(writer *bufio.Writer, blocks gcode.Blocks) {
	for _, block := range blocks {
		OutputBlock(writer, block)
	}
}

type Current struct {
	Y        float32
	Z        float32
	LastPass int
}

func (c *Current) Update(block gcode.Block) {
	if block.Y != nil && (*block.Y).Value != c.Y {
		c.Y = (*block.Y).Value
	}
	if block.Z != nil && (*block.Z).Value != c.Z {
		c.Z = (*block.Z).Value
	}
	if block.IsClamped { //if it is clamped then LastPass is valid
		c.LastPass = block.LastPass
	}
}

func NewCurrent() Current {
	return Current{Y: float32(math.MaxFloat32), Z: float32(math.MaxFloat32)}
}

func Process(writer *bufio.Writer, info gcode.Info) {
	logl.Debug("Processing")

	passes := info.Passes() // calculate passes
	logl.Debugf("Passes=%d", passes)

	for pass := 1; pass <= passes; pass++ {
		logl.Debugf("======================== Pass %d =============================", pass)
		writer.WriteString(fmt.Sprintf(";Pass %d\n", pass))

		if pass == passes { //last pass finish cut
			OutputBlocks(writer, info.Data)
			continue
		}

		current := NewCurrent() //current tool position
		last := NewCurrent()
		var lastBlock gcode.Block
		lastBlock.Init()

		safeHeight := false
		index := 0
		for index < len(info.Data) { //blocks
			block := info.Data[index]
			logl.Debugf("%d %s", index, block.String(false, " "))

			last = current
			clampedBlock := block.Copy() //copy the block
			clampedBlock.ToStepZ(&info, pass)
			current.Update(clampedBlock)

			if clampedBlock.IsClamped {
				logl.Debugf("Z clamped to %.3f", (*clampedBlock.Z).Value)
			}
			logl.Debugf("Current Y=%.3f Z=%.3f LastPass = %d", current.Y, current.Z, current.LastPass)
			skip := false
			if clampedBlock.NoChangeZ(last.Z) {
				skip = true
			} else {
				if safeHeight && current.LastPass < pass {
					skip = true
				}
			}
			skip = skip && clampedBlock.NoChangeY(last.Y)
			logl.Debugf("Skip = %t", skip)
			if skip {
				logl.Debugf("skip %d", index)
				index++
				if index == len(info.Data) {
					logl.Debug("Output lastBlock as it is end of data")
					OutputBlock(writer, &clampedBlock)
				} else {
					if logl.GetLevel() == logl.DEBUG {
						writer.WriteString(";skip ")
						OutputBlock(writer, &clampedBlock)
					}
				}
				if !lastBlock.IsSkip && current.LastPass < pass { //starting to skip, move to safe
					logl.Debug("Starting skip move to safe")
					writer.WriteString(fmt.Sprintf("G00 Z%.3f ;fast to safe\n", info.Safe))
					safeHeight = true
				}
				lastBlock = clampedBlock.Copy()
				lastBlock.SetZ(current.Z)
				lastBlock.LastPass = current.LastPass
				lastBlock.IsSkip = true

			} else { // Y or Z moved
				if lastBlock.IsSkip {
					if lastBlock.LastPass < pass {
						logl.Debug("Output fast lastBlock and slow to depth")
						lastZ := lastBlock.Z.Value
						lastBlock.SetZ(info.Safe)
						lastBlock.SetG(0)
						OutputBlock(writer, &lastBlock)
						writer.WriteString(fmt.Sprintf("G01 Z%.3f ;slow to depth\n", lastZ))
						safeHeight = false
					} else {
						logl.Debug("Output lastBlock")
						OutputBlock(writer, &lastBlock)
					}
					lastBlock.Init()
				}

				logl.Debugf("Output %d", index)
				OutputBlock(writer, &clampedBlock)
				if lastBlock.IsSkip && lastBlock.LastPass < pass { //point is from shallower pass
					writer.WriteString(fmt.Sprintf("G00 Z%.3f ;fast to safe after change\n", info.Safe))
					safeHeight = true
				}
				index++
			}

		}
	}

}

func Run(cli *CliType) error {
	logl.Info("Starting")
	var fout *os.File

	if cli.Outfile == "" {
		logl.Debug("Output to screen")
		fout = os.Stdout
	} else {
		logl.Debug("Output to file")
		var err error
		fout, err = os.OpenFile(cli.Outfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logl.Fatal("Error opening file")
		}
	}
	writer := bufio.NewWriter(fout)
	defer writer.Flush()
	blocks := ReadFile(cli.Infile)

	info := gcode.FindInfo(blocks)
	info.Increment = cli.Increment
	info.MinCut = cli.MinCut
	info.Safe = cli.Safe

	logl.Debugf("MinZ=%.3f MaxZ=%.3f Increment=%.3f minCut=%.3f safe=%.3f", info.MinZ, info.MaxZ, info.Increment, info.MinCut, info.Safe)

	OutputBlocks(writer, info.Setup)
	Process(writer, info)
	OutputBlocks(writer, info.Finish)
	return nil
}
