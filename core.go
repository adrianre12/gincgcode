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

func OutputBlock(writer *bufio.Writer, block *gcode.Block, pretty bool) {
	writer.WriteString(block.String(true, pretty))
}

func OutputBlocks(writer *bufio.Writer, blocks gcode.Blocks, pretty bool) {
	for _, block := range blocks {
		OutputBlock(writer, block, pretty)
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

func TernaryString(condition bool, strTrue string, strFalse string) string {
	if condition {
		return strTrue
	}
	return strFalse
}

func Process(writer *bufio.Writer, info gcode.Info) {
	passes := info.Passes() // calculate passes
	logl.Infof("Passes=%d", passes)

	for pass := 1; pass <= passes; pass++ {
		logl.Debugf("======================== Pass %d =============================", pass)
		writer.WriteString(fmt.Sprintf(";Pass %d\n", pass))

		if pass == passes { //last pass finish cut
			OutputBlocks(writer, info.Data, info.Pretty)
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
			logl.Debugf("%d %s", index, block.String(false, true))

			last = current
			clampedBlock := block.Copy() //copy the block
			clampedBlock.ToStepZ(&info, pass)
			current.Update(clampedBlock)

			if clampedBlock.IsClamped {
				logl.Debugf("Z clamped to %.3f", (*clampedBlock.Z).Value)
			}
			logl.Debugf("Current Y=%.3f Z=%.3f LastPass = %d", current.Y, current.Z, current.LastPass)

			if info.FeedRate > 0 && clampedBlock.F != nil {
				clampedBlock.SetF(info.FeedRate)
			}

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
					OutputBlock(writer, &clampedBlock, info.Pretty)
				} else {
					if logl.GetLevel() == logl.DEBUG {
						writer.WriteString(";skip ")
						OutputBlock(writer, &clampedBlock, info.Pretty)
					}
				}
				if !lastBlock.IsSkip && current.LastPass < pass { //starting to skip, move to skip height
					logl.Debug("Starting skip move to skip height")
					writer.WriteString(fmt.Sprintf("G00 Z%.3f%s\n", info.SkipHeight, TernaryString(info.Pretty, " ;fast to skip height", "")))
					safeHeight = true
				}
				lastBlock = clampedBlock.Copy()
				lastBlock.LastPass = current.LastPass
				lastBlock.IsSkip = true

			} else { // Y or Z moved
				if lastBlock.IsSkip {
					if lastBlock.LastPass < pass {
						logl.Debug("Output fast lastBlock and slow to depth")
						lastZ := last.Z
						lastBlock.SetZ(info.SkipHeight)
						lastBlock.SetG(0)
						OutputBlock(writer, &lastBlock, info.Pretty)
						writer.WriteString(fmt.Sprintf("G01 Z%.3f%s\n", lastZ, TernaryString(info.Pretty, " ;slow to depth", "")))
						safeHeight = false
					} else {
						logl.Debug("Output lastBlock")
						OutputBlock(writer, &lastBlock, info.Pretty)
					}
					lastBlock.Init()
				}

				logl.Debugf("Output %d", index)
				OutputBlock(writer, &clampedBlock, info.Pretty)
				if lastBlock.IsSkip && lastBlock.LastPass < pass { //point is from shallower pass
					writer.WriteString(fmt.Sprintf("G00 Z%.3f%s\n", info.SkipHeight, TernaryString(info.Pretty, " ;fast to skip height after change", "")))
					safeHeight = true
				}
				index++
			}

		}
	}

}

func Realign(info *gcode.Info, alignment string) {

	var offsetX float32 // to be added to positions
	var offsetY float32

	switch strings.ToLower(alignment) {
	case "none":
		{
			offsetX = 0
			offsetY = 0
		}
	case "corner":
		{
			offsetX = -info.X.Min
			offsetY = -info.Y.Min
		}
	case "center":
		{
			offsetX = -(info.X.Max + info.X.Min) / 2
			offsetY = -(info.Y.Max + info.Y.Min) / 2
		}
	default:
		{
			logl.Fatalf("Invalid Alignment %s", alignment)
		}
	}

	for _, block := range info.Data {
		block.Reposition(offsetX, offsetY)
	}

	info.X.Min = info.X.Min + offsetX
	info.X.Max = info.X.Max + offsetX
	info.Y.Min = info.Y.Min + offsetY
	info.Y.Max = info.Y.Max + offsetY
}

func Run(cli *CliType) error {
	logl.Info("Starting")
	var fout *os.File

	if cli.Outfile == "" {
		logl.Info("Output to screen")
		fout = os.Stdout
	} else {
		logl.Info("Output to file")
		var err error
		fout, err = os.OpenFile(cli.Outfile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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
	info.SkipHeight = cli.SkipHeight
	if cli.Feed <= 0 {
		logl.Fatal("Feed cannot be zero or negative")
	}
	info.FeedRate = cli.Feed
	info.Pretty = cli.Pretty

	Realign(&info, cli.Align)

	logl.Infof("MinX=%.3f MaxX=%.3f MinY=%.3f MaxY=%.3f MinZ=%.3f MaxZ=%.3f", info.X.Min, info.X.Max, info.Y.Min, info.Y.Max, info.Z.Min, info.Z.Max)
	logl.Infof("Increment=%.3f minCut=%.3f skipHeight=%.3f feedRate=%.0f", info.Increment, info.MinCut, info.SkipHeight, info.FeedRate)

	OutputBlocks(writer, info.Setup, info.Pretty)
	Process(writer, info)
	OutputBlocks(writer, info.Finish, info.Pretty)
	logl.Info("Finished")

	return nil
}
