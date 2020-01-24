package gossiper


type Best struct {
	From int    // Node the proposal is from (spoiler: -1 for tied tickets)
	Tkt  uint64 // Proposal's genetic fitness ticket
}

/* Define type of message inside QSC round */
type Type int 

const (
	Raw Type = iota
	Ack
	Wit
)
type QSCMessage struct {
	From int		// The source of msg
	Step int 		// Logical time step for msg
	Type Type		// Msg type: Raw, Ack, Wit
	Tkt uint64		// Fitness ticket
	QSC []Round		// QSC consensus state for rounds ending at Step or later
}

// Find the Best of two records primarily according to highest ticket number.
// For spoilers, detect and record ticket collisions with invalid node number.
func (b *Best) merge(o *Best, spoiler bool) {
	if o.Tkt > b.Tkt {
		*b = *o // strictly better ticket
	} else if o.Tkt == b.Tkt && o.From != b.From && spoiler {
		b.From = -1 // record ticket collision
	}
}

// Round encapsulates all the QSC state needed for one consensus round:
// the best potential "spoiler" proposal regardless of confirmation status,
// the best confirmed (witnessed) proposal we've seen so far in the round,
// and the best reconfirmed (double-witnessed) proposal we've seen so far.
// Finally, at the end of the round, we set Commit to true if
// the best confirmed proposal in Conf has definitely been committed.
type Round struct {
	Spoil  Best // Best potential spoiler(s) we've found so far
	Conf   Best // Best confirmed proposal we've found so far
	Reconf Best // Best reconfirmed proposal we've found so far
	Commit bool // Whether we confirm this round successfully committed
}

// Merge QSC round info from an incoming message into our round history
func mergeQSC(b, o []Round) {
	for i := range b {
		b[i].Spoil.merge(&o[i].Spoil, true)
		b[i].Conf.merge(&o[i].Conf, false)
		b[i].Reconf.merge(&o[i].Reconf, false)
	}
}

