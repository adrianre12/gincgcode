package gcode

import (
	"math"

	"github.com/adrianre12/logl"
)

type MinMax struct {
	Min float32
	Max float32
}

func (mm *MinMax) Init() {
	mm.Min = math.MaxFloat32
	mm.Max = -math.MaxFloat32
}

func (mm *MinMax) Update(value float32) {
	if value > mm.Max {
		mm.Max = value
	}
	if value < mm.Min {
		mm.Min = value
	}
}

type Info struct {
	Setup      Blocks
	Data       Blocks
	Finish     Blocks
	X          MinMax
	Y          MinMax
	Z          MinMax
	Increment  float32
	MinCut     float32
	SkipHeight float32
	FeedRate   float32
	Pretty     bool
}

func (i *Info) Init() {
	i.X.Init()
	i.Y.Init()
	i.Z.Init()
}

func (i *Info) Passes() int {
	if i.Z.Min > 0 {
		logl.Fatal("MinZ > 0")
	}
	return int(math.Ceil(float64(i.Z.Min / i.Increment)))
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
		if block.X != nil {
			info.X.Update((*block.X).Value)
		}
		if block.Y != nil {
			info.Y.Update((*block.Y).Value)
		}
		if block.Z != nil {
			info.Z.Update((*block.Z).Value)
		}
	}
	logl.Debugf("first=%d last=%d", first, last)
	info.Setup = (*blocks)[:first]
	info.Data = (*blocks)[first : last+1]
	info.Finish = (*blocks)[last+1:]
	return info
}
