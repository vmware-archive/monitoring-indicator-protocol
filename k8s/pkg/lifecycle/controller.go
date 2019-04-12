package lifecycle

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	types "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/indicatordocument/v1alpha1"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned/typed/indicatordocument/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var trueVal = true

type Controller struct {
	client v1alpha1.IndicatorsGetter
}

func NewController(idGetter v1alpha1.IndicatorsGetter) Controller {
	return Controller{
		client: idGetter,
	}
}

func (c Controller) OnAdd(obj interface{}) {
	doc, ok := obj.(*types.IndicatorDocument)
	if !ok {
		log.Printf("Invalid resource type OnAdd: %T", obj)
		return
	}
	log.Printf("Adding indicators for %s on namespace %s", doc.Name, doc.Namespace)
	for _, indicator := range doc.Spec.Indicators {
		id := toIndicator(indicator, doc)
		_, err := c.client.Indicators(id.Namespace).Get(id.Name, metav1.GetOptions{})
		if err == nil {
			log.Printf("Indicator was already created. Skipping creation of %s", id.Name)
			continue
		}

		createNewIndicator(id, c, doc)
	}
}

func (c Controller) OnUpdate(oldObj, newObj interface{}) {
	oldDoc, ok := oldObj.(*types.IndicatorDocument)
	newDoc, ok := newObj.(*types.IndicatorDocument)
	if !ok {
		log.Printf("Invalid resource type OnUpdate: %T", newObj)
		return
	}
	log.Printf("Updating indicators for %s on namespace %s", newDoc.Name, newDoc.Namespace)
	existingList, err := c.client.Indicators(oldDoc.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s-%s", oldDoc.Name, oldDoc.Namespace),
	})
	if err != nil {
		log.Printf("Error encountered when querying for existing Indicators by label: %s", err)
	}
	for _, newIndicatorSpec := range newDoc.Spec.Indicators {
		foundIndicator := find(existingList.Items, newIndicatorSpec)
		newIndicator := toIndicator(newIndicatorSpec, newDoc)
		if foundIndicator == nil {
			createNewIndicator(newIndicator, c, oldDoc)
		} else {
			updateExistingIndicator(newIndicator, foundIndicator, c, oldDoc)
		}
	}

	for _, oldIndicator := range existingList.Items {
		if !contains(newDoc.Spec.Indicators, oldIndicator.Spec) {
			deleteOldIndicator(oldIndicator, c, oldDoc)
		}
	}
}

func deleteOldIndicator(oldIndicator types.Indicator, c Controller, oldDoc *types.IndicatorDocument) {
	log.Printf("Deleting indicator %s", oldIndicator.Name)
	err := c.client.Indicators(oldDoc.Namespace).Delete(oldIndicator.Name, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error encountered when deleting indicator %s: %s", oldIndicator.Name, err)
	}
}

func updateExistingIndicator(newIndicator *types.Indicator, foundIndicator *types.Indicator, c Controller, oldDoc *types.IndicatorDocument) {
	newIndicator.ResourceVersion = foundIndicator.ResourceVersion
	if !reflect.DeepEqual(foundIndicator.Spec, newIndicator.Spec) {
		log.Printf("Updating indicator %s", newIndicator.Spec.Name)
		_, err := c.client.Indicators(oldDoc.Namespace).Update(newIndicator)
		if err != nil {
			log.Printf("Error encountered when updating indicator %s: %s", newIndicator.Spec.Name, err)
		}
	}
}

func createNewIndicator(newIndicator *types.Indicator, c Controller, oldDoc *types.IndicatorDocument) {
	log.Printf("Creating indicator %s", newIndicator.Spec.Name)
	_, err := c.client.Indicators(oldDoc.Namespace).Create(newIndicator)
	if err != nil {
		log.Printf("Error encountered when creating indicator %s: %s", newIndicator.Spec.Name, err)
	}
}

// TODO: do we need to handle non-indicatordocuments?
func (c Controller) OnDelete(obj interface{}) {
	doc, ok := obj.(*types.IndicatorDocument)
	if !ok {
		log.Printf("Invalid resource type OnDelete: %T", obj)
		return
	}
	log.Printf("Deleting indicators for %s on namespace %s", doc.Name, doc.Namespace)
}

func toIndicator(is types.IndicatorSpec, parent *types.IndicatorDocument) *types.Indicator {
	name := parent.Name + "-" + strings.ReplaceAll(is.Name, "_", "-")
	indicator := is.DeepCopy()
	indicator.Product = fmt.Sprintf("%s %s", parent.Spec.Product.Name, parent.Spec.Product.Version)
	return &types.Indicator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: parent.Namespace,
			Labels: map[string]string{
				"owner": parent.Name + "-" + parent.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "apps.pivotal.io/v1alpha1",
					Kind:               "IndicatorDocument",
					Name:               parent.Name,
					UID:                parent.UID,
					Controller:         &trueVal,
					BlockOwnerDeletion: nil,
				},
			},
		},
		Spec: *indicator,
	}
}

func find(list []types.Indicator, item types.IndicatorSpec) *types.Indicator {
	for _, listItem := range list {
		if listItem.Spec.Name == item.Name {
			return &listItem
		}
	}

	return nil
}

func contains(list []types.IndicatorSpec, item types.IndicatorSpec) bool {
	for _, listItem := range list {
		if listItem.Name == item.Name {
			return true
		}
	}

	return false
}
