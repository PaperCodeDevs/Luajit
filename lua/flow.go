package lua

import (
	"strconv"

	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || i > 0 && c >= '0' && c <= '9'
		if !ok {
			return false
		}
	}
	return true
}

func (c *gen) body(from, to int) {
	if c.skip == nil {
		c.skip = map[int]bool{}
	}
	nins := len(c.p.Ins)
	if from < 0 {
		from = 0
	}
	if to > nins {
		to = nins
	}
	for pc := from; pc < to; pc++ {
		if c.skip[pc] {
			continue
		}
		if pc >= len(c.p.Ins) {
			return
		}
		in := c.p.Ins[pc]
		code := c.code(in)
		if isCmp(code) && pc+1 < to && pc+1 < len(c.p.Ins) && c.code(c.p.Ins[pc+1]) == op.OpJMP {
			tgt := pc + 2 + c.p.Ins[pc+1].J()
			if tgt > pc+2 && tgt <= len(c.p.Ins) && tgt <= to {
				cond := c.cmp(code, in)
				c.line("if %s then", cond)
				c.indent++
				c.body(pc+2, tgt)
				c.indent--
				c.line("end")
				for i := pc; i < tgt; i++ {
					c.skip[i] = true
				}
				pc = tgt - 1
				continue
			}
		}
		if code == op.OpFORI {
			lim := pc + 1 + in.J()
			if lim > pc+1 && lim <= len(c.p.Ins) && lim <= to {
				idx := c.get(int(in.A) + 3)
				c.line("for %s = %s, %s, %s do", idx, c.get(int(in.A)), c.get(int(in.A)+1), c.get(int(in.A)+2))
				c.indent++
				c.body(pc+1, lim)
				c.indent--
				c.line("end")
				for i := pc; i < lim; i++ {
					c.skip[i] = true
				}
				pc = lim - 1
				continue
			}
		}
		c.stmt(in, code)
	}
}

func isCmp(code byte) bool {
	return code <= op.OpISNEP || code == op.OpIST || code == op.OpISF || code == op.OpISTC || code == op.OpISFC
}

func (c *gen) cmp(code byte, in parse.Ins) string {
	l, r := c.get(int(in.A)), c.get(int(in.D))
	switch code {
	case op.OpISLT:
		return l + " < " + r
	case op.OpISGE:
		return l + " >= " + r
	case op.OpISLE:
		return l + " <= " + r
	case op.OpISGT:
		return l + " > " + r
	case op.OpISEQV:
		return l + " == " + r
	case op.OpISNEV:
		return l + " ~= " + r
	case op.OpISEQS:
		return l + " == " + c.gcstr(in.D)
	case op.OpISNES:
		return l + " ~= " + c.gcstr(in.D)
	case op.OpISEQN:
		return l + " == " + c.numD(in.D)
	case op.OpISNEN:
		return l + " ~= " + c.numD(in.D)
	case op.OpISEQP:
		return l + " == " + pri(in.D)
	case op.OpISNEP:
		return l + " ~= " + pri(in.D)
	case op.OpIST:
		return r
	case op.OpISF:
		return "not " + r
	case op.OpISTC:
		c.set(int(in.A), r)
		return r
	case op.OpISFC:
		c.set(int(in.A), r)
		return "not " + r
	default:
		return "true"
	}
}

func (c *gen) numD(d uint16) string {
	k, ok := c.p.Num(d)
	if !ok {
		return "n" + strconv.Itoa(int(d))
	}
	return numLit(k)
}
