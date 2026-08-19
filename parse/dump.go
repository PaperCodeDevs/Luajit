package parse

import "fmt"

func Parse(raw []byte) (*Dump, error) {
	r := &reader{b: raw}
	d, err := readHeader(r)
	if err != nil {
		return nil, err
	}
	var ready []*Proto
	for {
		if r.i >= len(r.b) {
			return nil, r.err("truncated")
		}
		if r.b[r.i] == 0 {
			r.i++
			break
		}
		psz, err := r.uleb32()
		if err != nil {
			return nil, err
		}
		if psz == 0 {
			break
		}
		if psz > maxProto {
			return nil, r.err("proto size")
		}
		start := r.i
		end := start + int(psz)
		if end > len(r.b) {
			return nil, r.err("proto overflow")
		}
		pr := &reader{b: r.b[start:end]}
		p, err := readProto(pr, d.Flags, &ready)
		if err != nil {
			return nil, err
		}
		if pr.i != int(psz) {
			return nil, fmt.Errorf("luajit: proto len want %d got %d", psz, pr.i)
		}
		r.i = end
		ready = append(ready, p)
	}
	if len(ready) != 1 {
		return nil, fmt.Errorf("luajit: leftover proto %d", len(ready))
	}
	d.Main = ready[0]
	d.Size = r.i
	return d, nil
}

func DumpSize(raw []byte) (int, error) {
	d, err := Parse(raw)
	if err != nil {
		return 0, err
	}
	return d.Size, nil
}
