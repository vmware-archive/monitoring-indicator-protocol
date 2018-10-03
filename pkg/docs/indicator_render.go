package docs

import (
  "code.cloudfoundry.org/indicators/pkg/indicator"

  "bytes"
  "fmt"
  "gopkg.in/russross/blackfriday.v2"
  "html/template"
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

func (p indicatorPresenter) PromQL() template.HTML {
  return template.HTML(p.Indicator.PromQL)
}

func (p indicatorPresenter) Title() string {
  t, found := p.Documentation["title"]
  if !found {
    return p.Name
  }

  return t
}

func (p indicatorPresenter) Description() template.HTML {
  return p.markdownDocumentationField("description")
}

func (p indicatorPresenter) ThresholdNote() template.HTML {
  return p.markdownDocumentationField("threshold_note")
}

func (p indicatorPresenter) OtherDocumentationFields() map[string]template.HTML {
  fields := make(map[string]template.HTML, 0)

  for k, v := range p.Documentation {
    if isUnusedDocumentationField(k) {
      fields[k] = template.HTML(blackfriday.Run([]byte(v)))
    }
  }

  return fields
}

func isUnusedDocumentationField(fieldName string) bool {
  return fieldName != "title" && fieldName != "description" && fieldName != "threshold_note"
}

func (p indicatorPresenter) markdownDocumentationField(field string) template.HTML {
  d, found := p.Documentation[field]
  if !found {
    return ""
  }

  return template.HTML(blackfriday.Run([]byte(d)))
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
