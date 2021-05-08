package main

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"

	"github.com/fatih/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	meta_util "kmodules.xyz/client-go/meta"
)

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc2 := dynamic.NewForConfigOrDie(config)

}

func statusEqual(old, new interface{}) bool {
	oldStatus, oldExists := extractStatusFromObject(old)
	newStatus, newExists := extractStatusFromObject(new)
	if oldExists && newExists {
		result := reflect.DeepEqual(oldStatus, newStatus)
		if !result {
			diff, err := meta_util.JsonDiff(oldStatus, newStatus)
			if err == nil {
				fmt.Println(diff)
			} else {
				panic(err)
			}
		}
		return result
	}
	return !oldExists && !newExists
}

func extractStatusFromObject(o interface{}) (interface{}, bool) {
	switch obj := o.(type) {
	case *unstructured.Unstructured:
		v, ok, _ := unstructured.NestedFieldNoCopy(obj.Object, "status")
		return v, ok
	case metav1.Object:
		st := structs.New(obj)
		field, ok := st.FieldOk("Status")
		if !ok {
			return nil, ok
		}
		return field.Value(), true
	}
	panic(fmt.Errorf("unknown object %v", reflect.TypeOf(o)))
}
