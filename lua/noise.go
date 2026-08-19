package lua

import (
	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func printable(in parse.Ins) bool {
	for _, c := range []byte{in.Op, in.A, in.C, in.B} {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func skipAscii(d *parse.Dump, p *parse.Proto, in parse.Ins) bool {
	if !printable(in) {
		return false
	}
	return !keepAscii(d, p, in)
}

func keepAscii(d *parse.Dump, p *parse.Proto, in parse.Ins) bool {
	if p == nil {
		return false
	}
	code := in.Op
	if d != nil {
		code = op.Norm(d.Version, in.Op)
	}
	fr := int(p.Frame)
	switch code {
	case op.OpTGETS, op.OpTSETS:
		_, ok := p.StrOK(uint16(in.C))
		return ok
	case op.OpTGETV, op.OpTSETV, op.OpTGETR, op.OpTSETR, op.OpTGETB, op.OpTSETB:
		return int(in.A) < fr && int(in.B) < fr && int(in.C) < fr
	case op.OpGGET, op.OpGSET, op.OpKSTR:
		return p.Str(p.GCKey(in.D, in.C)) != ""
	case op.OpFNEW:
		return p.FNew(in.D, in.C) != nil
	case op.OpTDUP:
		k, ok := p.GC(p.GCKey(in.D, in.C))
		return ok && k.Tab != nil
	case op.OpCALL, op.OpCALLM, op.OpCALLT, op.OpCALLMT:
		return int(in.A) < fr && int(in.C) <= fr+2 && int(in.B) <= fr+2
	default:
		return false
	}
}

func cmpOK(p *parse.Proto, in parse.Ins, code byte) bool {
	if p == nil {
		return false
	}
	fr := int(p.Frame)
	switch code {
	case op.OpISLT, op.OpISGE, op.OpISLE, op.OpISGT, op.OpISEQV, op.OpISNEV, op.OpISTC, op.OpISFC:
		return int(in.A) < fr && int(in.D) < fr
	case op.OpISEQS, op.OpISNES, op.OpISEQN, op.OpISNEN, op.OpISEQP, op.OpISNEP:
		return int(in.A) < fr
	case op.OpIST, op.OpISF:
		return int(in.D) < fr
	default:
		return true
	}
}
