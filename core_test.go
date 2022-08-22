// core_test.go
package main

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
		assert.EqualValues(math.MaxFloat32, info.MinZ)
		assert.EqualValues(-math.MaxFloat32, info.MaxZ)
	})

	t.Run("UpdateZ", func(t *testing.T) {
		info.UpdateZ(1.0)
		assert.EqualValues(1.0, info.MaxZ)
		info.UpdateZ(-1.0)
		assert.EqualValues(-1.0, info.MinZ)
	})

	info.Increment = -3.0
	info.MinCut = 0.5
	info.MaxZ = 5.0
	info.MinZ = -8.25
	t.Run("Passes", func(t *testing.T) {
		passes := info.Passes()
		assert.EqualValues(3, passes)
	})
}
