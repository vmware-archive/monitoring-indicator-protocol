package docs

import (
	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"

	"bytes"
	"html/template"
	"strings"
	"gopkg.in/russross/blackfriday.v2"
)

var metricTmpl = template.Must(template.New("metric").Parse(htmlMetricTemplate))

func MetricToHTML(m indicator.Metric) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := metricTmpl.Execute(buffer, metricPresenter{m})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type metricPresenter struct {
	metric indicator.Metric
}

func (m metricPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(m.metric.Title), " "), "-")
}

func (m metricPresenter) AnchorID() string {
	return strings.Join(strings.Split(m.metric.Title, " "), "")
}

func (m metricPresenter) Title() string {
	return m.metric.Title
}

func (m metricPresenter) Name() string {
	return m.metric.Name
}

func (m metricPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(m.metric.Description)))
}
