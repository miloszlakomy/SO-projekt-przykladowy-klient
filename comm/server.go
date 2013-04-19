package comm

import "net/textproto"
import "fmt"
import "strings"

type Server struct {
	*Conn
	pipeline textproto.Pipeline
}

type Location struct {
	Pos Pos
	Sticks int
}

type WorldDesc struct {
	EdgeLength   int
	IslandCount  int
	BonfireLimit int
	BonfireCoeff int
	MoveTime     int // in seconds
	ResultCoeff  float64

	TurnsLeft int
	Fire bool
	BiggestLocations [5]Location
}

const (
	RoleNone    = "NONE"
	RoleGuard   = "GUARD"
	RoleBuilder = "BUILDER"
	RoleCaptain = "CAPTAIN"
)

const (
	CodeNotLand = 104
)

type FieldInfo struct {
	Land bool
	Pos  Pos
	//	PeopleCount int
	//	GuardCount int
	//	OwnedRafts int
	//	AbandonedRafts int
	//	OwnedRaftsSticks int
	//	AbandonedRaftsSticks int
}

type ManInfo struct {
	Pos Pos
	Role       string
	StickCount int
	BusyFor    int
}

func (srv *Server) GetManInfo(id int) (ManInfo, []FieldInfo, error) {
	var mi ManInfo
	var fis []FieldInfo
	err := srv.cmd(func() error {
		s, err := srv.ReadRawLine()
		if err != nil {
			return err
		}
		var busy string
		if _, err := fmt.Sscanf(s, "%d %d %d %s %s", &mi.Pos.X, &mi.Pos.Y, &mi.StickCount, &busy, &mi.Role); err != nil {
			return err
		}
		if busy == "UNKNOWN" {
			mi.BusyFor = -1
		} else {
			if _, err := fmt.Sscanf(busy, "%d", &mi.BusyFor); err != nil {
				return err
			}
		}
		for i := 0; i < 5; i++ {
			s, err := srv.ReadRawLine()
			if err != nil {
				return err
			}
			if s == "NIL" {
				continue
			}
			var fieldTy string
			var deltaX int
			var deltaY int
			var dummy int
			if _, err := fmt.Sscanf(s, "%s %d %d %d %d %d %d %d", &fieldTy, &deltaX, &deltaY, &dummy, &dummy, &dummy, &dummy, &dummy); err != nil {
				return err
			}
			fi := FieldInfo{
				Land: fieldTy == "LAND",
				Pos: Pos{
					X: mi.Pos.X + deltaX,
					Y: mi.Pos.Y + deltaY,
				},
			}
			fis = append(fis, fi)
		}
		return nil
	}, "INFO %d", id)
	return mi, fis, err
}

type IslandInfo struct {
	Pos Pos
	Sticks   int
	MySticks int
}

func (s *Server) cmd(result func() error, format string, a ...interface{}) error {
	id := s.pipeline.Next()
	s.pipeline.StartRequest(id)
	if err := s.Printf(format, a...); err != nil {
		panic(err)
		//		s.pipeline.EndRequest(id)
		//		s.pipeline.StartResponse(id)
		//		s.pipeline.EndResponse(id)
		//		return err
	}
	s.pipeline.EndRequest(id)

	s.pipeline.StartResponse(id)
	defer s.pipeline.EndResponse(id)
	if err := s.ReadResult(); err != nil {
		return err
	}

	return result()
}

func (srv *Server) Wait() error {
	return srv.cmd(func() error {
		if _, err := srv.ReadRawLine(); err != nil {
			return err
		}
		return srv.ReadResult()
	}, "WAIT")
}

func (srv *Server) ListMen() ([]int, error) {
	var ret []int
	err := srv.cmd(func() error {
		s, err := srv.ReadRawLine()
		if err != nil {
			return err
		}
		var count int
		_, err = fmt.Sscanf(s, "%d", &count)
		if err != nil {
			return err
		}
		ret = make([]int, count)
		s, err = srv.ReadRawLine()
		if err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
		ss := strings.Split(strings.Trim(s, " "), " ")
		if len(ss) != count {
			return fmt.Errorf("Expected %d values, got %d: [%s]", count, len(ss), s)
		}
		for i := 0; i < count; i++ {
			_, err = fmt.Sscanf(ss[i], "%d", &ret[i])
			if err != nil {
				return err
			}
		}
		return nil
	}, "LIST_SURVIVORS")
	return ret, err
}

func (srv *Server) ListWood(id int) ([]IslandInfo, error) {
	var ii []IslandInfo
	err := srv.cmd(func() error {
		s, err := srv.ReadRawLine()
		if err != nil {
			return err
		}
		var count int
		_, err = fmt.Sscanf(s, "%d", &count)
		if err != nil {
			return err
		}
		ii = make([]IslandInfo, count)
		for i := 0; i < count; i++ {
			s, err := srv.ReadRawLine()
			if err != nil {
				return err
			}
			_, err = fmt.Sscanf(s, "%d %d %d %d", &ii[i].Pos.X, &ii[i].Pos.Y, &ii[i].Sticks, &ii[i].MySticks)
			if err != nil {
				return err
			}
		}
		return nil
	}, "LIST_WOOD %d", id)
	return ii, err
}

var noResponse = func() error { return nil }

func (srv *Server) Take(id int) error {
	return srv.cmd(noResponse, "TAKE %d", id)
}

func (srv *Server) Drop(id int) error {
	return srv.cmd(noResponse, "GIVE %d", id)
}

func (srv *Server) Move(id int, deltaX int, deltaY int) error {
	return srv.cmd(noResponse, "MOVE %d %d %d", id, deltaX, deltaY)
}

func (srv *Server) GetWorldDesc() (WorldDesc, error) {
	wd := WorldDesc{}
	err := srv.cmd(func() error {
		res, err := srv.ReadRawLine()
		if err != nil {
			return err
		}
		if _, err := fmt.Sscanf(res, "%d %d %d %d %d %f", &wd.EdgeLength, &wd.IslandCount, &wd.BonfireLimit, &wd.BonfireCoeff, &wd.MoveTime, &wd.ResultCoeff); err != nil {
			return err
		}
		return nil
	},"DESCRIBE_WORLD")
	if err != nil {
		return wd, err
	}
	err = srv.cmd(func() error {
		s, err := srv.ReadRawLine()
		if err != nil {
			return err
		}
		var fireString string
		if _, err := fmt.Sscanf(s, "%s %d", &fireString, &wd.TurnsLeft); err != nil {
			return err
		}
		wd.Fire = fireString == "BURNING"
		for i := 0; i < 5; i++ {
			s, err := srv.ReadRawLine()
			if err != nil {
				return err
			}
			if _, err := fmt.Sscanf(s, "%d %d %d", &wd.BiggestLocations[i].Pos.X, &wd.BiggestLocations[i].Pos.Y, &wd.BiggestLocations[i].Sticks); err != nil {
				return err
			}
		}
		return nil
	}, "TIME_TO_RESCUE")
	return wd, err
}
