package parse

const (
	VerStd = 0x02
	VerMW  = 0x90

	FlagBE    = 0x01
	FlagStrip = 0x02
	FlagFFI   = 0x04
	FlagFR2   = 0x08

	KChild   = 0
	KTab     = 1
	KI64     = 2
	KU64     = 3
	KComplex = 4
	KStr     = 5

	TabNil   = 0
	TabFalse = 1
	TabTrue  = 2
	TabInt   = 3
	TabNum   = 4
	TabStr   = 5

	JumpBias = 0x8000
)

type Dump struct {
	Version byte
	Flags   uint32
	Name    string
	Main    *Proto
	Size    int
}

type Proto struct {
	Flags     byte
	Params    byte
	Frame     byte
	UV        []uint16
	KGC       []KGC
	KNum      []KNum
	Ins       []Ins
	FirstLine uint32
	NumLine   uint32
	Lines     []uint32
	UVName    []string
	Var       []Local
}

type Local struct {
	Name  string
	Start uint32
	End   uint32
}

type Ins struct {
	Op byte
	A  byte
	B  byte
	C  byte
	D  uint16
}

func (in Ins) J() int {
	return int(in.D) - JumpBias
}

type KGC struct {
	Kind  int
	Str   string
	Tab   *Tab
	Child *Proto
	Lo    uint32
	Hi    uint32
	ILo   uint32
	IHi   uint32
}

type Tab struct {
	Array []TabK
	Hash  [][2]TabK
}

type TabK struct {
	Kind int
	Str  string
	I    int32
	Lo   uint32
	Hi   uint32
}

type KNum struct {
	IsInt bool
	I     int32
	Lo    uint32
	Hi    uint32
}

func IsMiniWorld(raw []byte) bool {
	return len(raw) >= 4 && raw[0] == 0x1b && raw[1] == 'L' && raw[2] == 'J' && raw[3] == VerMW
}

func IsDump(raw []byte) bool {
	return len(raw) >= 4 && raw[0] == 0x1b && raw[1] == 'L' && raw[2] == 'J' && (raw[3] == VerStd || raw[3] == VerMW)
}
