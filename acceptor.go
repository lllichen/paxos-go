package paxos

import (
	"log"
	"time"
)

type acceptor struct {
	id       int
	learners []int

	accept   message
	promised promise

	nt network
}

func newAcceptor(id int, nt network, learners ...int) *acceptor {
	return &acceptor{id: id, nt: nt, promised: message{}, learners: learners}
}

func (a *acceptor) run() {
	for {
		m, ok := a.nt.recv(time.Hour)
		if !ok {
			continue
		}
		switch m.typ {
		case Propose:
			accepted := a.receivePropose(m)
			//send msg to learners
			if accepted {
				for _, l := range a.learners {
					m := a.accept
					m.from = a.id
					m.to = l
					a.nt.send(m)
				}
			}
		case Prepare:
			promise, ok := a.receivePrepare(m)
			if ok {
				a.nt.send(promise)
			}
		default:
			log.Panicf("acceptor: %d unexpected message type: %+v", a.id, m.typ)
		}
	}
}

// If an acceptor receives a prepare request with number  greater
// than that of any prepare request to which it has already responsed,
// the it responds to the request whth a promise not to accept any more
// proposals number less than n and with the highest-numbered proposal
//(if any) that it has accepted

func (a *acceptor) receivePrepare(prepare message) (message, bool) {
	if a.promised.number() >= prepare.number() {
		log.Printf("acceptor: %d [promised: %+v] ignore prepare %+v", a.id, a.promised, prepare)
		return message{}, false
	}
	log.Printf("aceptor: %d [promised: %+v] promised %+v", a.id, a.promised, prepare)
	a.promised = prepare
	m := message{
		typ:  Promise,
		from: a.id, to: prepare.from,
		n: a.promised.number(),
		//previously accepted proposal
		prevn: a.accept.n, value: a.accept.value,
	}
	return m, true
}

//If an acceptor reveives an accept request for a proposal numbered
//n, it accept the proposal unless it has already responded to a prepare
//request having a number greater than n.
func (a *acceptor) receivePropose(propose message) bool {
	if a.promised.number() > propose.number() {
		log.Printf("acceptor: %d [promised: %+v] ignored proposal %+v", a.id, a.promised, propose)
		return false
	}
	if a.promised.number() < propose.number() {
		log.Panicf("accepted: %d receive unexpected proposal %+v", a.id, propose)
	}
	log.Printf("accepted: %d [promised: %+v,accepted: %+v] accepted proposal %+v", a.id, a.promised, a.accept, propose)
	a.accept = propose
	a.accept.typ = Accept
	return true
}

func (a *acceptor) restart() {}

func (a *acceptor) delay() {}
