package docs

const htmlDocumentTemplate = `---
title: {{.Title}}
owner: {{.Owner}}
---
{{.Description}}

{{range .Sections}}
## <a id="{{.TitleID}}"></a>{{.Title}}</h2>
{{.Description}}

{{range .HTMLIndicators}}
{{.}}
{{end}}


{{range .HTMLMetrics}}
{{.}}
{{end}}
		
{{end}}
`

const htmlIndicatorTemplate = `### <a id="{{.TitleID}}"></a>{{.Title}}
<table>
    <tr>
        <th width="25%">Description</th>
        <td>
            {{.Description}}
			<br/><br/>
			<table style="background-color: #fafafa">
			{{range .MetricPresenters}}
                <tr>
                    <td>
                        <strong>{{.Name}}</strong>
                        <p style="margin-bottom: 0.2em; margin-top: 0.2em">{{.Description}}</p>
                        <span style="font-size: small">
							<span style="display: inline-block"><strong>firehose origin</strong>: {{.Origin}}</span>
                            <span style="display: inline-block"><strong>log-cache source_id</strong>: {{.SourceID}}</span>
                            <span style="display: inline-block"><strong>type</strong>: {{.Type}}</span>
                        </span>
                    </td>
                </tr>
			{{end}}
            </table>
        </td>
    </tr>
    <tr>
        <th>PromQL</th>
        <td>
			<code>{{.PromQL}}</code>
		</td>
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
