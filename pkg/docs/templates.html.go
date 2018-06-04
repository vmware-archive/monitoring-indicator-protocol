package docs

const htmlDocumentTemplate = `
<h1 class="title-container">{{.Title}}</h1>

<div id="js-quick-links">
    <div class="quick-links">
        <ul>
        {{range .Sections}}
            <li>
                <a href="#{{.TitleID}}">{{.Title}}</a>
                <ul>
                    {{range .Indicators}}
                    <li><a href="#{{.TitleID}}">{{.Title}}</a></li>
                    {{end}}

                    {{range .Metrics}}
                    <li><a href="#{{.TitleID}}">{{.Title}}</a></li>
                    {{end}}
                </ul>
            </li>
        {{end}}
        </ul>
    </div>
</div>

{{.Description}}

{{range .Sections}}
<div>
	<h2 id="{{.TitleID}}">{{.Title}}</h2>
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

const htmlIndicatorTemplate = `<h3 id="{{.TitleID}}">{{.Title}}</h3>
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

const htmlMetricTemplate = `<h3 id="{{.TitleID}}">{{.Title}}</h3>
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
