package docs

import (
	"bytes"
	"errors"
	"html/template"
	"strings"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"

	"gopkg.in/russross/blackfriday.v2"
)

func docToTemplate(d v1alpha1.IndicatorDocument, t *template.Template) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, documentPresenter{d.Spec.Layout, d.Spec.Indicators})

	if err != nil {
		return "", errors.New("failed to convert document to template")
	}

	return buffer.String(), nil
}

type documentPresenter struct {
	v1alpha1.Layout
	indicators []v1alpha1.IndicatorSpec
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
	v1alpha1.Section
	indicators []v1alpha1.IndicatorSpec
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
		indie := findIndicator(i, sp.indicators)

		indicatorPresenters = append(indicatorPresenters, indicatorPresenter{*indie})
	}
	return indicatorPresenters, nil
}

func findIndicator(name string, indicators []v1alpha1.IndicatorSpec) *v1alpha1.IndicatorSpec {
	for _, i := range indicators {
		if i.Name == name {
			return &i
		}
	}

	return nil
}

func (sp sectionPresenter) HTMLIndicators() ([]indicatorPresenter, error) {
	var renderedIndicators []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		indie := findIndicator(i, sp.indicators)
		renderedIndicators = append(renderedIndicators, indicatorPresenter{*indie})
	}
	return renderedIndicators, nil
}
