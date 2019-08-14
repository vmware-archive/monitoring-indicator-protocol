package prometheus_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/prometheus"
)

func TestController(t *testing.T) {
	t.Run("it adds", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapPatcher := &spyConfigMapPatcher{g: g}
		spyConfigRenderer := &spyConfigRenderer{
			g:      g,
			config: "new-config",
		}
		p := prometheus.NewController(spyConfigMapPatcher, spyConfigRenderer)

		i := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring",
			},
		}

		p.OnAdd(i)

		spyConfigRenderer.assertUpsert(0, i)
		spyConfigMapPatcher.expectPatches([]string{
			"new-config",
		})
	})

	t.Run("it updates existing indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapPatcher := &spyConfigMapPatcher{g: g}
		spyConfigRenderer := &spyConfigRenderer{
			g:      g,
			config: "new-config",
		}
		p := prometheus.NewController(spyConfigMapPatcher, spyConfigRenderer)

		i1 := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring-1",
			},
		}
		i2 := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring-2",
			},
		}

		p.OnAdd(i1)
		p.OnUpdate(i1, i2)

		spyConfigRenderer.assertUpsert(0, i1)
		spyConfigRenderer.assertUpsert(1, i2)
		spyConfigMapPatcher.expectPatches([]string{
			"new-config",
			"new-config",
		})
	})

	t.Run("it does not upsert if the indicator is unchanged", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapPatcher := &spyConfigMapPatcher{g: g}
		spyConfigRenderer := &spyConfigRenderer{
			g:      g,
			config: "new-config",
		}
		p := prometheus.NewController(spyConfigMapPatcher, spyConfigRenderer)

		i1 := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring-1",
			},
		}
		i2 := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring-1",
			},
		}

		p.OnAdd(i1)
		p.OnUpdate(i1, i2)

		spyConfigRenderer.assertUpsert(0, i1)
		spyConfigRenderer.assertUpsertLen(1)
		spyConfigMapPatcher.expectPatches([]string{
			"new-config",
		})
	})

	t.Run("it deletes existing indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapPatcher := &spyConfigMapPatcher{g: g}
		spyConfigRenderer := &spyConfigRenderer{
			g:      g,
			config: "new-config",
		}
		p := prometheus.NewController(spyConfigMapPatcher, spyConfigRenderer)

		i := &v1.IndicatorDocument{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rabbit-mq-monitoring-1",
			},
		}

		p.OnDelete(i)

		spyConfigRenderer.assertDelete(0, i)
		spyConfigRenderer.assertDeleteLen(1)
		spyConfigMapPatcher.expectPatches([]string{
			"new-config",
		})
	})

	t.Run("it does nothing when given non-indicators", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spyConfigMapPatcher := &spyConfigMapPatcher{g: g}
		spyConfigRenderer := &spyConfigRenderer{g: g}
		p := prometheus.NewController(spyConfigMapPatcher, spyConfigRenderer)

		p.OnAdd(nil)
		p.OnAdd("nothing")
		p.OnAdd(42)

		p.OnUpdate(nil, nil)
		p.OnUpdate("nothing", "something")
		p.OnUpdate(42, 23)

		p.OnDelete(nil)
		p.OnDelete("nothing")
		p.OnDelete(42)

		spyConfigRenderer.assertUpsertLen(0)
		g.Expect(spyConfigMapPatcher.patchCalled).To(BeFalse())
	})
}

type jsonPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type patch struct {
	name string
	pt   types.PatchType
	data []byte
}

type spyConfigMapPatcher struct {
	g           *GomegaWithT
	patchCalled bool
	patches     []patch
}

func (s *spyConfigMapPatcher) Patch(
	name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*corev1.ConfigMap, error) {
	s.patchCalled = true
	s.patches = append(s.patches, patch{
		name: name,
		pt:   pt,
		data: data,
	})
	return nil, nil
}

func (s *spyConfigMapPatcher) expectPatches(patches []string) {
	s.g.Expect(s.patches).To(HaveLen(len(patches)))
	for i, p := range patches {
		jp := jsonPatch{
			Op:    "replace",
			Path:  "/data/alerts",
			Value: p,
		}
		b, err := json.Marshal([]jsonPatch{jp})
		s.g.Expect(err).ToNot(HaveOccurred())

		s.g.Expect(s.patches[i].name).To(Equal(prometheus.ConfigMapName))
		s.g.Expect(s.patches[i].pt).To(Equal(types.JSONPatchType))
		s.g.Expect(s.patches[i].data).To(MatchJSON(b))
	}
}

type spyConfigRenderer struct {
	g      *GomegaWithT
	config string
	upsert []*v1.IndicatorDocument
	delete []*v1.IndicatorDocument
}

func (s *spyConfigRenderer) Upsert(i *v1.IndicatorDocument) {
	s.upsert = append(s.upsert, i)
}

func (s *spyConfigRenderer) Delete(i *v1.IndicatorDocument) {
	s.delete = append(s.delete, i)
}

func (s *spyConfigRenderer) String() string {
	return s.config
}

func (s *spyConfigRenderer) assertUpsertLen(count int) {
	s.g.Expect(s.upsert).To(HaveLen(count))
}

func (s *spyConfigRenderer) assertUpsert(position int, indicator *v1.IndicatorDocument) {
	s.g.Expect(s.upsert[position]).To(Equal(indicator))
}

func (s *spyConfigRenderer) assertDelete(position int, indicator *v1.IndicatorDocument) {
	s.g.Expect(s.delete[position]).To(Equal(indicator))
}

func (s *spyConfigRenderer) assertDeleteLen(count int) {
	s.g.Expect(s.delete).To(HaveLen(count))
}
