package grafana_dashboard

type GrafanaDashboard struct {
	Title string       `json:"title"`
	Rows  []GrafanaRow `json:"rows"`
}

type GrafanaRow struct {
	Title  string         `json:"title"`
	Panels []GrafanaPanel `json:"panels"`
}

type GrafanaPanel struct {
	Title      string             `json:"title"`
	Type       string             `json:"type"`
	Targets    []GrafanaTarget    `json:"targets"`
	Thresholds []GrafanaThreshold `json:"thresholds"`
}

type GrafanaTarget struct {
	Expression string `json:"expr"`
}

type GrafanaThreshold struct {
	Value     float64 `json:"value"`
	ColorMode string  `json:"colorMode"`
	Op        string  `json:"op"`
	Fill      bool    `json:"fill"`
	Line      bool    `json:"line"`
	Yaxis     string  `json:"yaxis"`
}
