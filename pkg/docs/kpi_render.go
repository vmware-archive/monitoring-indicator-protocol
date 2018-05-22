package docs

import (
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"

	"html/template"
	"bytes"
	"strings"
	"gopkg.in/russross/blackfriday.v2"
	"fmt"
)

var tmpl = template.Must(template.New("kpi").Parse(string(htmlTemplate)))

func HTML(kpi kpi.KPI) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := tmpl.Execute(buffer, kpiPresenter{kpi})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type kpiPresenter struct {
	kpi kpi.KPI
}

func (k kpiPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(k.kpi.Name), " "), "-")
}

func (k kpiPresenter) AnchorID() string {
	return strings.Join(strings.Split(k.kpi.Name, " "), "")
}

func (k kpiPresenter) Name() string {
	return k.kpi.Name
}

func (k kpiPresenter) Metrics() []string {
	return k.kpi.Metrics
}

func (k kpiPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(k.kpi.Description)))
}

func (k kpiPresenter) PromQL() template.HTML {
	return template.HTML(k.kpi.PromQL)
}

func (k kpiPresenter) Measurement() template.HTML {
	return template.HTML(blackfriday.Run([]byte(k.kpi.Measurement)))
}

func (k kpiPresenter) Response() template.HTML {
	return template.HTML(blackfriday.Run([]byte(k.kpi.Response)))
}

type thresholdPresenter struct {
	kpiThreshold kpi.Threshold
}

func (k kpiPresenter) Thresholds() []thresholdPresenter {
	var thresh []thresholdPresenter
	for _, t := range k.kpi.Thresholds {
		thresh = append(thresh, thresholdPresenter{t})
	}
	return thresh
}

func (t thresholdPresenter) Level() string {
	switch t.kpiThreshold.Level {
	case "warning":
		return "Yellow warning"
	case "critical":
		return "Red critical"
	default:
		return ""
	}
}

func (t thresholdPresenter) Operator() string {
	switch t.kpiThreshold.Operator {
	case kpi.LessThan:
		return "<"
	case kpi.LessThanOrEqualTo:
		return "<="
	case kpi.EqualTo:
		return "=="
	case kpi.NotEqualTo:
		return "!="
	case kpi.GreaterThanOrEqualTo:
		return ">="
	case kpi.GreaterThan:
		return ">"
	default:
		return ""
	}
}

func (t thresholdPresenter) Value() string {
	return fmt.Sprintf("%v", t.kpiThreshold.Value)
}
