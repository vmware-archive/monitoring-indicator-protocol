package grafana

import (
	"log"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

type ConfigMapEditor interface {
	Create(*corev1.ConfigMap) (*corev1.ConfigMap, error)
	Update(*corev1.ConfigMap) (*corev1.ConfigMap, error)
	Get(name string, options metav1.GetOptions) (*corev1.ConfigMap, error)
	Delete(name string, options *metav1.DeleteOptions) error
}

type Controller struct {
	cmEditor      ConfigMapEditor
	indicatorType v1.IndicatorType
}

func NewController(configMap ConfigMapEditor, indicatorType v1.IndicatorType) *Controller {
	return &Controller{
		cmEditor:      configMap,
		indicatorType: indicatorType,
	}
}

func (c *Controller) OnAdd(obj interface{}) {
	doc, ok := obj.(*v1.IndicatorDocument)
	if !ok {
		return
	}
	configMap, err := ConfigMap(doc, nil, c.indicatorType)
	if err != nil {
		log.Print("Failed to generate ConfigMap")
		return
	}

	if configMap == nil {
		c.deleteConfigMapIfExists(doc)
	} else if c.configMapAlreadyExists(GenerateObjectName(doc)) {
		c.updateConfigMap(configMap)
	} else {
		c.createConfigMap(configMap)
	}
}

// TODO: evaluate edge case where object might not exist
func (c *Controller) OnUpdate(oldObj, newObj interface{}) {
	newDoc, ok := newObj.(*v1.IndicatorDocument)
	if !ok {
		return
	}

	var oldDoc *v1.IndicatorDocument
	if oldObj != nil {
		oldDoc, ok = oldObj.(*v1.IndicatorDocument)
		if !ok {
			return
		}
		if reflect.DeepEqual(newDoc, oldDoc) {
			return
		}
	}
	configMap, err := ConfigMap(newDoc, nil, c.indicatorType)
	if err != nil {
		log.Printf("Failed to generate ConfigMap: %s", err)
		return
	}

	if configMap == nil {
		c.deleteConfigMapIfExists(oldDoc)
	} else if c.configMapAlreadyExists(configMap.Name) {
		c.updateConfigMap(configMap)
	} else {
		c.createConfigMap(configMap)
	}
}

// TODO: evaluate edge case where object might not exist
// TODO: do we need to handle non-indicatordocuments?
func (c *Controller) OnDelete(obj interface{}) {
	_, ok := obj.(*v1.IndicatorDocument)
	if !ok {
		log.Print("OnDelete received a non-indicatordocument")
		return
	}
	log.Print("Deleting Grafana config map")
}

func (c *Controller) createConfigMap(configMap *corev1.ConfigMap) {
	_, err := c.cmEditor.Create(configMap)
	if err != nil {
		log.Print("Failed to create ConfigMap")
	}
}

func (c *Controller) updateConfigMap(configMap *corev1.ConfigMap) {
	_, err := c.cmEditor.Update(configMap)
	if err != nil {
		log.Print("Failed to update while adding ConfigMap")
	}
}

func (c *Controller) deleteConfigMapIfExists(doc *v1.IndicatorDocument) {
	if c.configMapAlreadyExists(GenerateObjectName(doc)) {
		err := c.cmEditor.Delete(GenerateObjectName(doc), &metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete ConfigMap in add: %s", err)
		}
	}
}

func (c *Controller) configMapAlreadyExists(configMapName string) bool {
	_, err := c.cmEditor.Get(configMapName, metav1.GetOptions{})

	return err == nil
}
