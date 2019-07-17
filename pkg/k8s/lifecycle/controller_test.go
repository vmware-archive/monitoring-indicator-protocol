package lifecycle_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	types "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/lifecycle"
)

func TestController(t *testing.T) {
	t.Run("OnAdd", func(t *testing.T) {
		t.Run("creates multiple indicators", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)
			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-1"},
						{Name: "I2", PromQL: "promql-query-2"},
						{Name: "I3", PromQL: "promql-query-3"},
					},
				},
			}

			c.OnAdd(id)

			spyIndicatorsGetter.expectCreated(id.Spec.Indicators)
		})

		t.Run("takes no action if provided non-indicatordocument", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)
			id := 5

			c.OnAdd(id)

			spyIndicatorsGetter.expectThatNothingWasCreated()
		})

		t.Run("indicators are controlled by IndicatorDocuments", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)
			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
					UID:       "test_uid",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-1"},
					},
				},
			}

			c.OnAdd(id)

			truePtr := true
			g.Expect(spyIndicatorsGetter.createCalls[0].OwnerReferences).To(ConsistOf(v1.OwnerReference{
				APIVersion: "apps.pivotal.io/v1alpha1",
				Kind:       "IndicatorDocument",
				Name:       id.Name,
				UID:        id.UID,
				Controller: &truePtr,
			}))
		})

		t.Run("takes no action if indicator already exists", func(t *testing.T) {
			g := NewGomegaWithT(t)
			existingIndicatorSpec := types.IndicatorSpec{Product: "rabbit v1.2.3", Name: "I1", PromQL: "promql-query-1"}
			existingIndicatorList := types.IndicatorList{
				Items: []types.Indicator{{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-i1",
					},
					Spec: existingIndicatorSpec,
				}},
			}
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t, indicatorList: &existingIndicatorList}
			c := lifecycle.NewController(spyIndicatorsGetter)

			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
					UID:       "test_uid",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						existingIndicatorSpec,
					},
				},
			}

			c.OnAdd(id)
			spyIndicatorsGetter.expectThatNothingWasCreated()
		})

		t.Run("adds the appropriate labels", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)

			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
					UID:       "test_uid",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-1"},
					},
				},
			}

			c.OnAdd(id)

			expectedLabel := id.Name + "-" + id.Namespace
			g.Expect(spyIndicatorsGetter.createCalls[0].Labels["owner"]).To(Equal(expectedLabel))
		})

		t.Run("sanitizes indicator name", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)

			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
					UID:       "test_uid",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1_foo:Goo", PromQL: "promql-query-1"},
					},
				},
			}

			c.OnAdd(id)

			g.Expect(spyIndicatorsGetter.createCalls[0].Name).To(Equal("test-i1-foo-goo"))
		})

		t.Run("adds product info to indicator", func(t *testing.T) {
			g := NewGomegaWithT(t)
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t}
			c := lifecycle.NewController(spyIndicatorsGetter)

			id := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
					UID:       "test_uid",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1_foo", PromQL: "promql-query-1"},
					},
				},
			}

			c.OnAdd(id)

			g.Expect(spyIndicatorsGetter.createCalls[0].Spec.Product).To(Equal("rabbit v1.2.3"))
		})
	})

	t.Run("OnUpdate", func(t *testing.T) {
		t.Run("updates indicator if it exists", func(t *testing.T) {
			g := NewGomegaWithT(t)
			existingIndicatorList := types.IndicatorList{
				Items: []types.Indicator{{
					ObjectMeta: v1.ObjectMeta{
						ResourceVersion: "my-favorite-resource",
					},
					Spec: types.IndicatorSpec{
						Product: "rabbit v1.2.3",
						Name:    "I1",
						PromQL:  "promql-query-4",
					},
				}},
			}
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t, indicatorList: &existingIndicatorList}
			c := lifecycle.NewController(spyIndicatorsGetter)
			oldIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-4"},
					},
				},
			}
			newIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "new-promql-query-4"},
					},
				},
			}

			c.OnUpdate(oldIndicatorDoc, newIndicatorDoc)

			g.Expect(spyIndicatorsGetter.updateCalls).To(HaveLen(1))
			g.Expect(spyIndicatorsGetter.updateCalls[0].ResourceVersion).To(Equal("my-favorite-resource"))
			g.Expect(spyIndicatorsGetter.updateCalls[0].Spec.Name).To(Equal(oldIndicatorDoc.Spec.Indicators[0].Name))
			g.Expect(spyIndicatorsGetter.updateCalls[0].Spec.PromQL).To(Equal(newIndicatorDoc.Spec.Indicators[0].PromQL))

			g.Expect(spyIndicatorsGetter.createCalls).To(HaveLen(0))
		})

		t.Run("creates an indicator if it doesn't exist", func(t *testing.T) {
			g := NewGomegaWithT(t)
			existingIndicatorList := types.IndicatorList{
				Items: []types.Indicator{{
					Spec: types.IndicatorSpec{
						Product: "rabbit v1.2.3",
						Name:    "I1",
						PromQL:  "promql-query-4",
					},
				}},
			}
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t, indicatorList: &existingIndicatorList}
			c := lifecycle.NewController(spyIndicatorsGetter)
			oldIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-4"},
					},
				},
			}
			newIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-4"},
						{Name: "I2", PromQL: "new-promql-query-4"},
					},
				},
			}

			c.OnUpdate(oldIndicatorDoc, newIndicatorDoc)

			g.Expect(spyIndicatorsGetter.createCalls).To(HaveLen(1))
			g.Expect(spyIndicatorsGetter.createCalls[0].Spec.Name).To(Equal("I2"))
			g.Expect(spyIndicatorsGetter.createCalls[0].Spec.PromQL).To(Equal("new-promql-query-4"))

			g.Expect(spyIndicatorsGetter.updateCalls).To(HaveLen(0))
		})

		t.Run("deletes an indicator if it was removed from the IndicatorDocument", func(t *testing.T) {
			g := NewGomegaWithT(t)
			existingIndicatorList := types.IndicatorList{
				Items: []types.Indicator{{
					ObjectMeta: v1.ObjectMeta{Name: "test-i1"},
					Spec: types.IndicatorSpec{
						Product: "rabbit v1.2.3",
						Name:    "I1",
						PromQL:  "promql-query-1",
					},
				}, {
					ObjectMeta: v1.ObjectMeta{Name: "test-i2"},
					Spec: types.IndicatorSpec{
						Product: "rabbit v1.2.3",
						Name:    "I2",
						PromQL:  "promql-query-2",
					},
				}},
			}
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t, indicatorList: &existingIndicatorList}
			c := lifecycle.NewController(spyIndicatorsGetter)
			oldIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-1"},
						{Name: "I2", PromQL: "promql-query-2"},
					},
				},
			}
			newIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "I1", PromQL: "promql-query-1"},
					},
				},
			}

			c.OnUpdate(oldIndicatorDoc, newIndicatorDoc)

			g.Expect(spyIndicatorsGetter.deleteCalls).To(HaveLen(1))
			g.Expect(spyIndicatorsGetter.deleteCalls[0]).To(Equal("test-i2"))

			g.Expect(spyIndicatorsGetter.updateCalls).To(HaveLen(0))
			g.Expect(spyIndicatorsGetter.createCalls).To(HaveLen(0))
		})

		t.Run("sanitizes indicator name", func(t *testing.T) {
			g := NewGomegaWithT(t)
			existingIndicatorList := types.IndicatorList{
				Items: []types.Indicator{{
					ObjectMeta: v1.ObjectMeta{Name: "test-boo-i1-foo"},
					Spec: types.IndicatorSpec{
						Product: "rabbit v1.2.3",
						Name:    "boo_I1_foo:Goo",
						PromQL:  "promql-query-4",
					},
					Status: types.IndicatorStatus{},
				}},
			}
			spyIndicatorsGetter := &spyIndicatorsGetter{g: g, t: t, indicatorList: &existingIndicatorList}
			c := lifecycle.NewController(spyIndicatorsGetter)
			oldIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "boo_I1_foo:Goo", PromQL: "promql-query-4"},
					},
				},
			}
			newIndicatorDoc := &types.IndicatorDocument{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "test-namespace",
				},
				Spec: types.IndicatorDocumentSpec{
					Product: types.Product{
						Name:    "rabbit",
						Version: "v1.2.3",
					},
					Indicators: []types.IndicatorSpec{
						{Name: "boo_I1_foo:Goo", PromQL: "new-promql-query-4"},
					},
				},
			}

			c.OnUpdate(oldIndicatorDoc, newIndicatorDoc)

			g.Expect(spyIndicatorsGetter.updateCalls).To(HaveLen(1))
			g.Expect(spyIndicatorsGetter.updateCalls[0].Name).To(Equal("test-boo-i1-foo-goo"))
		})
	})
}

type spyIndicatorsGetter struct {
	v1alpha1.IndicatorInterface

	g *GomegaWithT
	t *testing.T

	indicatorList *types.IndicatorList

	createCalls []*types.Indicator
	updateCalls []*types.Indicator
	deleteCalls []string
}

func (s *spyIndicatorsGetter) Indicators(string) v1alpha1.IndicatorInterface {
	return s
}

func (s *spyIndicatorsGetter) Create(i *types.Indicator) (*types.Indicator, error) {
	s.createCalls = append(s.createCalls, i)
	return nil, nil
}

func (s *spyIndicatorsGetter) Update(i *types.Indicator) (*types.Indicator, error) {
	s.updateCalls = append(s.updateCalls, i)
	return nil, nil
}

func (s *spyIndicatorsGetter) List(opts v1.ListOptions) (*types.IndicatorList, error) {
	return s.indicatorList, nil
}

func (s *spyIndicatorsGetter) Delete(name string, options *v1.DeleteOptions) error {
	s.deleteCalls = append(s.deleteCalls, name)
	return nil
}

func (s *spyIndicatorsGetter) Get(name string, _ v1.GetOptions) (*types.Indicator, error) {
	indicator := types.Indicator{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{},
		Spec:       types.IndicatorSpec{},
	}
	if s.indicatorList == nil {
		return nil, errors.New("not found")
	}
	for _, i := range s.indicatorList.Items {
		if i.Name == name {
			return &indicator, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *spyIndicatorsGetter) expectCreated(indicators []types.IndicatorSpec) {
	s.t.Helper()
	s.g.Expect(s.createCalls).To(HaveLen(len(indicators)))
	for i, ind := range indicators {
		s.g.Expect(s.createCalls[i].Spec.Name).To(Equal(ind.Name))
		s.g.Expect(s.createCalls[i].Spec.PromQL).To(Equal(ind.PromQL))
	}
}

func (s *spyIndicatorsGetter) expectThatNothingWasCreated() {
	s.g.Expect(s.createCalls).To(BeNil())
}
