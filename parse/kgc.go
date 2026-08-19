package parse

func readKGC(r *reader, n int, ready *[]*Proto) ([]KGC, error) {
	out := make([]KGC, 0, n)
	for i := 0; i < n; i++ {
		tp, err := r.uleb32()
		if err != nil {
			return nil, err
		}
		var k KGC
		switch {
		case tp >= KStr:
			s, err := r.bytes(int(tp - KStr))
			if err != nil {
				return nil, err
			}
			k = KGC{Kind: KStr, Str: string(s)}
		case tp == KTab:
			tab, err := readTab(r)
			if err != nil {
				return nil, err
			}
			k = KGC{Kind: KTab, Tab: tab}
		case tp == KI64, tp == KU64:
			lo, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			hi, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			k = KGC{Kind: int(tp), Lo: lo, Hi: hi}
		case tp == KComplex:
			lo, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			hi, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			ilo, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			ihi, err := r.uleb32()
			if err != nil {
				return nil, err
			}
			k = KGC{Kind: KComplex, Lo: lo, Hi: hi, ILo: ilo, IHi: ihi}
		case tp == KChild:
			if len(*ready) == 0 {
				return nil, r.err("child")
			}
			last := len(*ready) - 1
			k = KGC{Kind: KChild, Child: (*ready)[last]}
			*ready = (*ready)[:last]
		default:
			return nil, r.err("kgc")
		}
		out = append(out, k)
	}
	return out, nil
}

func readTab(r *reader) (*Tab, error) {
	na, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	nh, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	if na > maxTab || nh > maxTab {
		return nil, r.err("tab")
	}
	if int(na)+2*int(nh) > r.remain() {
		return nil, r.err("tab")
	}
	t := &Tab{
		Array: make([]TabK, na),
		Hash:  make([][2]TabK, nh),
	}
	for i := range t.Array {
		if t.Array[i], err = readTabK(r); err != nil {
			return nil, err
		}
	}
	for i := range t.Hash {
		k, err := readTabK(r)
		if err != nil {
			return nil, err
		}
		v, err := readTabK(r)
		if err != nil {
			return nil, err
		}
		t.Hash[i] = [2]TabK{k, v}
	}
	return t, nil
}

func readTabK(r *reader) (TabK, error) {
	tp, err := r.uleb32()
	if err != nil {
		return TabK{}, err
	}
	if tp >= TabStr {
		s, err := r.bytes(int(tp - TabStr))
		if err != nil {
			return TabK{}, err
		}
		return TabK{Kind: TabStr, Str: string(s)}, nil
	}
	switch tp {
	case TabInt:
		v, err := r.uleb32()
		if err != nil {
			return TabK{}, err
		}
		return TabK{Kind: TabInt, I: int32(v)}, nil
	case TabNum:
		lo, err := r.uleb32()
		if err != nil {
			return TabK{}, err
		}
		hi, err := r.uleb32()
		if err != nil {
			return TabK{}, err
		}
		return TabK{Kind: TabNum, Lo: lo, Hi: hi}, nil
	case TabNil, TabFalse, TabTrue:
		return TabK{Kind: int(tp)}, nil
	default:
		return TabK{}, r.err("ktab")
	}
}
