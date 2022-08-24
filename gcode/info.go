package gcode

import (
	"math"

	"github.com/adrianre12/logl"
)

type Info struct {
	Setup      Blocks
	Data       Blocks
	Finish     Blocks
	MinZ       float32
	MaxZ       float32
	Increment  float32
	MinCut     float32
	SkipHeight float32
	FeedRate   float32
	Pretty     bool
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

func FindInfo(blocks *Blocks) Info {
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
	info.Setup = (*blocks)[:first]
	info.Data = (*blocks)[first : last+1]
	info.Finish = (*blocks)[last+1:]
	return info
}
