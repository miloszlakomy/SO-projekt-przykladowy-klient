package comm

import "os"
import "log"
import "encoding/json"

const StickCap = 5

type IslandsType map[Pos]*IslandInfo

func (w IslandsType) MarshalJSON() ([]byte, error) {
	m := make(map[string]*IslandInfo, len(w))
	for k, v := range w {
		kk, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		m[string(kk)] = v
	}
	r, err := json.Marshal(m)
	return r, err
}

func (w *IslandsType) UnmarshalJSON(buf []byte) error {
	var m map[string]*IslandInfo
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	*w = make(IslandsType)
	for k, v := range m {
		var kk Pos
		if err := json.Unmarshal([]byte(k), &kk); err != nil {
			return err
		}
		(*w)[kk] = v
	}
	return nil
}


type WaterType map[Pos]bool

func (w WaterType) MarshalJSON() ([]byte, error) {
	m := make(map[string]bool, len(w))
	for k, v := range w {
		kk, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		m[string(kk)] = v
	}
	r, err := json.Marshal(m)
	return r, err
}

func (w *WaterType) UnmarshalJSON(buf []byte) error {
	var m map[string]bool
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	*w = make(WaterType)
	for k, v := range m {
		var kk Pos
		if err := json.Unmarshal([]byte(k), &kk); err != nil {
			return err
		}
		(*w)[kk] = v
	}
	return nil
}

type MenType map[int]ManInfo

func (w MenType) MarshalJSON() ([]byte, error) {
	m := make(map[string]ManInfo, len(w))
	for k, v := range w {
		kk, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		m[string(kk)] = v
	}
	r, err := json.Marshal(m)
	return r, err
}

func (w *MenType) UnmarshalJSON(buf []byte) error {
	var m map[string]ManInfo
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	*w = make(MenType)
	for k, v := range m {
		var kk int
		if err := json.Unmarshal([]byte(k), &kk); err != nil {
			return err
		}
		(*w)[kk] = v
	}
	return nil
}


type Game struct {
	Srv *Server `json:"-"`

	// world state:
	Wd      WorldDesc
	Islands IslandsType
	Water   WaterType
	Men     MenType
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
