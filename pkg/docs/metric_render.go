package docs

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"

	"bytes"
	"html/template"
	"strings"
	"gopkg.in/russross/blackfriday.v2"
)

var metricTmpl = template.Must(template.New("Metric").Parse(htmlMetricTemplate))

func MetricToHTML(m indicator.Metric) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := metricTmpl.Execute(buffer, metricPresenter{m})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type metricPresenter struct {
	indicator.Metric
}

func (m metricPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(m.Title), " "), "-")
}

func (m metricPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(m.Metric.Description)))
}
