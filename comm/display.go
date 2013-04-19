package comm

//import "net/http"
import "html/template"

const menTemplateSrc = `
<html><head><title>Men</title></head><body>
<table>
<th><td>ID</td><td>Pos</td><td>Sticks</td></th>
{{range .}}
<tr><td>{{.}}</td></tr>
{{end}}
</body></html>
`

var menTemplate = template.Must(template.New("men").Parse(menTemplateSrc))
