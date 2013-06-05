package simple

import "fmt"
import "runtime"
import "ognisko/comm"
import "net/http"
import "html/template"
import "bufio"

type MenView Simple

var menViewTempl = template.Must(template.New("menview").Parse(`
<table><tr><td>ID</td><td>Pos</td><td>Last island</td><td>Destination</td><td>Role</td><td>BusyFor</td><td>StickCount</td></tr>
{{range $id, $val := .}}
<tr id="{{$id}}"><td>{{$id}}</td><td>{{if $val.Info}}{{$val.Info.Pos}}{{else}}Dead{{end}}</td><td>{{$val.Status.LastIsland}}</td><td>{{$val.Status.CurrentDestination}}</td>
{{if $val.Info}}<td>{{$val.Info.Role}}</td><td>{{$val.Info.BusyFor}}</td><td>{{$val.Info.StickCount}}</td>{{else}}<td/><td/><td/>{{end}}</tr>
{{end}}
</table>`))

func (mv *MenView) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	simp := (*Simple)(mv)
	simp.mu.Lock()
	defer simp.mu.Unlock()

	type man struct {
		Status ManStatus
		Info *comm.ManInfo
	}

	men := map[int]man{}
	for id, status := range simp.Men {
		var pmi *comm.ManInfo
		if mi, ok := simp.Game.Men[id]; ok {
			pmi = &mi
		}
		men[id] = man{Status: *status, Info: pmi}
	}
	menViewTempl.Execute(rw, men)
}

type MapView Simple

func (mv *MapView) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	simp := (*Simple)(mv)
    simp.mu.Lock()
    defer simp.mu.Unlock()
	wr := bufio.NewWriter(rw)

	wr.WriteString(`<html><head><title>Map</title></head><body><table>`)
	men := map[comm.Pos]int{}
	for id, m := range simp.Game.Men {
		men[m.Pos] = id
	}
	for x := 1; x <= simp.Game.Wd.EdgeLength; x++ {
		// not sure if needed:
		simp.mu.Unlock()
		runtime.Gosched()
		simp.mu.Lock()
		wr.WriteString(`<tr>`)
		for y := 1; y <= simp.Game.Wd.EdgeLength; y++ {
			water, known := simp.Game.Water[comm.Pos{x, y}]
			var s string
			if !known {
				s = "?"
			} else if water {
				s = "_"
			} else {
				ii := simp.Game.Islands[comm.Pos{x, y}]
				v := ii.Sticks - ii.MySticks
				v = (v + 9) / 10
				if v < 10 {
					s = string(rune('0' + v))
				} else if v < 36 {
					s = string(rune('A' + v - 10))
				} else {
					s = "*"
				}
			}
			if id, ok := men[comm.Pos{x, y}]; ok {
				s = fmt.Sprintf(`<a href="/men#%d">%s</a>`, id, s)
			}
			fmt.Fprintf(wr, `<td>%s</td>`, s)
		}
		wr.WriteString(`</tr>`)
	}
	wr.WriteString(`</table></body></html>`)
	wr.Flush()
}

type OverviewView Simple

var overviewViewTempl = template.Must(template.New("menview").Parse(`
<table>
<tr><td>Edge length</td><td>{{.Wd.EdgeLength}}</td></tr>
<tr><td>Islands</td><td>{{.Wd.IslandCount}}</td></tr>
<tr><td>Bonfire Limit</td><td>{{.Wd.BonfireLimit}}</td></tr>
<tr><td>Bonfire Coefficient</td><td>{{.Wd.BonfireCoeff}}</td></tr>
<tr><td>Move length [s]</td><td>{{.Wd.MoveTime}}</td></tr>
<tr><td>Result coefficient</td><td>{{.Wd.ResultCoeff}}</td></tr>
<tr><td>Turns left</td><td>{{.Wd.TurnsLeft}}</td></tr>
<tr><td>Fire?</td><td>{{.Wd.Fire}}</td></tr>
<tr><td>Stick points</td><td>{{.Wd.StickPoints}}</td></tr>
<tr><td>Marked sticks</td><td>{{.Wd.MarkedSticks}}</td></tr>
<tr><td>Held sticks</td><td>{{.Wd.HeldSticks}}</td></tr>
</table><p>Biggest locations:<ol>
{{range $val := .Wd.BiggestLocations}}
<li>{{$val.Sticks}} at ({{$val.Pos.X}}, {{$val.Pos.Y}})</li>
{{end}}
</ol>`))

func (ov *OverviewView) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	simp := (*Simple)(ov)
	simp.mu.Lock()
	defer simp.mu.Unlock()

	type man struct {
		Status ManStatus
		Info *comm.ManInfo
	}

	men := map[int]man{}
	for id, status := range simp.Men {
		var pmi *comm.ManInfo
		if mi, ok := simp.Game.Men[id]; ok {
			pmi = &mi
		}
		men[id] = man{Status: *status, Info: pmi}
	}
	overviewViewTempl.Execute(rw, simp.Game)
	menViewTempl.Execute(rw, men)

}






