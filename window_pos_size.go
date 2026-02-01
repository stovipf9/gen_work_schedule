package main

type Pos struct {
	X int
	Y int
}

func (p *Pos) Normalize() {
	p.X = max(p.X, 0)
	p.Y = max(p.Y, 0)
}

type Size struct {
	W int
	H int
}

func (s *Size) Normalize() {
	s.W = max(s.W, 100)
	s.H = max(s.H, 100)
}

type WindowPosSize struct {
	Pos  Pos
	Size Size
}

func (w *WindowPosSize) Normalize() {
	w.Pos.Normalize()
	w.Size.Normalize()
}
