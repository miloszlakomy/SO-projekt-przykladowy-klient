package comm

import "os"
import "log"

const StickCap = 5

type Game struct {
	Srv *Server

	// world state:
	Wd      WorldDesc
	Islands map[Pos]*IslandInfo
	Water   map[Pos]bool
	Men     map[int]ManInfo
}

// FIXME: add staleness to IIs

func (g *Game) Init() error {
	if g.Islands == nil {
		g.Islands = make(map[Pos]*IslandInfo)
	}
	if g.Water == nil {
		g.Water = make(map[Pos]bool)
	}
	if g.Men == nil {
		g.Men = make(map[int]ManInfo)
	}
	var err error
	wd, err := g.Srv.GetWorldDesc()
	if err != nil {
		return err
	}
	if g.Wd.EdgeLength != 0 && wd.TurnsLeft > g.Wd.TurnsLeft { // if not first time around and more turns than previously
		log.Printf("New game")
		os.Exit(0)
	}
	g.Wd = wd
	ids, err := g.Srv.ListMen()
	if err != nil {
		return err
	}
	for _, id := range ids {
		mi, fis, err := g.Srv.GetManInfo(id)
		if err != nil {
			return err
		}
		g.Men[id] = mi
		for _, fi := range fis {
			if !fi.Land {
				g.Water[fi.Pos] = true
			}
		}
		iis, err := g.Srv.ListWood(id)
		if err != nil {
			if err1, ok := err.(RemoteError); ok && err1.Code == CodeNotLand {
				continue
			}
			return err
		}
		for _, ii := range iis {
			tmp := ii
			g.Islands[ii.Pos] = &tmp
			g.Water[ii.Pos] = false
		}
		for x := mi.Pos.X - 8; x < mi.Pos.X + 8; x++ {
			if x <= 0 || x > g.Wd.EdgeLength {
				continue
			}
			for y := mi.Pos.Y - 8; y < mi.Pos.Y + 8; y ++ {
				if y <= 0 || y > g.Wd.EdgeLength {
					continue
				}
				if _, ok := g.Water[Pos{x, y}]; !ok {
					g.Water[Pos{x, y}] = true
				}
			}
		}
	}
	return nil
}
