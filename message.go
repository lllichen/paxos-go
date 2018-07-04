package paxos

type msgType int

//msgType
const (
	//1
	Prepare msgType = iota + 1
	//2 propose propose
	Propose
	//3 acceptor promise for prepare
	Promise
	//4 acceptor accept for propose
	Accept
)

type message struct {
	from, to int
	typ      msgType
	// message number
	n     int
	prevn int
	value string
}

func (m message) number() int {
	return m.n
}

func (m message) proposalValue() string {
	switch m.typ {
	case Promise, Accept:
		return m.value
	default:
		panic("unexpected proposalV")
	}
}

func (m message) proposalNumber() int {
	switch m.typ {
	case Promise:
		return m.prevn
	case Accept:
		return m.n
	default:
		panic("unexpected proposalN")
	}
}

type promise interface {
	number() int
}

type accept interface {
	proposalValue() string
	proposalNumber() int
}
