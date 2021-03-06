package paxos

import (
	"fmt"
	"log"
	"time"
)

type proposer struct {
	id int
	//stable
	lastSeq int

	value  string
	valueN int

	acceptors map[int]promise
	nt        network
}

func newProposer(id int, value string, nt network, acceptors ...int) *proposer {
	p := &proposer{id: id, nt: nt, lastSeq: 0, value: value, acceptors: make(map[int]promise)}
	for _, a := range acceptors {
		p.acceptors[a] = message{}
	}
	return p
}

func (p *proposer) run(value string) {
	var ok bool
	var m message

	//stage 1: do prepare until reach the majority
	for !p.majorityReached() {
		if !ok {
			ms := p.prepare()
			for i := range ms {
				log.Printf("proposer: %d send proposer %+v", p.id, ms[i])
				if len(value) > 0 {
					ms[i].value = value
				}
				p.nt.send(ms[i])
			}
		}
		m, ok = p.nt.recv(time.Second)
		if !ok {
			//the previous prepare is failed
			//contiue to do another prepare
			continue
		}
		switch m.typ {
		case Promise:
			p.receivePromise(m)
		default:
			log.Panicf("proposer: %d unexpected message type: %+v", p.id, m.typ)
		}
	}
	log.Printf("proposer: %d promise %d reached majority %d", p.id, p.n(), p.majority())

	//stage 2: do propose
	log.Printf("proposer: %d starts to proposer [%d: %s]", p.id, p.n(), p.value)
	ms := p.propose()
	for i := range ms {
		p.nt.send(ms[i])
	}
}

// if the proposer reveives the requested responses from a majority of
// the acceptors, then it can issue a proposal with number n and value
//v, where v is the value of the highest-numbered proposal among the
// responses, or is any value selected by the proposer of the responders
// reported no proposal.
func (p *proposer) propose() []message {
	ms := make([]message, p.majority())

	i := 0
	for to, promise := range p.acceptors {
		if promise.number() == p.n() {
			ms[i] = message{from: p.id, to: to, typ: Propose, n: p.n(), value: p.value}
			i++
		}
		if i == p.majority() {
			break
		}
	}
	return ms
}

//A proposer chooses a nw proposal number n and sends a request to
//each member of some set of acceptors, asking it to response with:
// (a) A promise never again to accept a proposal numbered less than n, adn
// (b) The proposal with the highest number less than n that is has accepted, if any.
func (p *proposer) prepare() []message {
	p.lastSeq++

	ms := make([]message, p.majority())
	i := 0

	// send Prepare to acceptors
	for to := range p.acceptors {
		ms[i] = message{from: p.id, to: to, typ: Prepare, n: p.n()}
		i++
		if i == p.majority() {
			break
		}
	}
	return ms
}

func (p *proposer) receivePromise(promise message) {
	prevPromise := p.acceptors[promise.from]

	log.Printf("proposer: %d prevPromise %+v", p.id, prevPromise)

	if prevPromise.number() < promise.number() {
		log.Printf("proposer: %d received a new promise %+v", p.id, promise)
		log.Printf("proposer: %d current proposer is %+v", p.id, p)
		//save promise
		p.acceptors[promise.from] = promise

		//update value to the value with a larger N
		if promise.proposalNumber() > p.valueN {
			log.Printf("proposer: %d updated the value [%s] to [%s]", p.id, p.value, promise.proposalValue())
			p.valueN = promise.proposalNumber()
			p.value = promise.proposalValue()
		}
	}
}

func (p *proposer) majority() int {
	return len(p.acceptors)/2 + 1
}

//
func (p *proposer) majorityReached() bool {
	m := 0
	fmt.Println("wait ---------")
	for _, promise := range p.acceptors {
		log.Printf("promise number : %+v, p.n: %+v \n", promise.number(), p.n())
		if promise.number() == p.n() {
			m++
		}
	}
	if m >= p.majority() {
		return true
	}
	return false
}

//proposer's number
func (p *proposer) n() int {
	return p.lastSeq<<16 | p.id
}
