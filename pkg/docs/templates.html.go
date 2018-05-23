package docs

const htmlIndicatorTemplate = `<h3 id="{{.TitleID}}"><a id="{{.AnchorID}}"></a>{{.Name}}</h3>
<table>
    <tr><th colspan="2" style="text-align: center;"><br/> {{range .Metrics}}{{.}}<br/>{{end}}<br/><br/></th></tr>
    <tr>
        <th width="25%%">Description</th>
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

const htmlMetricTemplate = `<h3 id="{{.TitleID}}"><a id="{{.AnchorID}}"></a>{{.Title}}</h3>
<table>
   <tbody><tr><th colspan="2" style="text-align: center;"><br> {{.Name}}<br><br></th></tr>
   <tr>
      <th width="25%">Description</th>
      <td>{{.Description}}</td>
   </tr>
</tbody></table>`
