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

var statePath = flag.String("state", "", "")

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
	if *statePath == "" {
		return simp
	}
	file, err := os.Open(*statePath)
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
	if *statePath == "" {
		return
	}
	file, err := ioutil.TempFile(filepath.Dir(*statePath), filepath.Base(*statePath))
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

	if err := os.Rename(file.Name(), *statePath); err != nil {
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

func (simp *Simple) oneStep() {
	if err := simp.Game.Init(); err != nil {
		panic(err)
	}
	for id, mi := range simp.Game.Men {
		status := simp.Men[id]
		if status == nil {
			status = new(ManStatus)
			simp.Men[id] = status
		}
		field := simp.Game.Islands[mi.Pos]
		if field != nil && mi.StickCount > 0 {
			if mi.Pos != status.LastIsland {
				simp.Game.Srv.Drop(id)
				status.LastIsland = comm.Pos{}
			}
		}
		if field != nil && field.Sticks > field.MySticks && mi.StickCount < 5 {
			simp.Game.Srv.Take(id)
			status.LastIsland = mi.Pos
			continue
		}
		minDist := 10000000
		dir := comm.Pos{0, 0}
		var dest *comm.IslandInfo
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
		}
		if dir != (comm.Pos{0, 0}) {
			log.Printf("Man %d going from %v to %v, in direction %v.\n", id, mi.Pos, dest.Pos, dir)
			simp.Game.Srv.Move(id, dir.X, dir.Y)
			status.CurrentDestination = dest.Pos
		} else {
			log.Printf("Man %d doesn't have anywhere to go.\n", id)
			status.CurrentDestination = comm.Pos{}
		}
	}
}
