package docs

import (
	"html/template"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
)

const htmlDocumentTemplate = `
<!DOCTYPE html>
<html>

<head>
  <title>{{.Title}}</title>
</head>

<body>
	<h1>{{.Title}}</h1>
	<p>{{.Description}}<p>
	
	{{range .Sections}}
		<h2><a id="{{.TitleID}}"></a>{{.Title}}</h2>
		{{.Description}}
		
		{{range .HTMLIndicators}}
			<h3><a id="{{.Name}}"></a>{{.Title}}</h3>
			{{.HTML}}
		{{- end}}
	{{- end}}
</body>

</html>
`

var htmlDocumatationTmpl = template.Must(template.New("html_document").Parse(htmlDocumentTemplate))

func DocumentToHTML(d v1alpha1.IndicatorDocument) (string, error) {
	return docToTemplate(d, htmlDocumatationTmpl)
}
