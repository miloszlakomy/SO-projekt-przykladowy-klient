package simple

import "ognisko/comm"
import "fmt"
import "log"
import "sync"

var _ = fmt.Printf

type ManStatus struct {
	LastIsland         comm.Pos
	CurrentDestination comm.Pos
}

type Simple struct {
	mu   sync.Mutex
	game *comm.Game
	men  map[int]*ManStatus
}

func NewSimple() *Simple {
	c, err := comm.NewConn()
	if err != nil {
		panic(err)
	}

	s := &comm.Server{Conn: c}
	return &Simple{
		game: &comm.Game{Srv: s},
		men:  make(map[int]*ManStatus),
	}
}

func (simp *Simple) Loop() {
	for {
		simp.mu.Lock()
		simp.oneStep()
		simp.mu.Unlock()
		simp.game.Srv.Wait()
	}
}

func (simp *Simple) oneStep() {
	if err := simp.game.Init(); err != nil {
		panic(err)
	}
	for id, mi := range simp.game.Men {
		status := simp.men[id]
		if status == nil {
			status = new(ManStatus)
			simp.men[id] = status
		}
		field := simp.game.Islands[mi.Pos]
		if field != nil && mi.StickCount > 0 {
			if mi.Pos != status.LastIsland {
				simp.game.Srv.Drop(id)
				status.LastIsland = comm.Pos{}
			}
		}
		if field != nil && field.Sticks > field.MySticks && mi.StickCount < 5 {
			simp.game.Srv.Take(id)
			status.LastIsland = mi.Pos
			continue
		}
		minDist := 10000000
		dir := comm.Pos{0, 0}
		var dest *comm.IslandInfo
		for _, fi := range simp.game.Islands {
			if fi.Pos == mi.Pos {
				continue
			}
			if fi.Sticks == fi.MySticks {
				continue
			}
			if status.LastIsland == fi.Pos {
				continue
			}
			if fi.Pos.Distance(mi.Pos) < minDist {
				dir = mi.Pos.Direction(fi.Pos)
				minDist = mi.Pos.Distance(fi.Pos)
				dest = fi
			}
		}
		if dir != (comm.Pos{0, 0}) {
			log.Printf("Man %d going from %v to %v, in direction %v.\n", id, mi.Pos, dest.Pos, dir)
			simp.game.Srv.Move(id, dir.X, dir.Y)
			status.CurrentDestination = dest.Pos
		} else {
			log.Printf("Man %d doesn't have anywhere to go.\n", id)
			status.CurrentDestination = comm.Pos{}
		}
	}
}
