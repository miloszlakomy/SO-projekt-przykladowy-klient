package main

import "ognisko/comm"
import "flag"
import "fmt"

var _ = fmt.Printf

func main() {
	flag.Parse()

	c, err := comm.NewConn()
	if err != nil {
		panic(err)
	}

	s := &comm.Server{Conn: c}
	game := &comm.Game{Srv: s}
	islandTaken := map[int]comm.Pos{}
	for {
		if err := game.Init(); err != nil {
			panic(err)
		}
		for id, mi := range game.Men {
			field := game.Islands[mi.Pos]
			if mi.StickCount > 0 {
				if v, ok := islandTaken[id]; !ok || mi.Pos != v {
					game.Srv.Drop(id)
					delete(islandTaken, id)
				}
			}
			if field != nil && field.Sticks > field.MySticks && mi.StickCount < 5 {
				game.Srv.Take(id)
				islandTaken[id] = mi.Pos
				continue
			}
			if field == nil || field.Sticks == field.MySticks {
				minDist := 10000000
				dir := comm.Pos{0, 0}
				var dest *comm.IslandInfo
				for _, fi := range game.Islands {
					if fi.Pos == mi.Pos {
						continue
					}
					if fi.Sticks == fi.MySticks {
						continue
					}
					if v, ok := islandTaken[id]; ok && v == fi.Pos {
						continue
					}
					if fi.Pos.Distance(mi.Pos) < minDist {
						dir = mi.Pos.Direction(fi.Pos)
						minDist = mi.Pos.Distance(fi.Pos)
						dest = fi
					}
				}
				if dir != (comm.Pos{0, 0}) {
					fmt.Printf("Man %d going from %v to %v, in direction %v.\n", id, mi.Pos, dest.Pos, dir)
					game.Srv.Move(id, dir.X, dir.Y)
				}
			}
		}
		game.Srv.Wait()
	}
}
