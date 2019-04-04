package grafana

import (
	"log"
	"reflect"

	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapEditor interface {
	Create(*v1.ConfigMap) (*v1.ConfigMap, error)
	Update(*v1.ConfigMap) (*v1.ConfigMap, error)
	Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error)
}

type Controller struct {
	cmEditor ConfigMapEditor
}

func NewController(configMap ConfigMapEditor) *Controller {
	return &Controller{
		cmEditor: configMap,
	}
}

func (c *Controller) OnAdd(obj interface{}) {
	doc, ok := obj.(*v1alpha1.IndicatorDocument)
	if !ok {
		return
	}
	configMap, err := ConfigMap(doc, nil)
	if err != nil {
		log.Printf("Failed to generate ConfigMap: %s", err)
		return
	}

	if c.configMapAlreadyExists(configMap) {
		_, err = c.cmEditor.Update(configMap)
		if err != nil {
			log.Printf("Failed to update while adding ConfigMap: %s", err)
		}
		return
	}

	_, err = c.cmEditor.Create(configMap)
	if err != nil {
		log.Printf("Failed to create ConfigMap: %s", err)
		return
	}
}

// TODO: evaluate edge case where object might not exist
func (c *Controller) OnUpdate(oldObj, newObj interface{}) {
	newDoc, ok := newObj.(*v1alpha1.IndicatorDocument)
	if !ok {
		return
	}
	if oldObj != nil {
		oldDoc, ok := oldObj.(*v1alpha1.IndicatorDocument)
		if !ok {
			return
		}
		if reflect.DeepEqual(newDoc, oldDoc) {
			return
		}
	}
	configMap, err := ConfigMap(newDoc, nil)
	if err != nil {
		log.Printf("Failed to generate ConfigMap: %s", err)
		return
	}
	_, err = c.cmEditor.Update(configMap)
	if err != nil {
		log.Printf("Failed to update ConfigMap: %s", err)
		return
	}
}

// TODO: evaluate edge case where object might not exist
// TODO: do we need to handle non-indicatordocuments?
func (c *Controller) OnDelete(obj interface{}) {
	doc, ok := obj.(*v1alpha1.IndicatorDocument)
	if !ok {
		log.Printf("OnDelete received a non-indicatordocument: %T", obj)
		return
	}
	log.Printf("Deleting Grafana config map for %s on namespace %s", doc.Name, doc.Namespace)
}

func (c *Controller) configMapAlreadyExists(configMap *v1.ConfigMap) bool {
	_, err := c.cmEditor.Get(configMap.Name, metav1.GetOptions{})

	return err == nil
}
