package lua

import (
	"strconv"
	"strings"

	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func (c *gen) stmt(in parse.Ins, code byte) {
	a := int(in.A)
	switch code {
	case op.OpMOV:
		c.set(a, c.get(int(in.D)))
	case op.OpNOT:
		c.set(a, "not "+c.get(int(in.D)))
	case op.OpUNM:
		c.set(a, "-"+c.get(int(in.D)))
	case op.OpLEN:
		c.set(a, "#"+c.get(int(in.D)))
	case op.OpBNOT:
		c.set(a, "bit.bnot("+c.get(int(in.D))+")")
	case op.OpKSTR:
		c.set(a, c.gcstr(in.D))
	case op.OpKSHORT:
		c.set(a, shortInt(in.D))
	case op.OpKNUM:
		c.set(a, c.numD(in.D))
	case op.OpKPRI:
		c.set(a, pri(in.D))
	case op.OpKNIL:
		for i := a; i <= int(in.D); i++ {
			c.set(i, "nil")
		}
	case op.OpUGET:
		c.set(a, c.uv(int(in.D)))
	case op.OpUSETV:
		c.line("%s = %s", c.uv(a), c.get(int(in.D)))
	case op.OpUSETS:
		c.line("%s = %s", c.uv(a), c.gcstr(in.D))
	case op.OpUSETN:
		c.line("%s = %s", c.uv(a), c.numD(in.D))
	case op.OpUSETP:
		c.line("%s = %s", c.uv(a), pri(in.D))
	case op.OpFNEW:
		ch := c.p.Child(in.D)
		var inner strings.Builder
		emitFn(&inner, c.d, ch, c.indent, false)
		c.set(a, strings.TrimSpace(inner.String()))
		c.line("local s%d = %s", a, c.get(a))
		c.set(a, "s"+strconv.Itoa(a))
	case op.OpTNEW:
		c.set(a, "{}")
	case op.OpTDUP:
		c.set(a, c.dup(in.D))
	case op.OpGGET:
		c.set(a, c.gcname(in.D))
	case op.OpGSET:
		c.line("%s = %s", c.gcname(in.D), c.get(a))
	case op.OpTGETV:
		c.set(a, c.get(int(in.B))+"["+c.get(int(in.C))+"]")
	case op.OpTGETS:
		c.set(a, c.idx(c.get(int(in.B)), c.p.Str(uint16(in.C))))
	case op.OpTGETB:
		c.set(a, c.get(int(in.B))+"["+strconv.Itoa(int(in.C))+"]")
	case op.OpTGETR:
		c.set(a, c.get(int(in.B))+"["+c.get(int(in.C))+"]")
	case op.OpTSETV:
		c.line("%s[%s] = %s", c.get(int(in.B)), c.get(int(in.C)), c.get(a))
	case op.OpTSETS:
		c.line("%s = %s", c.idx(c.get(int(in.B)), c.p.Str(uint16(in.C))), c.get(a))
	case op.OpTSETB:
		c.line("%s[%d] = %s", c.get(int(in.B)), in.C, c.get(a))
	case op.OpTSETR:
		c.line("%s[%s] = %s", c.get(int(in.B)), c.get(int(in.C)), c.get(a))
	case op.OpADDVV, op.OpSUBVV, op.OpMULVV, op.OpDIVVV, op.OpMODVV, op.OpPOW:
		c.set(a, "("+c.get(int(in.B))+c.binop(code)+c.get(int(in.C))+")")
	case op.OpADDVN, op.OpSUBVN, op.OpMULVN, op.OpDIVVN, op.OpMODVN:
		c.set(a, "("+c.get(int(in.B))+c.binop(code)+c.numD(uint16(in.C))+")")
	case op.OpADDNV, op.OpSUBNV, op.OpMULNV, op.OpDIVNV, op.OpMODNV:
		c.set(a, "("+c.numD(uint16(in.C))+c.binop(code)+c.get(int(in.B))+")")
	case op.OpCAT:
		c.set(a, c.cat(int(in.B), int(in.C)))
	case op.OpCALL, op.OpCALLM:
		c.call(in, code, false)
	case op.OpCALLT, op.OpCALLMT:
		c.call(in, code, true)
	case op.OpRET, op.OpRETM, op.OpRET0, op.OpRET1:
		c.ret(in, code)
	case op.OpJMP, op.OpUCLO, op.OpLOOP, op.OpILOOP, op.OpJLOOP, op.OpFORL, op.OpIFORL, op.OpJFORL, op.OpITERL, op.OpISNEXT, op.OpITERC, op.OpITERN, op.OpVARG:
	default:
		c.line("-- %s %d %d", op.Name(c.d.Version, in.Op), in.A, in.D)
	}
}

func (c *gen) binop(code byte) string {
	switch code {
	case op.OpADDVV, op.OpADDVN, op.OpADDNV:
		return " + "
	case op.OpSUBVV, op.OpSUBVN, op.OpSUBNV:
		return " - "
	case op.OpMULVV, op.OpMULVN, op.OpMULNV:
		return " * "
	case op.OpDIVVV, op.OpDIVVN, op.OpDIVNV:
		return " / "
	case op.OpMODVV, op.OpMODVN, op.OpMODNV:
		return " % "
	case op.OpPOW:
		return " ^ "
	default:
		return " + "
	}
}

func (c *gen) idx(obj, field string) string {
	if isIdent(field) {
		return obj + "." + field
	}
	if field == "" {
		return obj + "[?]"
	}
	return obj + "[" + quote(field) + "]"
}

func (c *gen) cat(b, cc int) string {
	if cc < b {
		return "''"
	}
	parts := make([]string, 0, cc-b+1)
	for i := b; i <= cc; i++ {
		parts = append(parts, c.get(i))
	}
	return strings.Join(parts, " .. ")
}

func (c *gen) uv(i int) string {
	if i >= 0 && i < len(c.p.UVName) && c.p.UVName[i] != "" && isIdent(c.p.UVName[i]) {
		return c.p.UVName[i]
	}
	return "u" + strconv.Itoa(i)
}

func (c *gen) dup(d uint16) string {
	k, ok := c.p.GC(d)
	if !ok || k.Tab == nil {
		return "{}"
	}
	return "{}"
}

func (c *gen) call(in parse.Ins, code byte, tail bool) {
	base := int(in.A)
	fn := c.get(base + c.fr2)
	narg := int(in.C) - 1
	if code == op.OpCALLM || code == op.OpCALLMT {
		narg = -1
	}
	args := []string{}
	start := base + c.fr2 + 1
	if narg < 0 {
		for i := start; i < len(c.slot) && i < start+16; i++ {
			args = append(args, c.get(i))
		}
		args = append(args, "...")
	} else {
		for i := 0; i < narg; i++ {
			args = append(args, c.get(start+i))
		}
	}
	call := fn + "(" + strings.Join(args, ", ") + ")"
	nres := int(in.B) - 1
	if tail {
		c.line("return %s", call)
		return
	}
	if nres < 0 {
		nres = 1
	}
	if nres == 0 {
		c.line("%s", call)
		return
	}
	if nres == 1 {
		c.set(base, call)
		c.line("local s%d = %s", base, call)
		c.set(base, "s"+strconv.Itoa(base))
		return
	}
	if nres > 32 {
		nres = 32
	}
	lhs := make([]string, nres)
	for i := 0; i < nres; i++ {
		lhs[i] = "s" + strconv.Itoa(base+i)
		c.set(base+i, lhs[i])
	}
	c.line("local %s = %s", strings.Join(lhs, ", "), call)
}

func (c *gen) ret(in parse.Ins, code byte) {
	if code == op.OpRET0 {
		c.line("return")
		return
	}
	n := 1
	base := int(in.A)
	if code == op.OpRET1 {
		n = 1
	} else if code == op.OpRET {
		n = int(in.D) - 1
	} else {
		n = -1
	}
	if n <= 0 {
		c.line("return %s", c.get(base))
		return
	}
	if n > 32 {
		n = 32
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = c.get(base + i)
	}
	c.line("return %s", strings.Join(parts, ", "))
}
