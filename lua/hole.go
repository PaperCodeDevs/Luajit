package lua

import (
	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

// holeSkip drops TSETB/TSETM/TSETV/CALL that index or call a slot never written
// by a kept instruction. Params start live. Does not invent object slots.
// Also drops statement CALL after the SSL ASCII hole: skipped TGETS B>=frame
// (32 02 01 15) plus skipped 64 61 74 61 "data".
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
	if wrote[obj] == 0 {
		return true
	}
	if code == op.OpCALL && in.B == 1 && sslDataCall(d, p, pc, int(in.A)) {
		return true
	}
	return false
}

// sslDataCall is the SSLBuyGiftUIAutoGen extra s2('MiniUIManager'): immediately
// before the CALL, skipped ins are TGETS A=fn B>=frame (raw 32 02 01 15) and
// ASCII data (64 61 74 61). Fn slot already holds GetInst's result.
func sslDataCall(d *parse.Dump, p *parse.Proto, callPC, fn int) bool {
	if callPC <= 0 || p == nil {
		return false
	}
	sawTGETS, sawData := false, false
	for i := callPC - 1; i >= 0 && i >= callPC-4; i-- {
		in := p.Ins[i]
		code := op.Norm(d.Version, in.Op)
		if legal(d, p, in, code, i) {
			break
		}
		if in.Op == 0x64 && in.A == 0x61 && in.C == 0x74 && in.B == 0x61 {
			sawData = true
			continue
		}
		if code == op.OpTGETS && int(in.A) == fn && int(in.B) >= int(p.Frame) {
			sawTGETS = true
			continue
		}
		if in.Op == 0x73 && in.A == 0 && int(in.B) == fn {
			continue
		}
	}
	return sawTGETS && sawData
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
