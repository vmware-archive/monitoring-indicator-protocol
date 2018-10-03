package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"gopkg.in/russross/blackfriday.v2"
)

var documatationTmpl = template.Must(template.New("Metric").Parse(htmlDocumentTemplate))

func DocumentToHTML(d indicator.Document) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := documatationTmpl.Execute(buffer, documentPresenter{d.Documentation})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type documentPresenter struct {
	indicator.Documentation
}

func (dp documentPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(dp.Documentation.Description)))
}

func (dp documentPresenter) Sections() []sectionPresenter {
	var s []sectionPresenter
	for _, section := range dp.Documentation.Sections {
		s = append(s, sectionPresenter{section})
	}
	return s
}

type sectionPresenter struct {
	indicator.Section
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
