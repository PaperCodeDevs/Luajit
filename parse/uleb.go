package parse

func (r *reader) uleb() (uint64, error) {
	var v uint64
	var s uint
	for i := 0; i < 10; i++ {
		c, err := r.byte()
		if err != nil {
			return 0, r.err("uleb")
		}
		v |= uint64(c&0x7f) << s
		if c&0x80 == 0 {
			return v, nil
		}
		s += 7
	}
	return 0, r.err("uleb overflow")
}

func (r *reader) uleb32() (uint32, error) {
	v, err := r.uleb()
	if err != nil {
		return 0, err
	}
	if v > 0xffffffff {
		return 0, r.err("uleb32")
	}
	return uint32(v), nil
}

func (r *reader) uleb33() (bool, uint32, error) {
	c, err := r.byte()
	if err != nil {
		return false, 0, r.err("uleb33")
	}
	isnum := c&1 != 0
	v := uint32(c) >> 1
	if v < 0x40 {
		return isnum, v, nil
	}
	v &= 0x3f
	sh := uint(6)
	for {
		c, err = r.byte()
		if err != nil {
			return false, 0, r.err("uleb33")
		}
		v |= uint32(c&0x7f) << sh
		if c < 0x80 {
			return isnum, v, nil
		}
		sh += 7
		if sh > 32 {
			return false, 0, r.err("uleb33 overflow")
		}
	}
}
