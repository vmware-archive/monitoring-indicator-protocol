package grafana_dashboard

import (
	"crypto/sha1"
	"fmt"
	"log"
	"regexp"
	"sort"

	"github.com/grafana-tools/sdk"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

var HEIGHT = 10
var WIDTH = 24

type pos struct {
	H *int `json:"h,omitempty"`
	W *int `json:"w,omitempty"`
	X *int `json:"x,omitempty"`
	Y *int `json:"y,omitempty"`
}

func ToGrafanaDashboard(document v1.IndicatorDocument, indicatorType v1.IndicatorType) (*sdk.Board, error) {
	board := sdk.NewBoard(document.Spec.Layout.Title)
	board.ID = 0
	board.Time = sdk.Time{
		From: "now-6h",
		To:   "now",
	}
	board.Timepicker = sdk.Timepicker{
		RefreshIntervals: []string{
			"5s",
			"10s",
			"30s",
			"1m",
			"5m",
			"15m",
			"30m",
			"1h",
			"2h",
			"1d",
		},
	}

	documentSectionsToPanelRows(document, board, indicatorType)
	if board.Panels == nil || len(board.Panels) == 0 {
		return nil, nil
	}

	alignPanels(board)

	return board, nil
}

func alignPanels(board *sdk.Board) {
	var y int
	for i := range board.Panels {
		board.Panels[i].ID = uint(i)
		if board.Panels[i].Type == "row" {
			board.Panels[i].GridPos = pos{H: intPointer(1), W: intPointer(24), X: intPointer(0), Y: intPointer(y)}
			y += 1
		} else {
			board.Panels[i].GridPos = pos{H: intPointer(10), W: intPointer(24), X: intPointer(0), Y: intPointer(y)}
			y += 10
		}
	}
}

func documentSectionsToPanelRows(document v1.IndicatorDocument, board *sdk.Board, indicatorType v1.IndicatorType) {
	for _, section := range document.Spec.Layout.Sections {
		if len(section.Indicators) == 0 {
			continue
		}

		var toAdd []*sdk.Panel
		for _, indicatorName := range section.Indicators {
			for _, indicatorSpec := range document.Spec.Indicators {
				if indicatorType != v1.UndefinedType && indicatorSpec.Type != indicatorType {
					continue
				}

				if indicatorSpec.Name == indicatorName {
					toAdd = append(toAdd, ToGrafanaPanel(indicatorSpec))
				}
			}
		}

		if len(toAdd) > 0 {
			appendPanelRowTitle(board, section)
			board.Panels = append(board.Panels, toAdd...)
		}
	}
}

func appendPanelRowTitle(board *sdk.Board, section v1.Section) {
	board.Panels = append(board.Panels, &sdk.Panel{
		CommonPanel: sdk.CommonPanel{
			Title:  section.Title,
			Type:   "row",
			OfType: sdk.RowType,
		},
		RowPanel: &sdk.RowPanel{
			Panels: nil,
		},
	})
}

func ToGrafanaPanel(spec v1.IndicatorSpec) *sdk.Panel {
	panel := sdk.NewGraph(spec.Name)
	panel.AddTarget(&sdk.Target{
		Expr: replaceStep(spec.PromQL),
		// TODO - this should increment per expression
		RefID: "A",
	})

	panel.GridPos = struct {
		H *int `json:"h,omitempty"`
		W *int `json:"w,omitempty"`
		X *int `json:"x,omitempty"`
		Y *int `json:"y,omitempty"`
	}{
		H: &HEIGHT,
		W: &WIDTH,
	}

	panel.GraphPanel.Description = ToGrafanaDescription(spec.Documentation)
	panel.GraphPanel.Thresholds = ToGrafanaThresholds(spec.Thresholds)
	panel.Alert = ToGrafanaAlert(spec.Thresholds)

	unit := "short"
	if spec.Presentation.Units != "" {
		unit = spec.Presentation.Units
	}

	panel.GraphPanel.Yaxes = []sdk.Axis{
		{
			Format: unit,
			Show:   true,
		},
		{
			Format: unit,
		},
	}

	panel.GraphPanel.Xaxis = sdk.Axis{
		Format: "time",
		Show:   true,
	}

	panel.GraphPanel.Lines = true
	panel.GraphPanel.Linewidth = 1

	// the following key has no particular use, but needs to be instantiated
	panel.GraphPanel.AliasColors = map[string]string{}

	return panel
}

func ToGrafanaAlert(thresholds []v1.Threshold) *sdk.Alert {
	var alert sdk.Alert

	if len(thresholds) == 0 {
		return &alert
	}

	for _, th := range thresholds {
		alert.Conditions = append(alert.Conditions, sdk.AlertCondition{
			Evaluator: struct {
				Params []float64 `json:"params,omitempty"`
				Type   string    `json:"type,omitempty"`
			}{
				Params: []float64{th.Value},
				Type:   v1.GetComparatorAbbrev(th.Operator),
			},
			Query: struct {
				Params []string `json:"params,omitempty"`
			}{Params: []string{
				"A",
				th.Alert.For,
				"now",
			}},
			Reducer: struct {
				Params []string `json:"params,omitempty"`
				Type   string   `json:"type,omitempty"`
			}{Type: "avg"},
			Type: "query",
		})
	}

	// TODO - reconsider & understand alert params
	//   Last threshold used for whole alert frequency
	//alert.Frequency = thresholds[len(thresholds)-1].Alert.Step
	//alert.For = thresholds[len(thresholds)-1].Alert.For

	alert.Frequency = "1m"
	alert.For = "5m"

	return &alert
}

func ToGrafanaDescription(docs map[string]string) *string {
	var description string

	if len(docs) == 0 {
		return nil
	}

	var keys []string

	for key := range docs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		description += fmt.Sprintf("## %s\n%s\n\n", key, docs[key])
	}
	return &description
}

func ToGrafanaThresholds(thresholds []v1.Threshold) []sdk.Threshold {
	var grafanaThresholds []sdk.Threshold
	for _, t := range thresholds {
		var comparator string
		switch {
		case t.Operator == v1.LessThanOrEqualTo || t.Operator == v1.LessThan:
			comparator = "lt"
		case t.Operator == v1.GreaterThanOrEqualTo || t.Operator == v1.GreaterThan:
			comparator = "gt"
		default:
			log.Printf("grafana dashboards only support lt/gt thresholds, threshold skipped: %v\n", t)
			continue
		}

		grafanaThresholds = append(grafanaThresholds, sdk.Threshold{
			Value:     float32(t.Value),
			ColorMode: t.Level,
			Op:        comparator,
			Fill:      true,
			Line:      true,
			Yaxis:     "left",
		})
	}

	return grafanaThresholds
}

func replaceStep(str string) string {
	reg := regexp.MustCompile(`(?i)\$step\b`)
	return reg.ReplaceAllString(str, `$$__interval`)
}

func intPointer(a int) *int {
	return &a
}

func DashboardFilename(documentBytes []byte, productName string) string {
	return fmt.Sprintf("%s_%x.json", productName, sha1.Sum(documentBytes))
}
