package docs

import (
	"code.cloudfoundry.org/cf-indicators/pkg/indicator"

	"html/template"
	"bytes"
	"strings"
	"gopkg.in/russross/blackfriday.v2"
	"fmt"
)

var indicatorTmpl = template.Must(template.New("Indicator").Parse(htmlIndicatorTemplate))

func IndicatorToHTML(i indicator.Indicator) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := indicatorTmpl.Execute(buffer, indicatorPresenter{i})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type indicatorPresenter struct {
	indicator.Indicator
}

func (p indicatorPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(p.Title), " "), "-")
}

func (p indicatorPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Indicator.Description)))
}

func (p indicatorPresenter) PromQL() template.HTML {
	return template.HTML(p.Indicator.PromQL)
}

func (p indicatorPresenter) Measurement() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Indicator.Measurement)))
}

func (p indicatorPresenter) Response() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Indicator.Response)))
}

func (p indicatorPresenter) MetricPresenters() []metricPresenter {
	metricPresenters := make([]metricPresenter, 0)

	for _, metric := range p.Metrics {
		metricPresenters = append(metricPresenters, metricPresenter{metric})
	}

	return metricPresenters
}

type thresholdPresenter struct {
	threshold indicator.Threshold
}

func (p indicatorPresenter) Thresholds() []thresholdPresenter {
	var tp []thresholdPresenter
	for _, t := range p.Indicator.Thresholds {
		tp = append(tp, thresholdPresenter{t})
	}
	return tp
}

func (t thresholdPresenter) Level() string {
	switch t.threshold.Level {
	case "warning":
		return "Yellow warning"
	case "critical":
		return "Red critical"
	default:
		return t.threshold.Level
	}
}

func (t thresholdPresenter) OperatorAndValue() string {
	if t.threshold.Dynamic {
		return "Dynamic"
	}

	return fmt.Sprintf("%s %s", t.operator(), t.value())
}

func (t thresholdPresenter) operator() string {
	switch t.threshold.Operator {
	case indicator.LessThan:
		return "<"
	case indicator.LessThanOrEqualTo:
		return "<="
	case indicator.EqualTo:
		return "=="
	case indicator.NotEqualTo:
		return "!="
	case indicator.GreaterThanOrEqualTo:
		return ">="
	case indicator.GreaterThan:
		return ">"
	default:
		return ""
	}
}

func (t thresholdPresenter) value() string {
	return fmt.Sprintf("%v", t.threshold.Value)
}
