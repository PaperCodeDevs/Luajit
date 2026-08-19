package lua

import (
	"github.com/PaperCodeDevs/Luajit/op"
)

func (c *gen) tryIf(pc, to int) int {
	if !isCmp(c.opAt(pc)) || pc+1 >= to || c.opAt(pc+1) != op.OpJMP {
		return -1
	}
	tgt := c.dest(pc + 1)
	if tgt <= pc+2 || tgt > to {
		return -1
	}
	after := c.emitBranch("if", pc, tgt, to)
	if after <= pc {
		return -1
	}
	return after - 1
}

func (c *gen) emitBranch(kw string, pc, tgt, to int) int {
	in := c.p.Ins[pc]
	cond := c.cmp(c.code(in), in)
	thenEnd, after := c.thenJoin(pc, tgt, to)
	c.line("%s %s then", kw, cond)
	c.indent++
	c.body(pc+2, thenEnd)
	c.indent--
	if after > tgt {
		elsePC := c.skipNoise(tgt, after)
		if elsePC+1 < after && isCmp(c.opAt(elsePC)) && c.opAt(elsePC+1) == op.OpJMP {
			etgt := c.dest(elsePC + 1)
			if etgt > elsePC+2 && etgt <= after && !c.whileHead(elsePC, etgt, after) {
				end := c.emitBranch("elseif", elsePC, etgt, after)
				c.mark(pc, end)
				return end
			}
		}
		c.line("else")
		c.indent++
		c.body(tgt, after)
		c.indent--
	}
	c.line("end")
	c.mark(pc, after)
	return after
}

func (c *gen) thenJoin(pc, tgt, to int) (int, int) {
	last := tgt - 1
	for last > pc+2 && c.opAt(last) == op.OpUCLO {
		last--
	}
	if last < pc+2 || last >= len(c.p.Ins) || c.opAt(last) != op.OpJMP {
		return tgt, tgt
	}
	if last > 0 && isCmp(c.opAt(last-1)) {
		return tgt, tgt
	}
	if isIter(c.opAt(tgt)) {
		return tgt, tgt
	}
	dest := c.dest(last)
	if dest > tgt && dest <= to {
		return last, dest
	}
	if dest > to && c.loopExit(pc, dest) {
		return last, to
	}
	return tgt, tgt
}

func (c *gen) loopExit(pc, dest int) bool {
	for i := 0; i < pc && i < len(c.p.Ins); i++ {
		if isLoop(c.opAt(i)) && c.dest(i) == dest {
			return true
		}
	}
	return false
}

func (c *gen) whileHead(pc, tgt, after int) bool {
	for l := pc + 2; l < tgt && l < after; l++ {
		if !isLoop(c.opAt(l)) {
			continue
		}
		ld := c.dest(l)
		if ld != tgt && ld != after {
			continue
		}
		lim := ld
		if lim > len(c.p.Ins) {
			lim = len(c.p.Ins)
		}
		for i := lim - 1; i > l; i-- {
			if c.opAt(i) != op.OpJMP {
				continue
			}
			d := c.dest(i)
			if d >= pc && d <= l {
				return true
			}
		}
	}
	return false
}
