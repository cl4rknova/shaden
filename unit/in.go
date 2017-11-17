package unit

import (
	"fmt"
	"math"

	"buddin.us/shaden/dsp"
	"buddin.us/shaden/graph"
)

// InMode is a mode of processing of an In.
type InMode int

// InModes
const (
	Block InMode = iota
	Sample
)

const controlPeriod = 64

// In is a module input
type In struct {
	Name                      string
	Mode                      InMode
	constant, defaultConstant dsp.Valuer
	frame, constantFrame      []float64
	module                    *Unit
	source                    *Out
	node                      *graph.Node

	controlLastF float64
	controlLastI int
}

// Read reads a specific sample from the input frame
func (in *In) Read(i int) float64 {
	if isSourceControlRate(in) {
		return in.frame[0]
	}
	if in.Mode == Sample {
		size := len(in.frame)
		i = (i - 1 + size) % size
	}
	return in.frame[i]
}

// ReadSlow reads a specific sample from the input frame at a slow rate
func (in *In) ReadSlow(i int, f func(float64) float64) float64 {
	if i%controlPeriod == 0 {
		in.controlLastF = f(in.Read(i))
	}
	return in.controlLastF
}

// ReadSlowInt reads a specific sample from the input frame at a slow rate
func (in *In) ReadSlowInt(i int, f func(int) int) int {
	if i%controlPeriod == 0 {
		in.controlLastI = f(int(in.Read(i)))
	}
	return in.controlLastI
}

// Fill fills the internal frame with a specific constant value
func (in *In) Fill(v dsp.Valuer) {
	in.constant = v
	for i := range in.frame {
		in.frame[i] = v.Float64()
	}
}

// Couple assigns the internal frame of this input to the frame of an output; binding them together. This in-of-itself
// does not define the connection. That is controlled by the the Nodes and Graph.
func (in *In) Couple(out Output) {
	o := out.Out()
	in.source = o
	in.frame = o.frame
}

// HasSource returns whether or not we have an inbound connection
func (in *In) HasSource() bool {
	return in.source != nil
}

// Reset disconnects an input from an output (if a connection has been established) and fills the frame with the normal
// constant value
func (in *In) Reset() {
	in.source = nil
	in.frame = in.constantFrame
	in.Fill(in.defaultConstant)
}

// ExternalNeighborCount returns the count of neighboring nodes outside of the parent Unit
func (in *In) ExternalNeighborCount() int {
	return in.node.InNeighborCount()
}

func (in *In) setNormal(v dsp.Valuer) {
	in.defaultConstant = v
	in.Fill(v)
}

func (in *In) String() string {
	return fmt.Sprintf("%s/%s", in.module.ID, in.Name)
}

func isSourceControlRate(in *In) bool {
	return in.HasSource() && in.source.Rate() == RateControl
}

func ident(v float64) float64   { return v }
func minZero(v float64) float64 { return math.Max(v, 0) }
