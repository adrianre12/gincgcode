package gcode

import (
	"errors"
	"strconv"
	"strings"
)

type RunesScanner struct {
	runes []rune
	index int
}

func NewRunesScanner(str string) *RunesScanner {
	return &RunesScanner{[]rune(str), 0}
}

func (rs *RunesScanner) Scan() bool {
	return rs.index < len(rs.runes)
}

func (rs *RunesScanner) Rune() rune {
	r := rs.runes[rs.index]
	rs.index++
	return r
}

func (rs *RunesScanner) GetValue(cmdType CmdType) (value float32, err error) {
	valueRunes := make([]rune, 0)
	for rs.Scan() {
		r := rs.runes[rs.index]
		if strings.ContainsRune("+-0123456789.", r) {
			valueRunes = append(valueRunes, r)
			rs.index++
		} else {
			break
		}
	}
	switch cmdType {
	case ValueInt, Address:
		{
			var vi int64
			vi, err = strconv.ParseInt(string(valueRunes), 10, 32)
			value = float32(vi)
		}
	case ValueFloat:
		{
			var vf float64
			vf, err = strconv.ParseFloat(string(valueRunes), 10)
			value = float32(vf)
		}
	default:
		{
			err = errors.New("Unsuported Value Type")
		}
	}

	return
}

// Returns all runes til the endRune or EOL, matching endRune is not returned.
func (rs *RunesScanner) Until(match bool, endRune rune) *[]rune {
	runes := make([]rune, 0)
	for rs.Scan() {
		r := rs.Rune()
		if match && r == endRune {
			break
		}
		runes = append(runes, r)
	}
	return &runes
}
