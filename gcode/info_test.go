package gcode

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	var info Info
	info.Init()
	assert := assert.New(t)
	t.Run("Init", func(t *testing.T) {
		assert.EqualValues(math.MaxFloat32, info.Z.Min)
		assert.EqualValues(-math.MaxFloat32, info.Z.Max)
	})

	t.Run("UpdateX", func(t *testing.T) {
		info.X.Update(1.0)
		assert.EqualValues(1.0, info.X.Max)
		info.X.Update(-1.0)
		assert.EqualValues(-1.0, info.X.Min)
	})

	t.Run("UpdateY", func(t *testing.T) {
		info.Y.Update(1.0)
		assert.EqualValues(1.0, info.Y.Max)
		info.Y.Update(-1.0)
		assert.EqualValues(-1.0, info.Y.Min)
	})

	t.Run("UpdateZ", func(t *testing.T) {
		info.Z.Update(1.0)
		assert.EqualValues(1.0, info.Z.Max)
		info.Z.Update(-1.0)
		assert.EqualValues(-1.0, info.Z.Min)
	})

	info.Increment = -3.0
	info.MinCut = 0.5
	info.Z.Max = 5.0
	info.Z.Min = -8.25
	t.Run("Passes", func(t *testing.T) {
		passes := info.Passes()
		assert.EqualValues(3, passes)
	})
}
