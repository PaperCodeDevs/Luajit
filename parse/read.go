package parse

import "fmt"

type reader struct {
	b []byte
	i int
}

func (r *reader) err(msg string) error {
	return fmt.Errorf("luajit: %s at %d", msg, r.i)
}

func (r *reader) remain() int {
	if r.i >= len(r.b) {
		return 0
	}
	return len(r.b) - r.i
}

func (r *reader) byte() (byte, error) {
	if r.i >= len(r.b) {
		return 0, r.err("eof")
	}
	c := r.b[r.i]
	r.i++
	return c, nil
}

func (r *reader) bytes(n int) ([]byte, error) {
	if n < 0 || r.i+n > len(r.b) {
		return nil, r.err("eof")
	}
	s := r.b[r.i : r.i+n]
	r.i += n
	return s, nil
}

func (r *reader) u16() (uint16, error) {
	s, err := r.bytes(2)
	if err != nil {
		return 0, err
	}
	return uint16(s[0]) | uint16(s[1])<<8, nil
}

func (r *reader) ins() (Ins, error) {
	s, err := r.bytes(4)
	if err != nil {
		return Ins{}, err
	}
	return Ins{
		Op: s[0],
		A:  s[1],
		C:  s[2],
		B:  s[3],
		D:  uint16(s[2]) | uint16(s[3])<<8,
	}, nil
}

func (r *reader) cstr(limit int) (string, int, error) {
	if limit < 0 {
		limit = len(r.b) - r.i
	}
	end := r.i + limit
	if end > len(r.b) {
		end = len(r.b)
	}
	for j := r.i; j < end; j++ {
		if r.b[j] == 0 {
			s := string(r.b[r.i:j])
			n := j - r.i + 1
			r.i = j + 1
			return s, n, nil
		}
	}
	return "", 0, r.err("cstr")
}
