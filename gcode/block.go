package gcode

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
)

type Blocks []*Block

type Block struct {
	Cmds      []CodeCmd
	HasData   bool
	IsClamped bool
	IsSkip    bool
	X         *CodeCmd
	Y         *CodeCmd
	Z         *CodeCmd
	G         *CodeCmd
	LastPass  int
}

func (b *Block) Init() {
	b.Cmds = make([]CodeCmd, 0)
	b.HasData = false
	b.IsClamped = false
	b.IsSkip = false
	b.X = nil
	b.Y = nil
	b.Z = nil
	b.G = nil
	b.LastPass = 0
}

func (b *Block) Copy() Block {
	var block Block
	block.Cmds = make([]CodeCmd, len(b.Cmds))
	for i, cmd := range b.Cmds {
		c := cmd
		block.Cmds[i] = c
	}
	block.Parse(false)

	block.IsClamped = b.IsClamped
	block.IsSkip = b.IsSkip
	block.LastPass = b.LastPass
	return block
}

func (b *Block) Parse(multiCheck bool) error {
	b.HasData = false

	for i, cmd := range b.Cmds {

		switch cmd.Cmd {
		case "G":
			{
				if cmd.Value <= 1 { //only select G0 and G1
					b.HasData = true
					if multiCheck && b.G != nil {
						return errors.New("Multiple G0/G1 in block")
					}
					b.G = &b.Cmds[i]
				}
			}
		case "X":
			{
				b.HasData = true
				if multiCheck && b.X != nil {
					return errors.New("Multiple X values in block")
				}
				b.X = &b.Cmds[i]
			}
		case "Y":
			{
				b.HasData = true
				if multiCheck && b.Y != nil {
					return errors.New("Multiple Y values in block")
				}
				b.Y = &b.Cmds[i]
			}
		case "Z":
			{
				b.HasData = true
				if multiCheck && b.Z != nil {
					return errors.New("Multiple Z values in block")
				}
				b.Z = &b.Cmds[i]
			}
		}
	}
	return nil
}

func (b *Block) String(newline bool, pretty bool) string {
	var sb strings.Builder
	for _, cmd := range b.Cmds {
		sb.WriteString(cmd.String(pretty))
		if pretty {
			sb.WriteString(" ")
		}
	}
	if newline {
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *Block) SetX(value float32) {
	if b.X != nil {
		b.X.Value = value
	} else {
		cmd := CodeCmd{Cmd: "X", Value: value, Type: ValueFloat}
		b.Cmds = append(b.Cmds, cmd)
		b.Parse(false)
		b.HasData = true
	}
}

func (b *Block) SetY(value float32) {
	if b.Y != nil {
		b.Y.Value = value
	} else {
		cmd := CodeCmd{Cmd: "Y", Value: value, Type: ValueFloat}
		b.Cmds = append(b.Cmds, cmd)
		b.Parse(false)
		b.HasData = true
	}
}

func (b *Block) SetZ(value float32) {
	if b.Z != nil {
		b.Z.Value = value
	} else {
		cmd := CodeCmd{Cmd: "Z", Value: value, Type: ValueFloat}
		b.Cmds = append(b.Cmds, cmd)
		b.Parse(false)
		b.HasData = true
	}
}

func (b *Block) SetG(value float32) {
	if b.G != nil {
		b.G.Value = value
	} else {
		cmd := CodeCmd{Cmd: "G", Value: value, Type: Address}
		b.Cmds = append([]CodeCmd{cmd}, b.Cmds...)
		b.Parse(false)
		b.HasData = true
	}
}

func (b *Block) NoChangeY(value float32) bool {
	if b.Y == nil {
		return true
	}
	return b.Y.Value == value
}

func (b *Block) NoChangeZ(value float32) bool {
	if b.Z == nil {
		return true
	}
	return b.Z.Value == value
}

func (b *Block) ToStepZ(info *Info, pass int) {
	if b.IsClamped || b.Z == nil {
		return
	}
	if (*b.Z).Value >= 0 {
		b.LastPass = 0
		b.IsClamped = false
		return
	}
	//working with negative Z
	zMaxCut := info.Increment * float32(pass)
	b.LastPass = int(math.Ceil(float64((*b.Z).Value/info.Increment)) - 1)

	zCut := info.Increment * float32(b.LastPass)
	if zCut < zMaxCut {
		zCut = zMaxCut
	}
	b.SetZ(zCut + info.MinCut)
	b.IsClamped = true
}

func ParseLine(line string) (*Block, error) {
	block := new(Block)
	block.Init()

	var err error
	if len(line) == 0 {
		return block, err
	}

	rs := NewRunesScanner(line)
	for rs.Scan() {
		var cc CodeCmd
		r := unicode.ToUpper(rs.Rune())

		switch r {
		case ' ', '\t':
			{
				// discard spaces
				continue
			}
		case '%':
			{
				cc = CodeCmd{Cmd: string(r), Type: Percent}
			}
		case 'G', 'M':
			{
				cc = CodeCmd{Cmd: string(r), Type: Address}
				cc.Value, err = rs.GetValue(Address)
				if err != nil {
					break
				}
			}
		case 'F', 'S':
			{
				cc = CodeCmd{Cmd: string(r), Type: ValueInt}
				cc.Value, err = rs.GetValue(ValueInt)
				if err != nil {
					break
				}
				cc.Type = ValueInt
			}
		case 'X', 'Y', 'Z':
			{
				cc = CodeCmd{Cmd: string(r), Type: ValueFloat}
				cc.Value, err = rs.GetValue(ValueFloat)
				if err != nil {
					break
				}
			}
		case '/':
			{
				if rs.index != 1 { //rs.index is incremented
					err = errors.New("Invalid position for '/'")
					break
				}
				cc = CodeCmd{Cmd: fmt.Sprintf("/%s", string(*rs.Until(false, ' '))), Type: Comment}
			}
		case ';':
			{
				cc = CodeCmd{Cmd: fmt.Sprintf(";%s", string(*rs.Until(false, ' '))), Type: Comment}
			}
		case '(':
			{
				cc = CodeCmd{Cmd: fmt.Sprintf("(%s)", string(*rs.Until(r == '(', ')'))), Type: Comment}
			}
		default:
			{
				err = errors.New("Unexpected character " + string(r))
				break
			}
		}

		if !cc.Supported() {
			err = errors.New(fmt.Sprintf("Unsuported command: %s", cc.String(true)))
		}
		if err != nil {
			break
		}
		block.Cmds = append(block.Cmds, cc)
	}
	if err == nil {
		err = block.Parse(true)
	}
	return block, err
}
