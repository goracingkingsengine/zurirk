// +build !coach

package engine

const disableCache = false

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
}

// Eval is a sum of scores.
type Eval struct {
	M, E int32 // mid game, end game
}

func (e *Eval) Feed(phase int32) int32 {
	return (e.M*(256-phase) + e.E*phase) / 256
}

func (e *Eval) Merge(o Eval) {
	e.M += o.M
	e.E += o.E
}

func (e *Eval) Add(s Score) {
	e.M += s.M
	e.E += s.E
}

func (e *Eval) AddN(s Score, n int32) {
	e.M += s.M * n
	e.E += s.E * n
}

func (e *Eval) Neg() {
	e.M = -e.M
	e.E = -e.E
}

func initWeights() {
}
