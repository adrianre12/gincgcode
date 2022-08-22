package gcode

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestRuneScanner(t *testing.T) {
	assert := assert.New(t)

	t.Run("Empty", func(t *testing.T) {
		rs := NewRunesScanner("")
		assert.False(rs.Scan(), "Empty")
	})
	t.Run("Not Empty", func(t *testing.T) {
		rs := NewRunesScanner("%")
		assert.True(rs.Scan(), "Not Empty")
		r := rs.Rune()
		assert.EqualValues("%", string(r), "Wrong rune")
	})
	t.Run("EOL", func(t *testing.T) {
		rs := NewRunesScanner("%")
		rs.Rune()
		assert.False(rs.Scan(), "EOL")
	})
	t.Run("Float+", func(t *testing.T) {
		rs := NewRunesScanner("1.1")
		v, err := rs.GetValue(ValueFloat)
		require.Emptyf(t, err, "failed to parse: %s", err)
		assert.EqualValues(1.1, v, "Wrong value")
	})
	t.Run("Float-", func(t *testing.T) {
		rs := NewRunesScanner("-1.1")
		v, err := rs.GetValue(ValueFloat)
		require.Emptyf(t, err, "failed to parse: %s", err)
		assert.EqualValues(-1.1, v, "Wrong value")
	})
	t.Run("Int+", func(t *testing.T) {
		rs := NewRunesScanner("1")
		v, err := rs.GetValue(ValueInt)
		require.Emptyf(t, err, "failed to parse: %s", err)
		assert.EqualValues(1, v, "Wrong value")
	})
	t.Run("Int-", func(t *testing.T) {
		rs := NewRunesScanner("-1")
		v, err := rs.GetValue(ValueInt)
		require.Emptyf(t, err, "failed to parse: %s", err)
		assert.EqualValues(-1, v, "Wrong value")
	})
	t.Run("Malformed Int", func(t *testing.T) {
		rs := NewRunesScanner("+-1")
		_, err := rs.GetValue(ValueInt)
		require.NotEmptyf(t, err, "failed to parse: %s", err)
	})
	t.Run("Malformed Float", func(t *testing.T) {
		rs := NewRunesScanner("1.1.")
		_, err := rs.GetValue(ValueFloat)
		require.NotEmptyf(t, err, "failed to parse: %s", err)
	})
	t.Run("Multi cmd", func(t *testing.T) {
		rs := NewRunesScanner("X1.1M3")
		rs.Rune() // move one rune
		v1, err := rs.GetValue(ValueFloat)
		require.Emptyf(t, err, "failed to parse: %s", err)
		rs.Rune()
		v2, err := rs.GetValue(ValueInt)
		require.Emptyf(t, err, "failed to parse: %s", err)
		assert.EqualValues(1.1, v1, "Wrong value")
		assert.EqualValues(3, v2, "Wrong value")
	})

	t.Run("Until EOL", func(t *testing.T) {
		rs := NewRunesScanner("(abcdef)")
		rs.Rune()
		runes := rs.Until(false, ' ')
		assert.EqualValues("abcdef)", string(*runes), "Until EOL")
	})
	t.Run("Until )", func(t *testing.T) {
		rs := NewRunesScanner("(abcdef)")
		rs.Rune()
		runes := rs.Until(true, ')')
		assert.EqualValues("abcdef", string(*runes), "Until )")
	})
}
