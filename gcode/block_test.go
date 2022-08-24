package gcode

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test1ParseBlock(t *testing.T) {
	passTests := map[string]struct {
		cmd   string
		value float32
		ctype CmdType
	}{
		"%":     {cmd: "%", value: 0, ctype: Percent},
		"F250":  {cmd: "F", value: 250, ctype: ValueInt},
		"G1":    {cmd: "G", value: 1, ctype: Address},
		"G01":   {cmd: "G", value: 1, ctype: Address},
		"S1000": {cmd: "S", value: 1000, ctype: ValueInt},
		"M3":    {cmd: "M", value: 3, ctype: Address},
		"X-1.1": {cmd: "X", value: -1.1, ctype: ValueFloat},
		"Y-1.1": {cmd: "Y", value: -1.1, ctype: ValueFloat},
		"Z1.1":  {cmd: "Z", value: 1.1, ctype: ValueFloat},
		"/bla":  {cmd: "/bla", value: 0, ctype: Comment},
	}

	assert := assert.New(t)

	for name, tc := range passTests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseLine(name)
			require.Emptyf(t, err, "failed to parse '%s': %s", name, err)
			assert.EqualValues(1, len((*got).Cmds), "result length is not 1")
			cc := ((*got).Cmds)[0]
			assert.EqualValues(tc.cmd, cc.Cmd, "Wrong Cmd")
			assert.EqualValues(tc.value, cc.Value, "Wrong Value")
			assert.EqualValues(tc.ctype, cc.Type, "Wrong Type")
		})
	}
}

func Test2ParseBlock(t *testing.T) {
	assert := assert.New(t)
	got, err := ParseLine("G00 X1.500 Y-1.500 Z5.000 S20000 M3")
	require.Emptyf(t, err, "failed to parse: %s", err)
	assert.EqualValues(6, len((*got).Cmds), "result length is not 2")
	cc := ((*got).Cmds)[0]
	assert.EqualValues("G", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(0, cc.Value, "Wrong Value")
	assert.EqualValues(Address, cc.Type, "Not Address")

	cc = ((*got).Cmds)[1]
	assert.EqualValues("X", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(1.5, cc.Value, "Wrong Value")
	assert.EqualValues(ValueFloat, cc.Type, "Not Value")

	cc = ((*got).Cmds)[2]
	assert.EqualValues("Y", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(-1.5, cc.Value, "Wrong Value")
	assert.EqualValues(ValueFloat, cc.Type, "Not Value")

	cc = ((*got).Cmds)[3]
	assert.EqualValues("Z", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(5.0, cc.Value, "Wrong Value")
	assert.EqualValues(ValueFloat, cc.Type, "Not Value")

	cc = ((*got).Cmds)[4]
	assert.EqualValues("S", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(20000, cc.Value, "Wrong Value")
	assert.EqualValues(ValueInt, cc.Type, "Not Value")

	cc = ((*got).Cmds)[5]
	assert.EqualValues("M", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(3, cc.Value, "Wrong Value")
	assert.EqualValues(Address, cc.Type, "Not Value")

	assert.EqualValues(1.5, (*got.X).Value)
	assert.EqualValues(-1.5, (*got.Y).Value)
	assert.EqualValues(5.0, (*got.Z).Value)

}

func Test3ParseBlock(t *testing.T) {
	assert := assert.New(t)
	got, err := ParseLine("G0(embeded)X1.1;trailing")
	require.Emptyf(t, err, "failed to parse: %s", err)
	assert.EqualValues(4, len((*got).Cmds), "result length is not 4")
	cc := ((*got).Cmds)[0]
	assert.EqualValues("G", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(0, cc.Value, "Wrong Value")
	assert.EqualValues(Address, cc.Type, "Not Address")

	cc = ((*got).Cmds)[1]
	assert.EqualValues("(embeded)", cc.Cmd)
	assert.EqualValues(Comment, cc.Type, "Not Comment")

	cc = ((*got).Cmds)[2]
	assert.EqualValues("X", cc.Cmd, "Wrong Cmd")
	assert.EqualValues(1.1, cc.Value, "Wrong Value")
	assert.EqualValues(ValueFloat, cc.Type, "Not Value")

	cc = ((*got).Cmds)[3]
	assert.EqualValues(";trailing", cc.Cmd)
	assert.EqualValues(Comment, cc.Type, "Not Comment")
}

func Test4ParseBlock(t *testing.T) {
	_, err := ParseLine("G91")
	require.NotEmpty(t, err, "Failed to reject unsuported command")
}

func Test5ParseBlock(t *testing.T) {
	assert := assert.New(t)
	got, err := ParseLine("G0X1.0")
	require.Empty(t, err, "Failed to pass cmd line")
	assert.EqualValues("G00 X1.000 ", got.String(false, true))
	assert.EqualValues("G00 X1.000 \n", got.String(true, true))
}
func Test6ParseBlock(t *testing.T) {
	assert := assert.New(t)
	got, err := ParseLine("G0")
	require.Empty(t, err, "Failed to pass cmd line")

	got.SetX(1.0)
	require.NotEmpty(t, (*got).X, "X not set")
	assert.EqualValues((*got).X.Value, 1.0)

	got.SetY(2.0)
	require.NotEmpty(t, (*got).Y, "Y not set")
	assert.EqualValues((*got).Y.Value, 2.0)

	got.SetZ(3.0)
	require.NotEmpty(t, (*got).Z, "Z not set")
	assert.EqualValues((*got).Z.Value, 3.0)

	assert.EqualValues("G00 X1.000 Y2.000 Z3.000 ", got.String(false, true))
}

func Test7ParseBlock(t *testing.T) {
	assert := assert.New(t)
	block := Block{}
	block.Init()

	//check against nil
	assert.True(block.NoChangeY(2.0))
	assert.True(block.NoChangeZ(3.0))

	block.SetY(2.0)
	block.SetZ(3.0)
	assert.True(block.NoChangeY(2.0))
	assert.True(block.NoChangeZ(3.0))
	assert.False(block.NoChangeY(2.1))
	assert.False(block.NoChangeZ(3.1))

	assert.Empty(block.G)
	block.SetG(0)

	assert.NotEmpty(block.G)
	assert.EqualValues(0, block.G.Value)
	assert.EqualValues(Address, block.G.Type)
	block.SetG(1)
	assert.EqualValues(1, block.G.Value)

}

func Test9ParseBlock(t *testing.T) { // ToStep
	assert := assert.New(t)

	tests := map[string]struct {
		X       float32
		Y       float32
		Z       float32
		inc     float32
		minCut  float32
		pass    int
		safe    float32
		isC     bool
		expZ    float32
		expPass int
	}{
		"pass 1 Z = 0":        {X: 1.0, Y: 2.0, Z: 0, inc: -3.0, minCut: 0.5, pass: 1, safe: 5.0, isC: false, expZ: 0, expPass: 0},
		"pass 1 Z = 3.0":      {X: 1.0, Y: 2.0, Z: 3.0, inc: -3.0, minCut: 0.5, pass: 1, safe: 5.0, isC: false, expZ: 3.0, expPass: 0},
		"pass 1 shallow":      {X: 1.0, Y: 2.0, Z: -2.0, inc: -3.0, minCut: 0.5, pass: 1, safe: 5.0, isC: true, expZ: 0.5, expPass: 0},
		"pass 1 deep":         {X: 1.0, Y: 2.0, Z: -4.0, inc: -3.0, minCut: 0.5, pass: 1, safe: 5.0, isC: true, expZ: -2.5, expPass: 1},
		"pass 1 very shallow": {X: 1.0, Y: 2.0, Z: -2.0, inc: -3.0, minCut: 0.5, pass: 1, safe: 5.0, isC: true, expZ: 0.5, expPass: 0},
		"pass 2 shallow":      {X: 1.0, Y: 2.0, Z: -4.0, inc: -3.0, minCut: 0.5, pass: 2, safe: 5.0, isC: true, expZ: -2.5, expPass: 1},
		"pass 2 deep":         {X: 1.0, Y: 2.0, Z: -7.0, inc: -3.0, minCut: 0.5, pass: 2, safe: 5.0, isC: true, expZ: -5.5, expPass: 2},
	}

	t.Run("Nil Z", func(t *testing.T) {
		tc := tests["pass 1 skip"]
		b := Block{}
		b.SetX(tc.X)
		b.SetY(tc.Y)
		info := Info{Increment: tc.inc, MinCut: tc.minCut, SkipHeight: tc.safe}
		b.ToStepZ(&info, tc.pass)
		assert.False(b.IsClamped, "Clamped returned wrong value")
		assert.Empty(b.Z, "Z should be nil")
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b := Block{}
			b.SetX(tc.X)
			b.SetY(tc.Y)
			b.SetZ(tc.Z)
			info := Info{Increment: tc.inc, MinCut: tc.minCut, SkipHeight: tc.safe}
			b.ToStepZ(&info, tc.pass)
			assert.EqualValues(tc.isC, b.IsClamped, "IsClamped")
			assert.EqualValues(tc.expZ, (*b.Z).Value, "Value")
			assert.EqualValues(tc.expPass, b.LastPass, "LastPass")
		})
	}
}
