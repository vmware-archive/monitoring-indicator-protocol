package registry

import (
	"code.cloudfoundry.org/indicators/pkg/indicator"
	"time"
)

type Document struct {
	Labels                map[string]string
	Indicators            []indicator.Indicator
	registrationTimestamp time.Time
}

type APIV0Metric struct {
	Title       string `json:"title"`
	Origin      string `json:"origin"`
	SourceID    string `json:"source_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
}

type APIV0Threshold struct {
	Level    string  `json:"level"`
	Dynamic  bool    `json:"dynamic"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type APIV0Indicator struct {
	Name        string           `json:"name"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	PromQL      string           `json:"promql"`
	Thresholds  []APIV0Threshold `json:"thresholds"`
	Metrics     []APIV0Metric    `json:"metrics"`
	Response    string           `json:"response"`
	Measurement string           `json:"measurement"`
}

type APIV0Document struct {
	Labels     map[string]string `json:"labels"`
	Indicators []APIV0Indicator  `json:"indicators"`
}

func (doc Document) ToAPIV0() APIV0Document {
	indicators := make([]APIV0Indicator, 0)
	for _, i := range doc.Indicators {
		thresholds := make([]APIV0Threshold, 0)
		for _, t := range i.Thresholds {
			thresholds = append(thresholds, APIV0Threshold{
				Level:    t.Level,
				Dynamic:  t.Dynamic,
				Operator: t.Operator.String(),
				Value:    t.Value,
			})
		}

		metrics := make([]APIV0Metric, 0)
		for _, m := range i.Metrics {
			metrics = append(metrics, APIV0Metric{
				Title:       m.Title,
				Origin:      m.Origin,
				SourceID:    m.SourceID,
				Name:        m.Name,
				Type:        m.Type,
				Description: m.Description,
				Frequency:   m.Frequency,
			})
		}

		indicators = append(indicators, APIV0Indicator{
			Name:        i.Name,
			Title:       i.Title,
			Description: i.Description,
			PromQL:      i.PromQL,
			Thresholds:  thresholds,
			Metrics:     metrics,
			Response:    i.Response,
			Measurement: i.Measurement,
		})
	}

	return APIV0Document{
		Labels:     doc.Labels,
		Indicators: indicators,
	}
}
