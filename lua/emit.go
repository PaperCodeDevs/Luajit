package lua

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaperCodeDevs/Luajit/op"
	"github.com/PaperCodeDevs/Luajit/parse"
)

func Decompile(raw []byte) (out []byte, err error) {
	defer func() {
		if x := recover(); x != nil {
			out = nil
			err = fmt.Errorf("luajit: panic %v", x)
		}
	}()
	d, e := parse.Parse(raw)
	if e != nil {
		return nil, e
	}
	src := Source(d)
	if src == "" {
		return nil, fmt.Errorf("luajit: empty")
	}
	return []byte(src), nil
}

func Source(d *parse.Dump) string {
	var b strings.Builder
	if d.Name != "" {
		fmt.Fprintf(&b, "-- %s\n", d.Name)
	}
	emitFn(&b, d, d.Main, 0, true)
	return b.String()
}

func emitFn(b *strings.Builder, d *parse.Dump, p *parse.Proto, indent int, top bool) {
	if p == nil {
		return
	}
	need := int(p.Frame) + 8
	if n := int(p.Params) + 8; n > need {
		need = n
	}
	if d.Flags&parse.FlagFR2 != 0 {
		need++
	}
	c := &gen{
		d:      d,
		p:      p,
		slot:   make([]string, need),
		indent: indent,
		out:    b,
		used:   map[*parse.Proto]bool{},
	}
	if d.Flags&parse.FlagFR2 != 0 {
		c.fr2 = 1
	}
	for i := range c.slot {
		c.slot[i] = "s" + strconv.Itoa(i)
	}
	for i := 0; i < int(p.Params); i++ {
		c.set(i, "a"+strconv.Itoa(i))
	}
	c.bindParams()
	if !top {
		c.line("function(%s)", strings.Join(c.params(), ", "))
		c.indent++
	}
	c.body(0, len(p.Ins))
	c.emitUnused()
	if !top {
		c.indent--
		c.line("end")
	}
}

type gen struct {
	d      *parse.Dump
	p      *parse.Proto
	slot   []string
	indent int
	fr2    int
	out    *strings.Builder
	skip   map[int]bool
	used   map[*parse.Proto]bool
}

func (c *gen) emitUnused() {
	if c.used == nil {
		c.used = map[*parse.Proto]bool{}
	}
	for i, k := range c.p.KGC {
		if k.Kind != parse.KChild || k.Child == nil || c.used[k.Child] {
			continue
		}
		var inner strings.Builder
		emitFn(&inner, c.d, k.Child, c.indent, false)
		c.line("local _c%d = %s", i, strings.TrimSpace(inner.String()))
	}
}

func (c *gen) params() []string {
	n := int(c.p.Params)
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = c.get(i)
	}
	if c.p.Flags&2 != 0 {
		out = append(out, "...")
	}
	return out
}

func (c *gen) bindParams() {
	for i := 0; i < int(c.p.Params); i++ {
		if n := c.p.SlotName(i, 0); okName(n) {
			c.set(i, n)
		}
	}
}

func (c *gen) localName(slot, pc int) string {
	if n := c.p.SlotName(slot, pc); okName(n) {
		return n
	}
	if n := c.p.SlotName(slot, pc+1); okName(n) {
		return n
	}
	return "s" + strconv.Itoa(slot)
}

func (c *gen) line(format string, args ...any) {
	c.out.WriteString(strings.Repeat("  ", c.indent))
	fmt.Fprintf(c.out, format, args...)
	c.out.WriteByte('\n')
}

func (c *gen) code(in parse.Ins) byte {
	return op.Norm(c.d.Version, in.Op)
}

func (c *gen) get(i int) string {
	if i < 0 || i >= len(c.slot) || c.slot[i] == "" {
		return "s" + strconv.Itoa(i)
	}
	return c.slot[i]
}

func (c *gen) set(i int, expr string) {
	for i >= len(c.slot) {
		c.slot = append(c.slot, "s"+strconv.Itoa(len(c.slot)))
	}
	c.slot[i] = expr
}

func (c *gen) gcstr(d uint16, cc byte) string {
	key := c.p.GCKey(d, cc)
	s := c.p.Str(key)
	if s == "" {
		return "k" + strconv.Itoa(int(key))
	}
	return quote(s)
}

func (c *gen) gcname(d uint16, cc byte) string {
	key := c.p.GCKey(d, cc)
	s := c.p.Str(key)
	if s != "" && isIdent(s) {
		return s
	}
	if s != "" {
		return "_G[" + quote(s) + "]"
	}
	return "_G[k" + strconv.Itoa(int(key)) + "]"
}
