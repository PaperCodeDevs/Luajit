package parse

func readHeader(r *reader) (*Dump, error) {
	if r.i+4 > len(r.b) {
		return nil, r.err("short")
	}
	if r.b[0] != 0x1b || r.b[1] != 'L' || r.b[2] != 'J' {
		return nil, r.err("magic")
	}
	ver := r.b[3]
	if ver != VerStd && ver != VerMW {
		return nil, r.err("version")
	}
	r.i = 4
	flags, err := r.uleb32()
	if err != nil {
		return nil, err
	}
	if flags > maxFlag {
		return nil, r.err("flags")
	}
	d := &Dump{Version: ver, Flags: flags}
	if flags&FlagStrip == 0 {
		n, err := r.uleb32()
		if err != nil {
			return nil, err
		}
		if n > maxName {
			return nil, r.err("name")
		}
		s, err := r.bytes(int(n))
		if err != nil {
			return nil, err
		}
		d.Name = string(s)
	}
	return d, nil
}
