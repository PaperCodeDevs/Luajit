package parse

func readProto(r *reader, fileFlags uint32, ready *[]*Proto) (*Proto, error) {
	p := &Proto{}
	var err error
	if p.Flags, err = r.byte(); err != nil {
		return nil, err
	}
	if p.Params, err = r.byte(); err != nil {
		return nil, err
	}
	if p.Frame, err = r.byte(); err != nil {
		return nil, err
	}
	nuv, err := r.byte()
	if err != nil {
		return nil, err
	}
	sizekgc, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	sizekn, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	numbc, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	if sizekgc > maxKGC || sizekn > maxKNum || numbc > maxIns {
		return nil, r.err("proto counts")
	}
	if int(numbc)*4 > r.remain() {
		return nil, r.err("bc overflow")
	}
	var sizedbg uint32
	if fileFlags&FlagStrip == 0 {
		sizedbg, err = r.uleb32()
		if err != nil {
			return nil, err
		}
		if sizedbg != 0 {
			if p.FirstLine, err = r.uleb32(); err != nil {
				return nil, err
			}
			if p.NumLine, err = r.uleb32(); err != nil {
				return nil, err
			}
		}
	}
	p.Ins = make([]Ins, numbc)
	for i := range p.Ins {
		if p.Ins[i], err = r.ins(); err != nil {
			return nil, err
		}
	}
	p.UV = make([]uint16, nuv)
	for i := range p.UV {
		if p.UV[i], err = r.u16(); err != nil {
			return nil, err
		}
	}
	if p.KGC, err = readKGC(r, int(sizekgc), ready); err != nil {
		return nil, err
	}
	if p.KNum, err = readKNum(r, int(sizekn)); err != nil {
		return nil, err
	}
	if sizedbg != 0 {
		if err = readDebug(r, p, int(sizedbg), int(numbc)); err != nil {
			return nil, err
		}
	}
	return p, nil
}
