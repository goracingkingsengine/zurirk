package engine

import (
	"sync"
	"time"
)

const (
	defaultMovesToGo = 30 // default number of more moves expected to play
	infinite         = 1000000000 * time.Second
	overhead         = 20 * time.Millisecond
)

// atomicFlag is an atomic bool that can only be set.
type atomicFlag struct {
	lock sync.Mutex
	flag bool
}

func (af *atomicFlag) set() {
	af.lock.Lock()
	af.flag = true
	af.lock.Unlock()
}

func (af *atomicFlag) get() bool {
	af.lock.Lock()
	tmp := af.flag
	af.lock.Unlock()
	return tmp
}

// TimeControl is a time control that tries to split the
// remaining time over MovesToGo.
type TimeControl struct {
	WTime, WInc time.Duration // time and increment for white.
	BTime, BInc time.Duration // time and increment for black
	Depth       int32         // maximum depth search (including)
	MovesToGo   int           // number of remaining moves

	sideToMove Color
	time, inc  time.Duration // time and increment for us
	limit      time.Duration

	predicted bool       // true if this move was predicted
	branch    int        // branching factor
	currDepth int32      // current depth searched
	stopped   atomicFlag // true to stop the search
	ponderhit atomicFlag // true if ponder was successful

	searchTime     time.Duration // alocated time for this move
	searchDeadline time.Time     // don't go to the next depth after this deadline
	stopDeadline   time.Time     // abort search after this deadline
}

// NewTimeControl returns a new time control with no time limit,
// no depth limit, zero time increment and zero moves to go.
func NewTimeControl(pos *Position, predicted bool) *TimeControl {
	// Branch more when there are more pieces. With fewer pieces
	// there is less mobility and hash table kicks in more often.
	branch := 2
	for np := (pos.ByColor[White] | pos.ByColor[Black]).Count(); np > 0; np /= 6 {
		branch++
	}

	return &TimeControl{
		WTime:      infinite,
		WInc:       0,
		BTime:      infinite,
		BInc:       0,
		Depth:      64,
		MovesToGo:  defaultMovesToGo,
		sideToMove: pos.SideToMove,
		predicted:  predicted,
		branch:     branch,
	}
}

// NewFixedDepthTimeControl returns a TimeControl which limits the search depth.
func NewFixedDepthTimeControl(pos *Position, depth int32) *TimeControl {
	tc := NewTimeControl(pos, false)
	tc.Depth = depth
	tc.MovesToGo = 1
	return tc
}

// NewDeadlineTimeControl returns a TimeControl corresponding to a single move before deadline.
func NewDeadlineTimeControl(pos *Position, deadline time.Duration) *TimeControl {
	tc := NewTimeControl(pos, false)
	tc.WTime = deadline
	tc.BTime = deadline
	tc.MovesToGo = 1
	return tc
}

// thinkingTime calculates how much time to think this round.
// t is the remaining time, i is the increment.
func (tc *TimeControl) thinkingTime() time.Duration {
	// The formula allows engine to use more of time in the begining
	// and rely more on the increment later.
	tmp := time.Duration(tc.MovesToGo)
	tt := (tc.time + (tmp-1)*tc.inc) / tmp

	if tt < 0 {
		return 0
	}
	if tc.predicted {
		tt = tt * 4 / 3
	}
	if tt < tc.limit {
		return tt
	}
	return tc.limit
}

// Start starts the timer.
// Should start as soon as possible to set the correct time.
func (tc *TimeControl) Start(ponder bool) {
	if tc.sideToMove == White {
		tc.time, tc.inc = tc.WTime, tc.WInc
	} else {
		tc.time, tc.inc = tc.BTime, tc.BInc
	}

	// Calcuates the last moment when the search should be stopped.
	if tc.time > 2*overhead {
		tc.limit = tc.time - overhead
	} else if tc.time > overhead {
		tc.limit = overhead
	} else {
		tc.limit = tc.time
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	for i := 4; i > 0; i /= 2 {
		if tc.MovesToGo <= i {
			tc.branch++
		}
	}

	tc.stopped = atomicFlag{flag: false}
	tc.ponderhit = atomicFlag{flag: !ponder}

	tc.searchTime = tc.thinkingTime()
	tc.updateDeadlines() // deadlines are ignored while pondering (ponderHit == false)
}

func (tc *TimeControl) updateDeadlines() {
	now := time.Now()
	tc.searchDeadline = now.Add(tc.searchTime / time.Duration(tc.branch))

	// stopDeadline is when to abort the search in case of an explosion.
	// We give a large overhead here so the search is not aborted very often.
	deadline := tc.searchTime * 4
	if deadline > tc.limit {
		deadline = tc.limit
	}
	tc.stopDeadline = now.Add(deadline)
}

// NextDepth returns true if search can start at depth.
// In any case Stopped() will return false.
func (tc *TimeControl) NextDepth(depth int32) bool {
	tc.currDepth = depth
	return tc.currDepth <= tc.Depth && !tc.hasStopped(tc.searchDeadline)
}

// PonderHit switch to our time control.
func (tc *TimeControl) PonderHit() {
	tc.updateDeadlines()
	tc.ponderhit.set()
}

// Stop marks the search as stopped.
func (tc *TimeControl) Stop() {
	tc.stopped.set()
}

func (tc *TimeControl) hasStopped(deadline time.Time) bool {
	if tc.currDepth <= 2 {
		// Run for at few depths at least otherwise mates can be missed.
		return false
	}
	if tc.stopped.get() {
		// Use a cached value if available.
		return true
	}
	if tc.ponderhit.get() && time.Now().After(deadline) {
		// Stop search if no longer pondering and deadline as passed.
		return true
	}
	return false
}

// Stopped returns true if the search has stopped because
// Stop() was called or the time has ran out.
func (tc *TimeControl) Stopped() bool {
	if !tc.hasStopped(tc.stopDeadline) {
		return false
	}
	// Time has ran out so flip the stopped flag.
	tc.stopped.set()
	return true
}
