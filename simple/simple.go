package simple

import "ognisko/comm"
import "fmt"
import "log"
import "sync"
import "flag"
import "io/ioutil"
import "path/filepath"
import "encoding/json"
import "os"

var StatePath = flag.String("state", "", "")

var _ = fmt.Printf

type ManStatus struct {
	LastIsland         comm.Pos
	CurrentDestination comm.Pos
}

type MenStatusArr map[int]*ManStatus

func (w MenStatusArr) MarshalJSON() ([]byte, error) {
	m := make(map[string]*ManStatus, len(w))
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

func (w *MenStatusArr) UnmarshalJSON(buf []byte) error {
	var m map[string]*ManStatus
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	*w = make(MenStatusArr)
	for k, v := range m {
		var kk int
		if err := json.Unmarshal([]byte(k), &kk); err != nil {
			return err
		}
		(*w)[kk] = v
	}
	return nil
}




type Simple struct {
	mu   sync.Mutex `json:"-"`
	Game *comm.Game
	Men  MenStatusArr
}

func NewSimple() *Simple {
	c, err := comm.NewConn()
	if err != nil {
		panic(err)
	}

	s := &comm.Server{Conn: c}
	simp := &Simple{
		Game: &comm.Game{Srv: s},
		Men:  make(map[int]*ManStatus),
	}
	if *StatePath == "" {
		return simp
	}
	file, err := os.Open(*StatePath)
	if err != nil {
		log.Printf("Can't open state file: %s", err.Error())
		return simp
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(&simp)
	if err != nil {
		log.Printf("Error unmarshalling: %s", err.Error())
		//fixme: delete and abort?
	}
	return simp
}

func (simp* Simple) saveState() {
	if *StatePath == "" {
		return
	}
	file, err := ioutil.TempFile(filepath.Dir(*StatePath), filepath.Base(*StatePath))
	if err != nil {
		log.Printf("Cannot create a temp file: %s", err.Error())
		return
	}
	defer os.Remove(file.Name())

	enc := json.NewEncoder(file)
	err = enc.Encode(*simp)
	if err == nil {
		err = file.Close()
	}
	if err != nil {
		log.Printf("Error serializing state: %s", err.Error())
		return
	}

	if err := os.Rename(file.Name(), *StatePath); err != nil {
		log.Printf("Error overwriting old serialized file: %s", err.Error())
	}
}

func (simp *Simple) Loop() {
	for {
		simp.mu.Lock()
		simp.oneStep()
		simp.saveState()
		simp.mu.Unlock()
		simp.Game.Srv.Wait()
	}
}

func (simp *Simple) myBonfire(id int) comm.Location {
	mi := simp.Game.Men[id]
	r := simp.Game.Wd.BiggestLocations[0]
	countCloser := 0
	for _, mmi := range simp.Game.Men {
		if mmi.Pos.Distance(r.Pos) <= mi.Pos.Distance(r.Pos) {
			countCloser++
		}
	}
	if countCloser < 2 {
		return r
	}
	for _, v := range simp.Game.Wd.BiggestLocations {
		if v.Sticks >= simp.Game.Wd.BonfireLimit / 2 && v.Pos.Distance(mi.Pos) <= r.Pos.Distance(mi.Pos) {
			r = v
		}
	}
	return r
}

func (simp *Simple) bonfireDirection(id int) comm.Pos {
	return simp.Game.Men[id].Pos.Direction(simp.myBonfire(id).Pos)
}

func (simp *Simple) hasGuard() bool {
	for _, mi := range simp.Game.Men {
		if mi.Role == comm.RoleGuard {
			return true
		}
	}
	return false
}

func (simp *Simple) findNext(id int) comm.Pos {
	var dest comm.Pos
	mi := simp.Game.Men[id]
	status := simp.Men[id] // exists, because oneStep()
	//bonfire := simp.Game.Wd.BiggestLocations[0]
	bonfire := simp.myBonfire(id)
	if float64(bonfire.Sticks + bonfire.Pos.Distance(mi.Pos)) > 0.5*float64(simp.Game.Wd.BonfireLimit) && status.LastIsland != bonfire.Pos && mi.Pos != bonfire.Pos {
		return bonfire.Pos
	}
	//if float64(bonfire.Sticks) > 0.1*float64(simp.Game.Wd.BonfireLimit) && status.LastIsland != bonfire.Pos && bonfire.Pos.Distance(mi.Pos) < 20 {
	//	return bonfire.Pos
	//}
	mult := float64(0.8)
	//if bonfire.Sticks < simp.Game.Wd.BonfireLimit / 2 {
	//	mult = 0
	//}
	minDist := float64((mult + 1) * 1e12)
	for _, fi := range simp.Game.Islands {
		if fi.Pos == mi.Pos {
			continue
		}
		// if i have nothing, i need to go somewhere i can get something
		if (fi.Sticks == 0 || simp.Game.Guarded[fi.Pos]) && mi.StickCount == 0 {
			continue
		}
		if mi.Role != comm.RoleCaptain && fi.Sticks + mi.StickCount >= 100 && !simp.Game.Guarded[fi.Pos] && mi.Pos.Distance(fi.Pos) < 30 {
			return fi.Pos
		}
		if (fi.Sticks == fi.MySticks || simp.Game.Guarded[fi.Pos]) && bonfire.Sticks < simp.Game.Wd.BonfireLimit / 3 {
			continue // not always!
		}
		if status.LastIsland == fi.Pos {
			continue
		}
		penalty := mult * float64(fi.Pos.Distance(bonfire.Pos) - mi.Pos.Distance(bonfire.Pos))
		dist := float64(mi.Pos.Distance(fi.Pos)) + penalty
		if dist < minDist {
			minDist = dist
			dest = fi.Pos
		}
	}
	return dest
}

func (simp *Simple) oneStep() {
	if err := simp.Game.Init(); err != nil {
		if err == comm.ErrNewGame {
			if *StatePath != "" {
				os.Remove(*StatePath)
			}
			os.Exit(0)
		}
		panic(err)
	}
	for id, mi := range simp.Game.Men {
		status := simp.Men[id]
		if status == nil {
			status = new(ManStatus)
			simp.Men[id] = status
		}
		if mi.Role == comm.RoleGuard && mi.Pos != simp.Game.Wd.BiggestLocations[0].Pos { // guarding will be funny with multiple bonfires
			simp.Game.Srv.StopGuard(id)
		}
		field := simp.Game.Islands[mi.Pos]
		if field != nil {
			status.CurrentDestination = comm.Pos{}
		}
		if field != nil && mi.StickCount > 0 {
			if mi.Pos != status.LastIsland {
				simp.Game.Srv.Drop(id)
				status.LastIsland = comm.Pos{}
				field.Sticks += mi.StickCount // hack
				mi.StickCount = 0 // hack: we've just dropped and don't know that yet
			}
		}
		if field != nil && field.Sticks >= simp.Game.Wd.BonfireLimit {
			simp.Game.Srv.Ignite(id)
			continue
		}
		if field != nil && field.Sticks >= 100 && !simp.Game.Guarded[mi.Pos] && mi.Role != comm.RoleCaptain {
			simp.Game.Srv.Build(id)
			continue
		}
		if mi.Pos == simp.Game.Wd.BiggestLocations[0].Pos && !simp.hasGuard() && mi.Role != comm.RoleCaptain && simp.Game.Wd.BiggestLocations[0].Sticks > 200 {
			simp.Game.Srv.Guard(id)
			continue
		}
		if field != nil && field.Sticks > field.MySticks && mi.StickCount < mi.Cap() && !simp.Game.Guarded[mi.Pos] {
			if float64(field.Sticks) > 0.8 * float64(simp.Game.Wd.BonfireLimit) && float64(field.MySticks) > 0.4 * float64(field.Sticks) {
				// don't take anything away
			} else {
				simp.Game.Srv.Take(id)
			}
			// set lastisland anyway to advance
			status.LastIsland = mi.Pos
			continue
		}
/*		var dest *comm.IslandInfo
		for _, fi := range simp.Game.Islands {
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
		}*/
		if status.CurrentDestination == (comm.Pos{}) {
			status.CurrentDestination = simp.findNext(id)
		}
		if status.CurrentDestination != (comm.Pos{0, 0}) {
			dir := mi.Pos.Direction(status.CurrentDestination)
			log.Printf("Man %d going from %v to %v, in direction %v.\n", id, mi.Pos, status.CurrentDestination, dir)
			simp.Game.Srv.Move(id, dir.X, dir.Y)
		} else {
			log.Printf("Man %d doesn't have anywhere to go.\n", id)
		}
	}
}
