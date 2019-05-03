package docs

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/indicator"
	"gopkg.in/russross/blackfriday.v2"
)

func docToTemplate(d indicator.Document, t *template.Template) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, documentPresenter{d.Layout, d.Indicators})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type documentPresenter struct {
	indicator.Layout
	indicators []indicator.Indicator
}

func (dp documentPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(dp.Layout.Description)))
}

func (dp documentPresenter) Sections() []sectionPresenter {
	var s []sectionPresenter
	for _, section := range dp.Layout.Sections {
		s = append(s, sectionPresenter{section, dp.indicators})
	}
	return s
}

type sectionPresenter struct {
	indicator.Section
	indicators []indicator.Indicator
}

func (sp sectionPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(sp.Title), " "), "-")
}

func (sp sectionPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(sp.Section.Description)))
}

func (sp sectionPresenter) Indicators() ([]indicatorPresenter, error) {
	var indicatorPresenters []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		indie, err := findIndicator(i, sp.indicators)
		if err != nil {
			return nil, err
		}
		indicatorPresenters = append(indicatorPresenters, indicatorPresenter{*indie})
	}
	return indicatorPresenters, nil
}

func findIndicator(name string, indicators []indicator.Indicator) (*indicator.Indicator, error) {
	for _, i := range indicators {
		if i.Name == name {
			return &i, nil
		}
	}

	return nil, nil
}

func (sp sectionPresenter) HTMLIndicators() ([]indicatorPresenter, error) {
	var renderedIndicators []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		indie, err := findIndicator(i, sp.indicators)
		if err != nil {
			return nil, err
		}
		renderedIndicators = append(renderedIndicators, indicatorPresenter{*indie})
	}
	return renderedIndicators, nil
}
