package e2e_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os/user"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/api_versions"
	"gopkg.in/yaml.v2"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	clientSetV1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/clientset/versioned/typed/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_alerts"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/grafana_dashboard"
)

type k8sClients struct {
	k8sClientset *kubernetes.Clientset
	idClient     *clientSetV1.AppsV1Client
}

var (
	clients                                                     k8sClients
	httpClient                                                  *http.Client
	grafanaURI, grafanaAdminUser, grafanaAdminPw, prometheusURI *string
)

func init() {
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	rand.Seed(time.Now().UnixNano())
	grafanaURI = flag.String("grafana-uri", "", "")
	grafanaAdminUser = flag.String("grafana-admin-user", "", "")
	grafanaAdminPw = flag.String("grafana-admin-pw", "", "")
	prometheusURI = flag.String("prometheus-uri", "", "")
	flag.Parse()
	if *grafanaURI == "" {
		log.Panic("Oh no! Grafana URI not provided")
	}
	if *grafanaAdminUser == "" {
		log.Panic("Oh no! Grafana user not provided")
	}
	if *grafanaAdminPw == "" {
		log.Panic("Oh no! Grafana password not provided")
	}
	if *prometheusURI == "" {
		log.Panic("Oh no! Prometheus URI not provided")
	}
	config, err := clientcmd.BuildConfigFromFlags("", expandHome("~/.kube/config"))
	if err != nil {
		log.Panic(err.Error())
	}

	clients.k8sClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	clients.idClient, err = clientSetV1.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}
}

func TestControllers(t *testing.T) {
	const testTimeout = 120 * time.Second

	setup := func (t *testing.T) (func(), *v1.IndicatorDocument) {
		t.Parallel()

		ns, cleanup := createNamespace(t)
		indiDoc := indicatorDocument(ns)
		t.Logf("Creating indicator document in namespace: %s", ns)
		_, err := clients.idClient.IndicatorDocuments(ns).Create(indiDoc)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		return cleanup, indiDoc
	}

	t.Run("grafana", func(t *testing.T) {
		g := NewGomegaWithT(t)
		cleanup, indiDoc := setup(t)
		defer cleanup()

		grafanaMatches := func() bool {
			cm, err := clients.k8sClientset.CoreV1().
				ConfigMaps("grafana").
				Get(grafanaDashboardFilename(indiDoc), metav1.GetOptions{})

			if err != nil {
				t.Logf("Unable to get config map, retrying: %s", err)
				return false
			}
			match := grafanaConfigMapMatch(t, grafanaDashboardFilename(indiDoc)+".json", cm, indiDoc)
			if !match {
				t.Logf("Unable to match grafana config")
				return false
			}

			return grafanaApiResponseMatch(t, indiDoc)
		}

		g.Eventually(grafanaMatches, testTimeout).Should(BeTrue())
	})

	t.Run("prometheus", func(t *testing.T) {
		g := NewGomegaWithT(t)
		cleanup, indiDoc := setup(t)
		defer cleanup()

		prometheusMatches := func() bool {
			cm, err := clients.k8sClientset.CoreV1().
				ConfigMaps("prometheus").
				Get("prometheus-server", metav1.GetOptions{})
			if err != nil {
				t.Logf("Unable to get config map, retrying: %s", err)
				return false
			}
			match := prometheusConfigMapMatch(t, cm, indiDoc)
			if !match {
				t.Logf("Unable to match prometheus config")
				return false
			}
			return prometheusApiResponseMatch(t, indiDoc)
		}
		g.Eventually(prometheusMatches, testTimeout).Should(BeTrue())
	})

	t.Run("lifecycle", func(t *testing.T) {
		g := NewGomegaWithT(t)
		cleanup, indiDoc := setup(t)
		defer cleanup()

		lifecycleExisting := func() bool {
			resources := clients.idClient.Indicators(indiDoc.Namespace)
			resource, err := getIndicator(resources, indiDoc.Name, indiDoc.Spec.Indicators[0].Name)
			if err != nil {
				t.Logf("Unable to get new indicator, retrying: %s", err)
				return false
			}
			return resource != nil
		}

		g.Eventually(lifecycleExisting, testTimeout).Should(BeTrue())
	})

}

func TestAdmission(t *testing.T) {
	t.Run("patches default values", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns, cleanup := createNamespace(t)
		defer cleanup()
		id := indicatorDocument(ns)
		t.Logf("Creating indicator document in namespace: %s", ns)
		_, err := clients.idClient.IndicatorDocuments(ns).Create(id)

		g.Expect(err).ToNot(HaveOccurred())

		defaultPresentation := &v1.Presentation{
			ChartType:    "step",
			CurrentValue: false,
			Frequency:    0,
			Labels:       []string{},
		}
		g.Eventually(getIndicatorPresentation(t, ns, id.Name, id.Spec.Indicators[0].Name), 10).
			Should(Equal(defaultPresentation))
	})
	t.Run("rejects invalid indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns, cleanup := createNamespace(t)
		defer cleanup()
		id := indicatorDocument(ns)
		id.Spec.Indicators[0].Thresholds[0].Operator = v1.ThresholdOperator(0)
		t.Logf("Creating indicator document in namespace: %s", ns)
		_, err := clients.idClient.IndicatorDocuments(ns).Create(id)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("indicators[0].thresholds[0] operator [lt, lte, eq, neq, gte, gt] is required"))
	})
}

func TestStatus(t *testing.T) {
	t.Run("it updates indicator status", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns, cleanup := createNamespace(t)
		defer cleanup()

		timestamp := time.Now().UnixNano()
		indicatorName := fmt.Sprintf("my-name-%d", timestamp)
		indicator := &v1.Indicator{
			ObjectMeta: metav1.ObjectMeta{
				Name:      indicatorName,
				Namespace: ns,
			},
			Spec: v1.IndicatorSpec{
				Name:   "my_cool_name",
				PromQL: `prometheus_http_request_duration_seconds_sum{handler="/-/healthy"}`,
				Alert: v1.Alert{
					For:  "5m",
					Step: "2m",
				},
				Thresholds: []v1.Threshold{
					{
						Level:    "critical",
						Operator: v1.GreaterThan,
						Value:    float64(10),
					},
				},
			},
		}
		t.Logf("Creating indicator in namespace: %s", ns)
		_, err := clients.idClient.Indicators(ns).Create(indicator)
		g.Expect(err).To(Not(HaveOccurred()))

		resources := clients.idClient.Indicators(ns)

		g.Eventually(func() string {
			indicator, err = getSoloIndicator(resources, indicatorName)
			if err != nil {
				t.Logf("error: %s", err)
			}
			return indicator.Status.Phase
		}, 35*time.Second).Should(Equal("HEALTHY"))
	})
}

func getIndicatorPresentation(t *testing.T, ns string, indicatorDocName string, indicatorName string) func() *v1.Presentation {
	return func() *v1.Presentation {
		resources := clients.idClient.Indicators(ns)
		resource, err := getIndicator(resources, indicatorDocName, indicatorName)
		if err != nil {
			t.Logf("Unable to get new indicator, retrying: %s", err)
			return nil
		}
		return &resource.Spec.Presentation
	}
}

func getIndicator(resources clientSetV1.IndicatorInterface, indicatorDocName string, indicatorName string) (*v1.Indicator, error) {
	return resources.Get(fmt.Sprintf("%s-%s", indicatorDocName, strings.Replace(indicatorName, "_", "-", -1)), metav1.GetOptions{})
}

func getSoloIndicator(resources clientSetV1.IndicatorInterface, indicatorName string) (*v1.Indicator, error) {
	return resources.Get(indicatorName, metav1.GetOptions{})
}

func grafanaApiResponseMatch(t *testing.T, document *v1.IndicatorDocument) bool {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/search?query=%s", *grafanaURI, document.Spec.Product.Name), nil)
	if err != nil {
		t.Logf("Unable to create request to get Grafana config through API, retrying: %s", err)
		return false
	}
	request.SetBasicAuth(*grafanaAdminUser, *grafanaAdminPw)
	response, err := httpClient.Do(request)
	if err != nil {
		t.Logf("Unable to retrieve config through Grafana API, retrying: %s", err)
		return false
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Logf("Unable to read Grafana config response body, retrying: %s", err)
		return false
	}
	var results []grafanaSearchResult
	err = json.Unmarshal(body, &results)
	if err != nil {
		t.Logf("Unable to unmarshal Grafana config response body, retrying: %s", err)
		return false
	}

	if len(results) == 0 {
		t.Log("Grafana API returned 0 results")
		return false
	}
	return len(results) == 1
}

type grafanaSearchResult struct {
	Title string `json:"title"`
}

func prometheusApiResponseMatch(t *testing.T, document *v1.IndicatorDocument) bool {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/v1/rules", *prometheusURI), nil)
	if err != nil {
		t.Logf("Unable to create request to get Prometheus config through API, retrying: %s", err)
		return false
	}
	response, err := httpClient.Do(request)
	if err != nil {
		t.Logf("Unable to retrieve config through Prometheus API, retrying: %s", err)
		return false
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Logf("Unable to read Prometheus config response body, retrying: %s", err)
		return false
	}
	var result promResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Logf("Unable to unmarshal Prometheus config response body, retrying: %s", err)
		return false
	}

	for _, g := range result.Data.Groups {
		for _, r := range g.Rules {
			if r.Name == document.Spec.Indicators[0].Name &&
				strings.Contains(r.Query, document.Spec.Indicators[0].PromQL) {
				return true
			}
		}
	}
	return false
}

type promResult struct {
	Data struct {
		Groups []struct {
			Rules []struct {
				Name  string `json:"name"`
				Query string `json:"query"`
			} `json:"rules"`
		} `json:"groups"`
	} `json:"data"`
}

func grafanaConfigMapMatch(t *testing.T, dashboardFilename string, cm *coreV1.ConfigMap, id *v1.IndicatorDocument) bool {
	grafanaDashboard, err := grafana_dashboard.DocumentToDashboard(*id)
	if err != nil {
		t.Logf("Unable to convert to grafana dashboard: %s", err)
		return false
	}
	dashboard := grafanaDashboard
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

func prometheusConfigMapMatch(t *testing.T, cm *coreV1.ConfigMap, id *v1.IndicatorDocument) bool {
	alerts := prometheus_alerts.AlertDocumentFrom(*id)
	alerts.Groups[0].Name = id.Namespace + "/" + id.Name
	expected, err := yaml.Marshal(alerts)
	if err != nil {
		t.Logf("Unable to marshal: %s", err)
		return false
	}

	var (
		cmAlerts map[string][]map[string]interface{}
		cmAlert interface{}
	)
	err = yaml.Unmarshal([]byte(cm.Data["alerts"]), &cmAlerts)
	if err != nil {
		t.Logf("Unable to unmarshal: %s", err)
		return false
	}

	for _, group := range cmAlerts["groups"] {
		if group["name"] == alerts.Groups[0].Name {
			cmAlert = group
		}
	}
	if cmAlert == nil {
		t.Log("Unable to find alert group")
		return false
	}

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

func indicatorDocument(ns string) *v1.IndicatorDocument {

	indicatorName := fmt.Sprintf("e2e_test_indicator_%d", rand.Intn(math.MaxInt32))
	return &v1.IndicatorDocument{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("e2etest%d", rand.Intn(math.MaxInt32)),
			Namespace: ns,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: api_versions.V1,
		},
		Spec: v1.IndicatorDocumentSpec{
			Product: v1.Product{
				Name:    fmt.Sprintf("e2etestproduct%d", rand.Intn(math.MaxInt32)),
				Version: "v1.2.3-rc1",
			},
			Indicators: []v1.IndicatorSpec{
				{
					Name:   indicatorName,
					PromQL: "rate(some_metric[10m])",
					Alert: v1.Alert{
						For:  "5m",
						Step: "2m",
					},
					Thresholds: []v1.Threshold{
						{
							Level:    "critical",
							Operator: v1.GreaterThanOrEqualTo,
							Value:    float64(500),
						},
					},
				},
			},
			Layout: v1.Layout{
				Sections: []v1.Section{{
					Title:      "Metrics",
					Indicators: []string{indicatorName},
				}},
			},
		},
	}
}

func expandHome(s string) string {
	usr, err := user.Current()
	if err != nil {
		log.Panicf("Enable to expand user: %s", err)
	}
	return strings.Replace(s, "~", usr.HomeDir, -1)
}

func createNamespace(t *testing.T) (string, func()) {
	nsName := fmt.Sprintf("e2e-test-%d", rand.Intn(math.MaxInt32))
	ns := &coreV1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	t.Logf("Creating namespace: %s", nsName)
	nsr, err := clients.k8sClientset.CoreV1().Namespaces().Create(ns)
	if err != nil {
		t.Fatalf("Enable to create namespace: %s", err)
	}
	return nsName, func() {
		t.Logf("Deleting namespace: %s", nsName)
		err := clients.k8sClientset.CoreV1().Namespaces().Delete(nsName, &metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{
				UID: &nsr.UID,
			},
		})
		if err != nil {
			t.Fatalf("Enable to delete namespace: %s", err)
		}
	}
}

func grafanaDashboardFilename(id *v1.IndicatorDocument) string {
	return fmt.Sprintf("indicator-protocol-grafana-dashboard.%s.%s", id.Namespace, id.Name)
}
