package comm

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
	g.Wd, err = g.Srv.GetWorldDesc()
	if err != nil {
		return err
	}
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
	}
	for _, id := range ids {
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
		}
		// add sure water based on prox to a man -- fixme
	}
	return nil
}
