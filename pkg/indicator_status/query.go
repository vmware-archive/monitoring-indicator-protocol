package indicator_status

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

// QueryValues takes a promql string and returns the latest values.
func QueryValues(client PromQLClient, promql string) ([]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value, err := client.Query(ctx, promql, time.Time{})
	if err != nil {
		return nil, err
	}

	vector, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("could not assert result from prometheus as Vector: %T", value)
	}

	values := make([]float64, 0, len(vector))
	for _, v := range vector {
		values = append(values, float64(v.Value))
	}
	return values, nil
}
