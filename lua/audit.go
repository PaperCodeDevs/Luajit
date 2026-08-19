package lua

import (
	"strings"

	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func countBadOp(src string) int {
	n := 0
	for _, line := range strings.Split(src, "\n") {
		s := strings.TrimSpace(line)
		if !strings.HasPrefix(s, "-- ") {
			continue
		}
		name := strings.Fields(s)
		if len(name) < 2 {
			continue
		}
		switch name[1] {
		case "LOOP", "ILOOP", "JLOOP", "ITERC", "ITERN", "ITERL", "IITERL",
			"TSETM", "VARG", "BAND", "BOR", "BXOR", "BSHL", "BSHR", "BSAR",
			"IFUNCV", "FUNCV", "FUNCF", "IFUNCF", "JFUNCF", "JFUNCV", "FUNCC", "FUNCCW", "?":
			n++
		}
	}
	return n
}

func auditFn(d *parse.Dump, src string, cov *Cover) {
	need := 0
	walkProto(d.Main, func(p *parse.Proto) {
		for _, in := range p.Ins {
			if op.Norm(d.Version, in.Op) != op.OpFNEW {
				continue
			}
			if p.FNew(in.D, in.C) != nil {
				need++
			}
		}
	})
	if need == 0 {
		return
	}
	cov.NeedFn += need
	nAssign := strings.Count(src, " = function(")
	nLeft := strings.Count(src, "local _c")
	if nAssign-nLeft <= 0 {
		cov.MissFn += need
		cov.note("fnew")
	}
}
