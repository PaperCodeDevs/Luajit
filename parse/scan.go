package parse

type Hit struct {
	Off  int
	Size int
	Name string
	Err  error
}

func Scan(blob []byte) []Hit {
	var out []Hit
	for i := 0; i+4 <= len(blob); i++ {
		if !IsDump(blob[i:]) {
			continue
		}
		d, err := Parse(blob[i:])
		h := Hit{Off: i, Err: err}
		if err == nil {
			h.Size = d.Size
			h.Name = d.Name
			if d.Size > 1 {
				i += d.Size - 1
			}
		}
		out = append(out, h)
	}
	return out
}
