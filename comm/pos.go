package comm

type Pos struct {
	X int
	Y int
}

func abs(x int) int {
	if x < 0 {
		x = -x
	}
	return x
}

func (p Pos) Distance(q Pos) int {
	return abs(p.X-q.X) + abs(p.Y-q.Y)
}

func (p Pos) Direction(q Pos) Pos {
	if p.X < q.X {
		return Pos{1, 0}
	} else if p.X > q.X {
		return Pos{-1, 0}
	} else if p.Y < q.Y {
		return Pos{0, 1}
	} else if p.Y > q.Y {
		return Pos{0, -1}
	} else {
		return Pos{0, 0}
	}
	panic("unreached")
}
