package indicator_status_test

import (
	"bytes"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	. "github.com/onsi/gomega"
	types "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/test_fixtures"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestController(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	log.SetOutput(buffer)

	t.Run("adds all preexisting indicators on Start()", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := indicator_status.NewIndicatorStore()
		fakeIndicatorsGetter := &fakeIndicatorsGetter{listedIndicators: []types.Indicator{
			test_fixtures.Indicator("my-indicator", "rate(love[8m])"),
			test_fixtures.Indicator("my-indicator2", "rate(love[5m])"),
		},
			store: store,
		}
		fakePromqlClient := &fakePromqlClient{response: []float64{float64(-1)}}
		mockClock := clock.NewMock()
		c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)

		go c.Start()

		g.Eventually(fakeIndicatorsGetter.getUpdateCalls).Should(HaveLen(2))
	})

	t.Run("can call Start concurrently with another call", func(t *testing.T) {
		g := NewGomegaWithT(t)
		store := indicator_status.NewIndicatorStore()
		fakeIndicatorsGetter := &fakeIndicatorsGetter{listedIndicators: []types.Indicator{
			test_fixtures.Indicator("my-indicator", "rate(love[8m])"),
			test_fixtures.Indicator("my-indicator2", "rate(love[5m])"),
		}, store: store}
		fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
		mockClock := clock.NewMock()
		anotherIndicator := test_fixtures.Indicator("my-indicator3", "rate(love[5m])")
		c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)

		go c.OnDelete(&anotherIndicator)
		go c.Start()

		g.Eventually(fakeIndicatorsGetter.getUpdateCalls).Should(HaveLen(2))
	})

	t.Run("OnAdd", func(t *testing.T) {
		t.Run("starts updating indicator status", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)
			indicator := test_fixtures.Indicator("name", "rate(errors[5m])")

			go c.Start()
			c.OnAdd(&indicator)
			mockClock.Add(3 * time.Second)

			g.Consistently(fakePromqlClient.getQueries).Should(ContainElement("rate(errors[5m])"))
			g.Consistently(fakeIndicatorsGetter.getUpdateCalls).Should(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Name).To(Equal("name"))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Status.Phase).To(Equal("critical"))
		})

		t.Run("it updates indicator status to UNDEFINED when there is no threshold", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)

			indicator := types.Indicator{
				ObjectMeta: v1.ObjectMeta{
					Name: "a name",
				},
				Spec: types.IndicatorSpec{
					Product:    "??",
					Name:       "test",
					Promql:     "rate(errors[5m])",
					Thresholds: []types.Threshold{},
				},
			}

			go c.Start()
			c.OnAdd(&indicator)
			mockClock.Add(time.Second)

			g.Expect(fakePromqlClient.getQueries()).To(HaveLen(0))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Name).To(Equal("a name"))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Status.Phase).To(Equal("UNDEFINED"))
		})
	})

	t.Run("OnDelete", func(t *testing.T) {
		t.Run("stops updating indicator status", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)
			indicator := test_fixtures.Indicator("name", "rate(errors[10m])")

			go c.Start()
			c.OnAdd(&indicator)
			mockClock.Add(time.Second)

			g.Expect(fakePromqlClient.getQueries()).To(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(1))

			fakePromqlClient.resetQueryArgs()
			fakeIndicatorsGetter.resetUpdateCalls()
			c.OnDelete(&indicator)
			mockClock.Add(time.Second)

			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(0))
			g.Expect(fakePromqlClient.getQueries()).To(Not(ContainElement(indicator.Spec.Promql)))
		})

		t.Run("deleting non-existent indicator is a noop", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)

			indicator1 := test_fixtures.Indicator("name", "rate(errors[10m])")
			indicator2 := test_fixtures.Indicator("new-name", "rate(errors[5m])")

			go c.Start()

			c.OnAdd(&indicator1)
			c.OnDelete(&indicator2)
			mockClock.Add(time.Second)

			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(1))
			g.Expect(fakePromqlClient.getQueries()).To(ContainElement(indicator1.Spec.Promql))
		})
	})

	t.Run("OnUpdate", func(t *testing.T) {
		t.Run("updates indicator status to UNDEFINED when threshold is removed", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)
			indicator := test_fixtures.Indicator("name", "rate(errors[10m])")

			go c.Start()
			c.OnAdd(&indicator)
			mockClock.Add(time.Second)
			newIndicator := indicator.DeepCopy()
			newIndicator.Spec.Thresholds = []types.Threshold{}

			fakePromqlClient.resetQueryArgs()
			fakeIndicatorsGetter.resetUpdateCalls()
			c.OnUpdate(&indicator, newIndicator)
			mockClock.Add(time.Second)

			g.Expect(fakePromqlClient.getQueries()).To(Not(ContainElement(newIndicator.Spec.Promql)))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Name).To(Equal("name"))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Status.Phase).To(Equal("UNDEFINED"))
		})

		t.Run("updates indicator status when threshold is added", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)
			indicator := types.Indicator{
				ObjectMeta: v1.ObjectMeta{
					Name: "new name",
				},
				Spec: types.IndicatorSpec{
					Product:    "??",
					Name:       "test",
					Promql:     "rate(errors[5m])",
					Thresholds: []types.Threshold{},
				},
			}

			go c.Start()

			c.OnAdd(&indicator)
			mockClock.Add(time.Second)

			thresholdLevel := float64(0)

			newIndicator := indicator.DeepCopy()
			newIndicator.Spec.Thresholds = []types.Threshold{{
				Level: "critical",
				Gt:    &thresholdLevel,
			}}

			fakePromqlClient.resetQueryArgs()
			fakeIndicatorsGetter.resetUpdateCalls()
			c.OnUpdate(&indicator, newIndicator)
			mockClock.Add(time.Second)

			g.Consistently(fakeIndicatorsGetter.getUpdateCalls()).Should(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Status.Phase).To(Equal("HEALTHY"))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Name).To(Equal("new name"))
			g.Expect(fakePromqlClient.getQueries()).To(ContainElement(newIndicator.Spec.Promql))
		})

		t.Run("updates indicator status with new threshold when threshold is changed", func(t *testing.T) {
			g := NewGomegaWithT(t)
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{store: store}
			fakePromqlClient := &fakePromqlClient{response: []float64{0}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)
			indicator := test_fixtures.Indicator("my-fave-indicator", "rate(error[6m])")

			go c.Start()

			c.OnAdd(&indicator)
			mockClock.Add(time.Second)

			thresholdLevel := float64(0)

			newIndicator := indicator.DeepCopy()
			newIndicator.Spec.Thresholds = []types.Threshold{{
				Level: "pamplemousse",
				Gte:   &thresholdLevel,
			}}

			fakePromqlClient.resetQueryArgs()
			fakeIndicatorsGetter.resetUpdateCalls()
			c.OnUpdate(&indicator, newIndicator)
			mockClock.Add(time.Second)

			g.Consistently(fakeIndicatorsGetter.getUpdateCalls()).Should(HaveLen(1))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Status.Phase).To(Equal("pamplemousse"))
			g.Expect(fakeIndicatorsGetter.getUpdateCalls()[0].Name).To(Equal("my-fave-indicator"))
			g.Expect(fakePromqlClient.getQueries()).To(ContainElement(newIndicator.Spec.Promql))
		})

		t.Run("does not update status when it has not changed", func(t *testing.T) {
			g := NewGomegaWithT(t)
			indicator := test_fixtures.Indicator("my-fave-indicator", "rate(error[6m])")
			indicator.Status = types.IndicatorStatus{
				Phase: "critical",
			}
			store := indicator_status.NewIndicatorStore()
			fakeIndicatorsGetter := &fakeIndicatorsGetter{
				listedIndicators: []types.Indicator{indicator},
				store: store,
			}
			fakePromqlClient := &fakePromqlClient{response: []float64{-1}}
			mockClock := clock.NewMock()
			c := indicator_status.NewController(fakeIndicatorsGetter, fakePromqlClient, time.Second, mockClock, "cool-namespace-name", store)

			go c.Start()
			c.OnAdd(&indicator)
			mockClock.Add(time.Second)

			g.Expect(fakeIndicatorsGetter.getUpdateCalls()).To(HaveLen(0))
		})
	})
}

//********** Fake Prometheus Client **********//

type fakePromqlClient struct {
	response []float64
	queries  []string
	sync.Mutex
}

func (s *fakePromqlClient) QueryVectorValues(query string) ([]float64, error) {
	s.Lock()
	defer s.Unlock()
	s.queries = append(s.queries, query)

	return s.response, nil
}

func (s *fakePromqlClient) getQueries() []string {
	s.Lock()
	defer s.Unlock()
	return s.queries
}

func (s *fakePromqlClient) resetQueryArgs() {
	s.Lock()
	defer s.Unlock()
	s.queries = make([]string, 0)
}

type fakeIndicatorsGetter struct {
	v1alpha1.IndicatorInterface
	listedIndicators []types.Indicator

	updateCalls []*types.Indicator
	sync.Mutex
	store *indicator_status.IndicatorStore
}

//********** Fake Indicators Getter **********//
func (s *fakeIndicatorsGetter) Indicators(string) v1alpha1.IndicatorInterface {
	return s
}

//********** Fake Indicator Interface **********//
func (s *fakeIndicatorsGetter) Update(i *types.Indicator) (*types.Indicator, error) {
	s.Lock()
	defer s.Unlock()
	s.updateCalls = append(s.updateCalls, i)
	s.store.Update(*i)
	return nil, nil
}

func (s *fakeIndicatorsGetter) List(opts v1.ListOptions) (*types.IndicatorList, error) {
	return &types.IndicatorList{
		TypeMeta: v1.TypeMeta{},
		ListMeta: v1.ListMeta{},
		Items:    s.listedIndicators,
	}, nil
}

func (s *fakeIndicatorsGetter) resetUpdateCalls() {
	s.Lock()
	defer s.Unlock()
	s.updateCalls = make([]*types.Indicator, 0)
}

func (s *fakeIndicatorsGetter) getUpdateCalls() []*types.Indicator {
	s.Lock()
	defer s.Unlock()
	return s.updateCalls
}
