package docs

const htmlDocumentTemplate = `
---
title: {{.Title}}
owner: {{.Owner}}
---

{{.Description}}

{{range .Sections}}
## <a id="{{.TitleID}}"></a>{{.Title}}</h2>
{{.Description}}

{{range .HTMLIndicators}}
<div>
	{{.}}
</div>
{{end}}


{{range .HTMLMetrics}}
<div>
	{{.}}
</div>
{{end}}
		
</div>
{{end}}
`

const htmlIndicatorTemplate = `### <a id="{{.TitleID}}"></a>{{.Title}}
<table>
    <tr><th colspan="2" style="text-align: center;"><br/> {{range .MetricRefs}}{{.Name}}<br/>{{end}}<br/></th></tr>
    <tr>
        <th width="25%">Description</th>
        <td>
            {{.Description}}
        </td>
    </tr>
    <tr>
        <th>PromQL</th>
        <td>{{.PromQL}}</td>
    </tr>
    <tr>
        <th>Recommended measurement</th>
        <td>{{.Measurement}}</td>
    </tr>
    <tr>
        <th>Recommended alert thresholds</th>
        <td>
            {{range .Thresholds}} <em>{{.Level}}</em>: {{.OperatorAndValue}}<br/> {{end}}
        </td>
    </tr>
    <tr>
        <th>Recommended response</th>
        <td>
            {{.Response}}
        </td>
    </tr>
</table>`

const htmlMetricTemplate = `### <a id="{{.TitleID}}"></a>{{.Title}}
<table>
   <tbody><tr><th colspan="2" style="text-align: center;"><br> {{.Name}}<br><br></th></tr>
   <tr>
      <th width="25%">Description</th>
      <td>
		{{.Description}}
		<span><strong>Firehose Origin</strong>: {{.Origin}}</span>
		<span><strong>Log Cache Source ID</strong>: {{.SourceID}}</span>
      </td>
   </tr>
</tbody></table>`
