package gcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeCmd(t *testing.T) {
	assert := assert.New(t)

	t.Run("String", func(t *testing.T) {
		passTests := map[string]struct {
			cmd    string
			value  float32
			ctype  CmdType
			pretty bool
			result string
		}{
			"%":            {cmd: "%", value: 0, ctype: Percent, pretty: false, result: "%"},
			"F250":         {cmd: "F", value: 250, ctype: ValueInt, pretty: false, result: "F250"},
			"G1":           {cmd: "G", value: 1, ctype: Address, pretty: false, result: "G1"},
			"X-1.1":        {cmd: "X", value: -1.1, ctype: ValueFloat, pretty: false, result: "X-1.1"},
			"X-1.0":        {cmd: "X", value: -1.0, ctype: ValueFloat, pretty: false, result: "X-1"},
			"/bla":         {cmd: "/bla", value: 0, ctype: Comment, pretty: false, result: "/bla"},
			"pretty %":     {cmd: "%", value: 0, ctype: Percent, pretty: true, result: "%"},
			"pretty F250":  {cmd: "F", value: 250, ctype: ValueInt, pretty: true, result: "F250"},
			"pretty G1":    {cmd: "G", value: 1, ctype: Address, pretty: true, result: "G01"},
			"pretty X-1.1": {cmd: "X", value: -1.1, ctype: ValueFloat, pretty: true, result: "X-1.100"},
			"pretty X-1.0": {cmd: "X", value: -1.0, ctype: ValueFloat, pretty: true, result: "X-1.000"},
			"pretty /bla":  {cmd: "/bla", value: 0, ctype: Comment, pretty: true, result: "/bla"},
		}

		for name, tc := range passTests {
			t.Run(name, func(t *testing.T) {
				cmd := CodeCmd{Cmd: tc.cmd, Value: tc.value, Type: tc.ctype}
				result := cmd.String(tc.pretty)
				assert.EqualValues(tc.result, result)
			})
		}
	})
}
