package simple

import "fmt"
import "runtime"
import "ognisko/comm"
import "net/http"
import "html/template"
import "bufio"

type MenView Simple

var menViewTempl = template.Must(template.New("menview").Parse(`
<html><head><title>Men</title></head><body>
<table><tr><td>ID</td><td>Pos</td><td>Last island</td><td>Destination</td></tr>
{{range $id, $val := .}}
<tr id="{{$id}}"><td>{{$id}}</td><td>{{if $val.Info}}{{$val.Info.Pos}}{{else}}Dead{{end}}</td><td>{{$val.Status.LastIsland}}</td><td>{{$val.Status.CurrentDestination}}</td></tr>
{{end}}
</table></body></html>`))

func (mv *MenView) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	simp := (*Simple)(mv)
	simp.mu.Lock()
	defer simp.mu.Unlock()

	type man struct {
		Status ManStatus
		Info *comm.ManInfo
	}

	men := map[int]man{}
	for id, status := range simp.men {
		var pmi *comm.ManInfo
		if mi, ok := simp.game.Men[id]; ok {
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
	for id, m := range simp.game.Men {
		men[m.Pos] = id
	}
	for x := 1; x <= simp.game.Wd.EdgeLength; x++ {
		// not sure if needed:
		simp.mu.Unlock()
		runtime.Gosched()
		simp.mu.Lock()
		wr.WriteString(`<tr>`)
		for y := 1; y <= simp.game.Wd.EdgeLength; y++ {
			water, known := simp.game.Water[comm.Pos{x, y}]
			var s string
			if !known {
				s = "?"
			} else if water {
				s = "_"
			} else {
				s = "X"
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






