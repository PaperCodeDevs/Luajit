package parse

func readDebug(r *reader, p *Proto, sizedbg, numbc int) error {
	raw, err := r.bytes(sizedbg)
	if err != nil {
		return err
	}
	width := 1
	if p.NumLine >= 65536 {
		width = 4
	} else if p.NumLine >= 256 {
		width = 2
	}
	need := numbc * width
	if need > len(raw) {
		return nil
	}
	p.Lines = make([]uint32, numbc)
	off := 0
	for i := 0; i < numbc; i++ {
		switch width {
		case 1:
			p.Lines[i] = uint32(raw[off])
		case 2:
			p.Lines[i] = uint32(raw[off]) | uint32(raw[off+1])<<8
		default:
			p.Lines[i] = uint32(raw[off]) | uint32(raw[off+1])<<8 | uint32(raw[off+2])<<16 | uint32(raw[off+3])<<24
		}
		off += width
	}
	dr := &reader{b: raw, i: off}
	p.UVName = make([]string, len(p.UV))
	for i := range p.UVName {
		s, _, err := dr.cstr(len(dr.b) - dr.i)
		if err != nil {
			p.UVName = p.UVName[:i]
			return nil
		}
		p.UVName[i] = s
	}
	for dr.i < len(dr.b) {
		if dr.b[dr.i] == 0 {
			break
		}
		if dr.b[dr.i] < 8 {
			dr.i++
		} else {
			s, _, err := dr.cstr(len(dr.b) - dr.i)
			if err != nil {
				break
			}
			if s == "" {
				break
			}
			start, err1 := dr.uleb32()
			end, err2 := dr.uleb32()
			if err1 != nil || err2 != nil {
				break
			}
			p.Var = append(p.Var, Local{Name: s, Start: start, End: end})
			continue
		}
		start, err1 := dr.uleb32()
		end, err2 := dr.uleb32()
		if err1 != nil || err2 != nil {
			break
		}
		p.Var = append(p.Var, Local{Start: start, End: end})
	}
	return nil
}

func (p *Proto) SlotName(slot, pc int) string {
	if p == nil || slot < 0 {
		return ""
	}
	n := 0
	for _, v := range p.Var {
		if int(v.Start) > pc || pc >= int(v.End) {
			continue
		}
		if n == slot {
			return v.Name
		}
		n++
	}
	return ""
}
