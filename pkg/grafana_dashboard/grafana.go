package grafana_dashboard

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"
	"encoding/json"
	"github.com/grafana-tools/sdk"
)

func DocumentToDashboard(document indicator.Document) (string, error) {
	board := sdk.NewBoard(document.Documentation.Title)
	setBoardDefaults(board)

	for _, i := range document.Indicators {
		row := board.AddRow(i.Name)

		graph := generateIndicatorGraph(i)

		row.Add(graph)
	}

	text, err := json.Marshal(board)
	return string(text), err
}

func setBoardDefaults(board *sdk.Board) {
	board.Time = sdk.Time{
		From: "now-1h",
		To:   "now",
	}
	now := new(bool)
	*now = true
	board.Timepicker = sdk.Timepicker{
		Now: now,
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
		TimeOptions: []string{
			"5m",
			"15m",
			"1h",
			"6h",
			"12h",
			"24h",
			"2d",
			"7d",
			"30d",
		},
	}
}

func generateIndicatorGraph(i indicator.Indicator) *sdk.Panel {
	graph := sdk.NewGraph(i.Name)
	graph.AddTarget(&sdk.Target{
		Expr:           i.PromQL,
		Format:         "time_series",
		IntervalFactor: 1,
	})
	graph.Lines = true
	graph.Linewidth = 2
	graph.Yaxes = []sdk.Axis{{
		Format:  "short",
		LogBase: 1,
		Show:    true,
	}, {
		Format:  "short",
		LogBase: 1,
		Show:    true,
	}}
	graph.Legend.Show = true
	graph.Tooltip = sdk.Tooltip{
		ValueType: "individual",
		Shared:    true,
		Sort:      0,
	}
	graph.AliasColors = struct{}{}
	return graph
}
