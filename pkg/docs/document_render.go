package docs

import (
	"bytes"
	"html/template"
	"gopkg.in/russross/blackfriday.v2"
	"fmt"
	"os"
	"strings"
	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"
)

var documatationTmpl = template.Must(template.New("Metric").Parse(htmlDocumentTemplate))

func ConvertIndicatorDocument(d indicator.Document) (Documentation, error) {
	var sections []Section
	for _, s := range d.Documentation.Sections {

		var indicators []indicator.Indicator
		for _, i := range s.Indicators {
			found, err := lookupIndicator(i, d.Indicators)
			if err != nil {
				return Documentation{}, err
			}
			indicators = append(indicators, found)
		}

		var metrics []indicator.Metric
		for _, m := range s.Metrics {
			found, err := lookupMetric(m, d.Metrics)
			if err != nil {
				return Documentation{}, err
			}
			metrics = append(metrics, found)
		}

		sections = append(sections, Section{
			Title:       s.Title,
			Description: s.Description,
			Indicators:  indicators,
			Metrics:     metrics,
		})
	}

	return Documentation{
		Title:       d.Documentation.Title,
		Description: d.Documentation.Description,
		Sections:    sections,
	}, nil
}

func lookupIndicator(indicatorName string, indicators []indicator.Indicator) (indicator.Indicator, error) {
	for _, i := range indicators {
		if i.Name == indicatorName {
			return i, nil
		}
	}

	return indicator.Indicator{}, fmt.Errorf("indicator %s not found in indicators section of yaml document", indicatorName)
}

func lookupMetric(metricName string, metrics []indicator.Metric) (indicator.Metric, error) {
	for _, m := range metrics {
		if m.Name == metricName {
			return m, nil
		}
	}

	return indicator.Metric{}, fmt.Errorf("metric %s not found in metrics section of yaml document", metricName)
}

func DocumentToHTML(d Documentation) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := documatationTmpl.Execute(buffer, documentPresenter{d})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type documentPresenter struct {
	documentation Documentation
}

func (dp documentPresenter) Title() string {
	return dp.documentation.Title
}

func (dp documentPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(dp.documentation.Description)))
}

func (dp documentPresenter) Sections() []sectionPresenter {
	var s []sectionPresenter
	for _, section := range dp.documentation.Sections {
		s = append(s, sectionPresenter{section})
	}
	return s
}

type sectionPresenter struct {
	section Section
}

func (sp sectionPresenter) Title() string {
	return sp.section.Title
}

func (sp sectionPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(sp.section.Title), " "), "-")
}

func (sp sectionPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(sp.section.Description)))
}

func (sp sectionPresenter) Indicators() []template.HTML {
	var renderedIndicators []template.HTML
	for _, i := range sp.section.Indicators {
		rendered, err := IndicatorToHTML(i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not render Indicator: %s", err)
			continue
		}

		renderedIndicators = append(renderedIndicators, template.HTML(rendered))
	}
	return renderedIndicators
}

func (sp sectionPresenter) Metrics() []template.HTML {
	var renderedMetrics []template.HTML
	for _, m := range sp.section.Metrics {
		rendered, err := MetricToHTML(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not render Metric: %s", err)
			continue
		}

		renderedMetrics = append(renderedMetrics, template.HTML(rendered))
	}
	return renderedMetrics
}
