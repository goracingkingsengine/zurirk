// +build coach

package engine

const disableCache = true

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
	I    int   // index in Weights
}

// Eval is a sum of scores.
type Eval struct {
	M, E   int32              // mid game, end game
	Values [len(Weights)]int8 // input values
}

func (e *Eval) Feed(phase int32) int32 {
	return (e.M*(256-phase) + e.E*phase) / 256
}

func (e *Eval) Merge(o Eval) {
	e.M += o.M
	e.E += o.E
	for i := range o.Values {
		e.Values[i] += o.Values[i]
	}
}

func (e *Eval) Add(s Score) {
	e.M += s.M
	e.E += s.E
	e.Values[s.I] += 1
}

func (e *Eval) AddN(s Score, n int32) {
	e.M += s.M * n
	e.E += s.E * n
	e.Values[s.I] += int8(n)
}

func (e *Eval) Neg() {
	e.M = -e.M
	e.E = -e.E
	for i, v := range e.Values {
		e.Values[i] = -v
	}
}

func initWeights() {
	for i := range Weights {
		Weights[i].I = i
	}
}
