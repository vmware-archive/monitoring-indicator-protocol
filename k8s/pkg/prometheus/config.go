package prometheus

import (
	"log"
	"sync"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"gopkg.in/yaml.v2"
)

type Config struct {
	mu                 sync.Mutex
	indicatorDocuments map[string]*v1alpha1.IndicatorDocument
}

func NewConfig() *Config {
	return &Config{
		indicatorDocuments: map[string]*v1alpha1.IndicatorDocument{},
	}
}

func (c *Config) Upsert(i *v1alpha1.IndicatorDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.indicatorDocuments[key(i)] = i
}

func (c *Config) Delete(i *v1alpha1.IndicatorDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.indicatorDocuments, key(i))
}

func key(i *v1alpha1.IndicatorDocument) string {
	return i.Namespace + "/" + i.Name
}

// String will render out the prometheus config for alert rules.
func (c *Config) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	groups := make([]prometheus_alerts.Group, 0, len(c.indicatorDocuments))
	for k, v := range c.indicatorDocuments {
		doc := domain.Map(v)
		alertDocument := prometheus_alerts.AlertDocumentFrom(doc)
		alertDocument.Groups[0].Name = k
		groups = append(groups, alertDocument.Groups[0])
	}

	out, err := yaml.Marshal(prometheus_alerts.Document{Groups: groups})
	if err != nil {
		log.Printf("Could not marshal alert rules: %s", err)
		return "groups: []"
	}

	return string(out)
}
