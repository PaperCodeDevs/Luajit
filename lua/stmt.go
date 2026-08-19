package lua

import (
	"strconv"
	"strings"

	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func (c *gen) stmt(in parse.Ins, code byte) {
	a := int(in.A)
	if code == op.OpISNES && op.MagicRet(in.A, in.D) {
		c.line("return")
		return
	}
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
		c.set(a, c.gcstr(in.D, in.C))
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
		c.set(a, c.uv(c.p.UVKey(in.D, in.C)))
	case op.OpUSETV:
		c.line("%s = %s", c.uv(a), c.get(int(in.D)))
	case op.OpUSETS:
		c.line("%s = %s", c.uv(a), c.gcstr(in.D, in.C))
	case op.OpUSETN:
		c.line("%s = %s", c.uv(a), c.numD(in.D))
	case op.OpUSETP:
		c.line("%s = %s", c.uv(a), pri(in.D))
	case op.OpFNEW:
		ch := c.p.FNew(in.D, in.C)
		if ch != nil {
			if c.used == nil {
				c.used = map[*parse.Proto]bool{}
			}
			c.used[ch] = true
		}
		var inner strings.Builder
		emitFn(&inner, c.d, ch, c.indent, false)
		c.set(a, strings.TrimSpace(inner.String()))
		c.line("local s%d = %s", a, c.get(a))
		c.set(a, "s"+strconv.Itoa(a))
	case op.OpTNEW:
		c.set(a, "{}")
	case op.OpTDUP:
		lit := c.dup(in.D, in.C)
		c.line("local s%d = %s", a, lit)
		c.set(a, "s"+strconv.Itoa(a))
	case op.OpGGET:
		c.set(a, c.gcname(in.D, in.C))
	case op.OpGSET:
		c.line("%s = %s", c.gcname(in.D, in.C), c.get(a))
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
	case op.OpTSETM:
		c.tsetm(in)
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
	case op.OpVARG:
		c.set(a, "...")
		c.line("local s%d = ...", a)
		c.set(a, "s"+strconv.Itoa(a))
	case op.OpBAND:
		c.set(a, "bit.band("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpBOR:
		c.set(a, "bit.bor("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpBXOR:
		c.set(a, "bit.bxor("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpBSHL:
		c.set(a, "bit.lshift("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpBSHR:
		c.set(a, "bit.rshift("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpBSAR:
		c.set(a, "bit.arshift("+c.get(int(in.B))+", "+c.get(int(in.C))+")")
	case op.OpJMP, op.OpUCLO, op.OpLOOP, op.OpILOOP, op.OpJLOOP, op.OpFORL, op.OpIFORL, op.OpJFORL, op.OpITERL, op.OpIITERL, op.OpJITERL, op.OpISNEXT, op.OpITERC, op.OpITERN:
	default:
		if !op.DumpOK(code) {
			return
		}
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
