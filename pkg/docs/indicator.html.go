package docs

import "html/template"

const htmlTemplate = template.HTML(`<h3 id="{{.TitleID}}"><a id="{{.AnchorID}}"></a>{{.Name}}</h3>
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
            {{range .Thresholds}} <em>{{.Level}}</em>: {{.Operator}} {{.Value}}<br/> {{end}}
        </td>
    </tr>
    <tr>
        <th>Recommended response</th>
        <td>
            {{.Response}}
        </td>
    </tr>
</table>`)
