package lua

import (
	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

// holeSkip drops TSETB/TSETM/TSETV/CALL that index or call a slot never written
// by a kept instruction. Params start live. Does not invent object slots.
func holeSkip(d *parse.Dump, p *parse.Proto, in parse.Ins, code byte, pc int) bool {
	switch code {
	case op.OpTSETB, op.OpTSETV, op.OpTSETM, op.OpCALL, op.OpCALLM, op.OpCALLT, op.OpCALLMT:
	default:
		return false
	}
	fr := int(p.Frame)
	if fr <= 0 {
		return false
	}
	wrote := make([]byte, fr)
	for i := 0; i < int(p.Params) && i < fr; i++ {
		wrote[i] = 1
	}
	n := len(p.Ins)
	if pc > n {
		pc = n
	}
	for i := 0; i < pc; i++ {
		ins := p.Ins[i]
		c := ins.Op
		if d != nil {
			c = op.Norm(d.Version, ins.Op)
		}
		if !legal(d, p, ins, c, i) {
			continue
		}
		markWrite(wrote, ins, c)
	}
	obj := -1
	switch code {
	case op.OpTSETB, op.OpTSETV:
		obj = int(in.B)
	case op.OpTSETM:
		obj = int(in.A) - 1
	default:
		obj = int(in.A)
	}
	if obj < 0 || obj >= fr {
		return true
	}
	return wrote[obj] == 0
}

func markWrite(wrote []byte, in parse.Ins, code byte) {
	fr := len(wrote)
	set := func(s int) {
		if s >= 0 && s < fr {
			wrote[s] = 1
		}
	}
	a := int(in.A)
	dd := int(in.D)
	switch code {
	case op.OpMOV:
		if dd >= 0 && dd < fr && wrote[dd] != 0 {
			set(a)
		}
	case op.OpKNIL:
		for i := a; i <= dd && i < fr; i++ {
			set(i)
		}
	case op.OpNOT, op.OpUNM, op.OpLEN, op.OpBNOT,
		op.OpKSTR, op.OpKSHORT, op.OpKNUM, op.OpKPRI, op.OpUGET, op.OpFNEW,
		op.OpTNEW, op.OpTDUP, op.OpGGET,
		op.OpTGETV, op.OpTGETS, op.OpTGETB, op.OpTGETR,
		op.OpADDVN, op.OpSUBVN, op.OpMULVN, op.OpDIVVN, op.OpMODVN,
		op.OpADDNV, op.OpSUBNV, op.OpMULNV, op.OpDIVNV, op.OpMODNV,
		op.OpADDVV, op.OpSUBVV, op.OpMULVV, op.OpDIVVV, op.OpMODVV, op.OpPOW, op.OpCAT,
		op.OpBAND, op.OpBOR, op.OpBXOR, op.OpBSHL, op.OpBSHR, op.OpBSAR, op.OpVARG,
		op.OpCALL, op.OpCALLM:
		set(a)
	}
}
