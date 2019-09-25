package v1

import (
	"fmt"
)

// If the given document is missing data, fills it in with sane defaults. Populates the layout as
// the standard SLI/KLI/Metrics three-row setup. Defaults the title of the layout to "<name> - <version>".
// Defaults the alert to `[1m]` steps. Ensures that some values, for example chart's labels, are [] instead of nil.
func PopulateDefaults(doc *IndicatorDocument) {
	populateDefaultAlert(doc)
	populateDefaultPresentation(doc)
	populateDefaultLayout(doc)
	populateDefaultTitle(doc)
}

func populateDefaultLayout(id *IndicatorDocument) {
	if id.Spec.Layout.Sections != nil || len(id.Spec.Layout.Sections) != 0 {
		return
	}

	id.Spec.Layout = Layout{
		Sections: getLayoutSections(id.Spec.Indicators),
	}
}

func getLayoutSections(indicators []IndicatorSpec) []Section {
	sectionItems := make(map[IndicatorType][]string)

	for _, indicator := range indicators {
		sectionItems[indicator.Type] = append(sectionItems[indicator.Type], indicator.Name)
	}

	titles := []struct {
		indicatorType IndicatorType
		title         string
	}{
		{ServiceLevelIndicator, "Service Level Indicators"},
		{KeyPerformanceIndicator, "Key Performance Indicators"},
		{DefaultIndicator, "Metrics"},
	}

	sections := make([]Section, 0)
	for _, section := range titles {
		if items, found := sectionItems[section.indicatorType]; found {
			sections = append(sections, Section{
				Title:       section.title,
				Description: "",
				Indicators:  items,
			})
		}
	}
	return sections
}

func populateDefaultPresentation(doc *IndicatorDocument) {
	for i, indicator := range doc.Spec.Indicators {
		if indicator.Presentation.ChartType == "" {
			doc.Spec.Indicators[i].Presentation.ChartType = "step"
		}
		if indicator.Presentation.Labels == nil {
			doc.Spec.Indicators[i].Presentation.Labels = []string{}
		}
	}
}

func populateDefaultAlert(doc *IndicatorDocument) {
	for i, indicator := range doc.Spec.Indicators {
		if indicator.Alert.For == "" {
			doc.Spec.Indicators[i].Alert.For = "1m"
		}
		if indicator.Alert.Step == "" {
			doc.Spec.Indicators[i].Alert.Step = "1m"
		}
	}
}

func populateDefaultTitle(doc *IndicatorDocument) {
	if doc.Spec.Layout.Title == "" {
		doc.Spec.Layout.Title =
			fmt.Sprintf("%s - %s", doc.Spec.Product.Name, doc.Spec.Product.Version)
	}
}
