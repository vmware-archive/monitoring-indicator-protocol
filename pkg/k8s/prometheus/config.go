package prometheus

import (
	"log"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
)

type Config struct {
	mu                 sync.Mutex
	indicatorDocuments map[string]*v1.IndicatorDocument
}

func NewConfig() *Config {
	return &Config{
		indicatorDocuments: map[string]*v1.IndicatorDocument{},
	}
}

func (c *Config) Upsert(i *v1.IndicatorDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.indicatorDocuments[key(i)] = i
}

func (c *Config) Delete(i *v1.IndicatorDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.indicatorDocuments, key(i))
}

func key(i *v1.IndicatorDocument) string {
	return i.Namespace + "/" + i.Name
}

// String will render out the prometheus config for alert rules.
func (c *Config) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	groups := make([]prometheus_alerts.Group, 0, len(c.indicatorDocuments))
	for k, v := range c.indicatorDocuments {
		alertDocument := prometheus_alerts.AlertDocumentFrom(*v)
		alertDocument.Groups[0].Name = k
		groups = append(groups, alertDocument.Groups[0])
	}

	out, err := yaml.Marshal(prometheus_alerts.Document{Groups: groups})
	if err != nil {
		log.Print("Could not marshal alert rules")
		return "groups: []"
	}

	return string(out)
}
