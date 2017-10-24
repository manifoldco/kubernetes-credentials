package crd

import (
	"fmt"
	"time"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CreateCRD is a wrapper to create a CRD from scratch with a set of params.
func CreateCRD(cs apiextensionsclient.Interface, name, plural, group, version string) error {
	fullName := plural + "." + version

	crd := &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fullName,
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Scope:   apiextv1beta1.NamespaceScoped,
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   name,
			},
		},
	}
	_, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return err
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(fullName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextv1beta1.Established:
				if cond.Status == apiextv1beta1.ConditionTrue {
					return true, err
				}
			case apiextv1beta1.NamesAccepted:
				if cond.Status == apiextv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})

	if err != nil {
		deleteErr := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(fullName, nil)
		if deleteErr != nil {
			return errors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}

	return nil
}
