// core
package main

import (
	"bufio"
	"fmt"
	"gincgcode/gcode"
	"math"
	"os"

	"github.com/adrianre12/logl"
)

type Info struct {
	Setup     *gcode.Blocks
	Data      *gcode.Blocks
	Finish    *gcode.Blocks
	MinZ      float32
	MaxZ      float32
	Increment float32
	MinCut    float32
	Safe      float32
}

func (i *Info) Init() {
	i.MinZ = math.MaxFloat32
	i.MaxZ = -math.MaxFloat32
}

func (i *Info) Passes() int {
	if i.MinZ > 0 {
		logl.Fatal("MinZ > 0")
	}
	return int(math.Ceil(float64(i.MinZ / i.Increment)))
}

func (i *Info) UpdateZ(z float32) {
	if z > i.MaxZ {
		i.MaxZ = z
	}
	if z < i.MinZ {
		i.MinZ = z
	}
}

func ReadFile(fileName string) *gcode.Blocks {
	fileIn, err := os.Open(fileName)
	if err != nil {
		logl.Fatalf("Failed to open file: %s", err)
	}
	defer fileIn.Close()

	scanner := bufio.NewScanner(fileIn)
	blocks := make(gcode.Blocks, 0)

	for scanner.Scan() {
		block, err := gcode.ParseLine(scanner.Text())
		if err != nil {
			logl.Fatalf("Failed to parse file: %s", err)
		}
		blocks = append(blocks, block)
	}

	if scanner.Err() != nil {
		logl.Fatalf("Failed to read file: %s", scanner.Err())
	}
	logl.Debugf("loaded %d blocks", len(blocks))
	return &blocks
}

func FindInfo(blocks *gcode.Blocks) Info {
	//find first and last gcode blocks
	first := math.MaxInt
	last := 0
	var info Info
	info.Init()

	for i, block := range *blocks {
		if block.HasData {
			if i < first {
				first = i
			}
			if i > last {
				last = i
			}
		}
		if block.Z != nil {
			info.UpdateZ((*block.Z).Value)
		}
	}
	logl.Debugf("first=%d last=%d", first, last)
	s := (*blocks)[:first]
	info.Setup = &s
	d := (*blocks)[first : last+1]
	info.Data = &d
	f := (*blocks)[last+1:]
	info.Finish = &f
	return info
}

func OutputBlock(writer *bufio.Writer, block *gcode.Block) {
	writer.WriteString(block.String(true))
}

func OutputBlocks(writer *bufio.Writer, blocks *gcode.Blocks) {
	for _, block := range *blocks {
		OutputBlock(writer, block)
	}
}

func Process(writer *bufio.Writer, info Info) {
	logl.Debug("Processing")

	// skip := false
	// clamp := true

	// // calculate passes
	passes := info.Passes()
	logl.Debugf("Passes=%d", passes)
	for pass := 1; pass <= passes; pass++ {
		logl.Debugf("Pass=%d", pass)
		writer.WriteString(fmt.Sprintf("/Pass %d\n", pass))
		if pass == passes { //last pass finish cut
			OutputBlocks(writer, info.Data)
			continue
		}

		currentY := float32(math.MaxFloat32)
		currentZ := float32(math.MaxFloat32)
		lastBlock := gcode.Block{}

		i := 0
		for i < len(*info.Data) { //blocks
			block := (*info.Data)[i]
			logl.Debugf("%d %s", i, block.String(false))
			clampedBlock := *block //copy the block
			clampedBlock.ClampZ(info.Increment, info.MinCut, pass, info.Safe)
			if clampedBlock.Y != nil && clampedBlock.Y.Value != currentY {
				currentY = clampedBlock.Y.Value
			}
			if clampedBlock.Z != nil && clampedBlock.Z.Value != currentZ {
				currentZ = clampedBlock.Z.Value
			}
			logl.Debugf("Current Y=%.3f Z=%.3f", currentY, currentZ)

			if clampedBlock.NoChangeY(currentY) && !clampedBlock.NoChangeZ(currentZ) {
				logl.Debugf("skip %d", i)
				i++
				if i == len(*info.Data) {
					logl.Debug("Output lastBlock to to end of data")
					OutputBlock(writer, &clampedBlock)
				} else {
					lastBlock = clampedBlock
					lastBlock.Skip = true
				}
			} else { // Y or Z moved
				if lastBlock.Skip {
					logl.Debugf("Output lastBlock")
					OutputBlock(writer, &lastBlock)
					lastBlock.Init()
				}
				logl.Debugf("Output %d", i)
				OutputBlock(writer, &clampedBlock)
				if clampedBlock.Y != nil {
					currentY = (*clampedBlock.Y).Value
				}
				if clampedBlock.Z != nil {
					currentZ = (*clampedBlock.Z).Value
				}
				i++
				continue
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

	info := FindInfo(blocks)
	info.Increment = cli.Increment
	info.MinCut = cli.MinCut
	info.Safe = cli.Safe

	logl.Debugf("MinZ=%.3f MaxZ=%.3f Increment=%.3f minCut=%.3f safe=%.3f", info.MinZ, info.MaxZ, info.Increment, info.MinCut, info.Safe)

	OutputBlocks(writer, info.Setup)
	Process(writer, info)
	OutputBlocks(writer, info.Finish)
	return nil
}