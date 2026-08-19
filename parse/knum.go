package parse

import "math"

func readKNum(r *reader, n int) ([]KNum, error) {
	out := make([]KNum, n)
	for i := 0; i < n; i++ {
		isnum, lo, err := r.uleb33()
		if err != nil {
			return nil, err
		}
		if !isnum {
			out[i] = KNum{IsInt: true, I: int32(lo)}
			continue
		}
		hi, err := r.uleb32()
		if err != nil {
			return nil, err
		}
		out[i] = KNum{Lo: lo, Hi: hi}
	}
	return out, nil
}

func (k KNum) Float64() float64 {
	if k.IsInt {
		return float64(k.I)
	}
	u := uint64(k.Lo) | uint64(k.Hi)<<32
	return math.Float64frombits(u)
}

func (p *Proto) GCKey(d uint16, c byte) uint16 {
	if int(d) < len(p.KGC) {
		return d
	}
	return uint16(c)
}

func (p *Proto) UVKey(d uint16, c byte) int {
	if int(d) < len(p.UV) {
		return int(d)
	}
	return int(c)
}

func (p *Proto) GC(d uint16) (KGC, bool) {
	i := len(p.KGC) - 1 - int(d)
	if i < 0 || i >= len(p.KGC) {
		return KGC{}, false
	}
	return p.KGC[i], true
}

func (p *Proto) Str(d uint16) string {
	s, _ := p.StrOK(d)
	return s
}

func (p *Proto) StrOK(d uint16) (string, bool) {
	k, ok := p.GC(d)
	if !ok || k.Kind != KStr {
		return "", false
	}
	return k.Str, true
}

func (p *Proto) StrC(d uint16, c byte) string {
	return p.Str(p.GCKey(d, c))
}

func (p *Proto) Child(d uint16) *Proto {
	k, ok := p.GC(d)
	if !ok {
		return nil
	}
	return k.Child
}

func (p *Proto) FNew(d uint16, c byte) *Proto {
	return p.Child(p.GCKey(d, c))
}

func (p *Proto) NumKey(d uint16, c byte) uint16 {
	if int(d) < len(p.KNum) {
		return d
	}
	return uint16(c)
}

func (p *Proto) Num(d uint16) (KNum, bool) {
	if int(d) >= len(p.KNum) {
		return KNum{}, false
	}
	return p.KNum[d], true
}
