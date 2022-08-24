package gcode

import (
	"fmt"
	"strings"
)

type CmdType int

const (
	Address CmdType = iota
	ValueFloat
	ValueInt
	Comment
	Percent
)

type CodeCmd struct {
	Cmd   string
	Value float32
	Type  CmdType
}

func (c CodeCmd) String(pretty bool) string {
	switch c.Type {
	case Address:
		{
			if pretty {
				return fmt.Sprintf("%s%02.0f", c.Cmd, c.Value)
			} else {
				return fmt.Sprintf("%s%.0f", c.Cmd, c.Value)
			}
		}
	case ValueFloat:
		{
			if pretty {
				return fmt.Sprintf("%s%.3f", c.Cmd, c.Value)
			} else {
				return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%s%.3f", c.Cmd, c.Value), "0"), ".")
			}
		}
	case ValueInt:
		{
			return fmt.Sprintf("%s%.0f", c.Cmd, c.Value)
		}
	}

	return c.Cmd
}

func (c CodeCmd) Supported() bool {
	if c.Type == Percent || c.Type == Comment {
		return true
	}
	switch c.Cmd {
	case "F", "S", "X", "Y", "Z":
		{
			return true
		}
	case "G":
		{
			switch c.Value {
			case 0, 1, 20, 21, 90:
				{
					return true
				}
			}
		}
	case "M":
		{
			switch c.Value {
			case 3, 5, 30:
				{
					return true
				}
			}
		}
	}
	return false
}
