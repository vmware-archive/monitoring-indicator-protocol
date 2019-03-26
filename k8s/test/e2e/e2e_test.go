package e2e_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os/user"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	clientsetV1alpha1 "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/domain"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
)

type k8sClients struct {
	k8sClientset *kubernetes.Clientset
	idClient     *clientsetV1alpha1.AppsV1alpha1Client
}

var clients k8sClients

func init() {
	rand.Seed(time.Now().UnixNano())
	config, err := clientcmd.BuildConfigFromFlags("", expandHome("~/.kube/config"))
	if err != nil {
		log.Panic(err.Error())
	}

	clients.k8sClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	clients.idClient, err = clientsetV1alpha1.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}
}

func TestConfigMaps(t *testing.T) {
	testCases := map[string]func(*v1alpha1.IndicatorDocument) func() bool{
		"grafana": func(id *v1alpha1.IndicatorDocument) func() bool {
			return func() bool {
				cm, err := clients.k8sClientset.CoreV1().
					ConfigMaps("grafana").
					Get(grafanaDashboardFilename(id), metav1.GetOptions{})
				if err != nil {
					t.Logf("Unable to get config map: %s", err)
					return false
				}
				return grafanaConfigMapMatch(t, grafanaDashboardFilename(id)+".json", cm, id)
			}
		},
		"prometheus": func(id *v1alpha1.IndicatorDocument) func() bool {
			return func() bool {
				cm, err := clients.k8sClientset.CoreV1().
					ConfigMaps("prometheus").
					Get("prometheus-server", metav1.GetOptions{})
				if err != nil {
					t.Logf("Unable to get config map: %s", err)
					return false
				}
				return prometheusConfigMapMatch(t, cm, id)
			}
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			g := NewGomegaWithT(t)
			ns, cleanup := createNamespace(t)
			defer cleanup()
			id := indicatorDocument(ns)

			_, err := clients.idClient.IndicatorDocuments(ns).Create(id)

			g.Expect(err).ToNot(HaveOccurred())
			g.Eventually(tc(id), 5).Should(BeTrue())
		})
	}
}

func grafanaConfigMapMatch(t *testing.T, dashboardFilename string, cm *v1.ConfigMap, id *v1alpha1.IndicatorDocument) bool {
	dashboard := grafana_dashboard.DocumentToDashboard(domain.Map(id))
	data, err := json.Marshal(dashboard)
	if err != nil {
		t.Logf("Unable to marshal: %s", err)
		return false
	}

	match, err := MatchJSON(data).Match(cm.Data[dashboardFilename])
	if err != nil {
		t.Logf("Unable to match: %s", err)
		return false
	}
	return match
}

func prometheusConfigMapMatch(t *testing.T, cm *v1.ConfigMap, id *v1alpha1.IndicatorDocument) bool {
	t.Log("Converting indicator document to prometheus alerts yaml")
	alerts := prometheus_alerts.AlertDocumentFrom(domain.Map(id))
	alerts.Groups[0].Name = id.Namespace + "/" + id.Name
	expected, err := yaml.Marshal(alerts)
	if err != nil {
		t.Logf("Unable to marshal: %s", err)
		return false
	}

	t.Log("Unmarshaling config map prometheus alerts yaml")
	var (
		cmAlerts map[string][]map[string]interface{}
		cmAlert  interface{}
	)
	err = yaml.Unmarshal([]byte(cm.Data["alerts"]), &cmAlerts)
	if err != nil {
		t.Logf("Unable to unmarshal: %s", err)
		return false
	}

	t.Log("Selecting the alerting rules we are concerned about")
	for _, group := range cmAlerts["groups"] {
		if group["name"] == alerts.Groups[0].Name {
			cmAlert = group
		}
	}
	if cmAlert == nil {
		t.Log("Unable to find alert group")
		return false
	}

	t.Log("Remarshaling the specific alert for asserting")
	newCmAlerts := map[string][]interface{}{
		"groups": {cmAlert},
	}
	actual, err := yaml.Marshal(newCmAlerts)
	if err != nil {
		t.Logf("Unable to marshal: %s", err)
		return false
	}

	match, err := MatchYAML(expected).Match(actual)
	if err != nil {
		t.Logf("Unable to match: %s", err)
		return false
	}
	if !match {
		t.Logf(cmp.Diff(expected, actual))
		return false
	}
	return true
}

// TODO: generate random values for these:
func indicatorDocument(ns string) *v1alpha1.IndicatorDocument {
	var threshold float64 = 500
	return &v1alpha1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("e2e-test-%d", rand.Intn(math.MaxInt32)),
			Namespace: ns,
		},
		Spec: v1alpha1.IndicatorDocumentSpec{
			Product: v1alpha1.Product{
				Name:    "e2e-test-product",
				Version: "v1.2.3-rc1",
			},
			Indicators: []v1alpha1.Indicator{
				{
					Name:   "e2d-test-indicator",
					Promql: "rate(some_metric[10m])",
					Alert: v1alpha1.Alert{
						For:  "5m",
						Step: "2m",
					},
					Thresholds: []v1alpha1.Threshold{
						{
							Level: "critical",
							Gte:   &threshold,
						},
					},
				},
			},
		},
	}
}

func expandHome(s string) string {
	usr, err := user.Current()
	if err != nil {
		log.Panicf("unable to expand user: %s", err)
	}
	return strings.Replace(s, "~", usr.HomeDir, -1)
}

func createNamespace(t *testing.T) (string, func()) {
	nsName := fmt.Sprintf("e2e-test-%d", rand.Intn(math.MaxInt32))
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	nsr, err := clients.k8sClientset.CoreV1().Namespaces().Create(ns)
	if err != nil {
		t.Errorf("unable to create namespace: %s", err)
	}
	return nsName, func() {
		err := clients.k8sClientset.CoreV1().Namespaces().Delete(nsName, &metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{
				UID: &nsr.UID,
			},
		})
		if err != nil {
			t.Errorf("unable to delete namespace: %s", err)
		}
	}
}

func grafanaDashboardFilename(id *v1alpha1.IndicatorDocument) string {
	return fmt.Sprintf("indicator-protocol-grafana-dashboard.%s.%s", id.Namespace, id.Name)
}
