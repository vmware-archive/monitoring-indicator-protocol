package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"code.cloudfoundry.org/cf-indicators/pkg/indicator"
	"gopkg.in/russross/blackfriday.v2"
)

var documatationTmpl = template.Must(template.New("Metric").Parse(htmlDocumentTemplate))

func ConvertIndicatorDocument(d indicator.Document) (Documentation, error) {
	var sections []Section
	for _, s := range d.Documentation.Sections {

		var indicators []indicator.Indicator
		for _, indicatorTitle := range s.Indicators {
			found, ok := indicator.FindIndicator(indicatorTitle, d.Indicators)
			if !ok {
				return Documentation{}, fmt.Errorf("indicator %s not found in indicators section of yaml document", indicatorTitle)
			}
			indicators = append(indicators, found)
		}

		var metrics []indicator.Metric
		for _, metricTitle := range s.Metrics {
			found, ok := indicator.FindMetric(metricTitle, d.Metrics)
			if !ok {
				return Documentation{}, fmt.Errorf("metric %s not found in metrics section of yaml document", metricTitle)
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
	Section
}

func (sp sectionPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(sp.Title), " "), "-")
}

func (sp sectionPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(sp.Section.Description)))
}

func (sp sectionPresenter) Indicators() []indicatorPresenter {
	var indicatorPresenters []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		indicatorPresenters = append(indicatorPresenters, indicatorPresenter{i})
	}
	return indicatorPresenters
}

func (sp sectionPresenter) HTMLIndicators() []template.HTML {
	var renderedIndicators []template.HTML
	for _, i := range sp.Section.Indicators {
		rendered, err := IndicatorToHTML(i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not render Indicator: %s", err)
			continue
		}

		renderedIndicators = append(renderedIndicators, template.HTML(rendered))
	}
	return renderedIndicators
}

func (sp sectionPresenter) Metrics() []metricPresenter {
	var metricsPresenters []metricPresenter
	for _, m := range sp.Section.Metrics {
		metricsPresenters = append(metricsPresenters, metricPresenter{m})
	}
	return metricsPresenters
}

func (sp sectionPresenter) HTMLMetrics() []template.HTML {
	var renderedMetrics []template.HTML
	for _, m := range sp.Section.Metrics {
		rendered, err := MetricToHTML(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not render Metric: %s", err)
			continue
		}

		renderedMetrics = append(renderedMetrics, template.HTML(rendered))
	}
	return renderedMetrics
}
