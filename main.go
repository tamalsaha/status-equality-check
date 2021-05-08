package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/fatih/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	meta_util "kmodules.xyz/client-go/meta"
)

type Condition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
}

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc := dynamic.NewForConfigOrDie(config)

	// kube-system          kube-apiserver-kind-control-plane
	gvrPod := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	pod, err := dc.Resource(gvrPod).Namespace("kube-system").Get(context.TODO(), "kube-apiserver-kind-control-plane", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	v := pod.Object["rt"]
	fmt.Println(reflect.TypeOf(v).Kind())
}

func StatusEqual(old, new interface{}) bool {
	oldStatus, oldExists := extractStatusFromObject(old)
	newStatus, newExists := extractStatusFromObject(new)
	if oldExists && newExists {
		oldKind := reflect.TypeOf(oldStatus).Kind()
		newKind := reflect.TypeOf(newStatus).Kind()
		if oldKind != newKind {
			klog.Warningf("old status kind %s does not match new status kind %s", oldKind, newKind)
			return false
		}

		var result bool
		if oldKind == reflect.Map {
			result = StatusMapEqual(oldStatus.(map[string]interface{}), newStatus.(map[string]interface{}))
		} else {
			oldStruct := structs.New(oldStatus)
			oldStruct.TagName = "json"

			newStruct := structs.New(newStatus)
			newStruct.TagName = "json"

			// map does not handle nested maps?
			result = StatusMapEqual(oldStruct.Map(), newStruct.Map())
		}
		if !result {
			if diff, err := meta_util.JsonDiff(oldStatus, newStatus); err == nil {
				klog.V(8).Infoln(diff)
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

func condEqual(old, nu []Condition) bool {
	// optimization
	if len(old) != len(nu) {
		return false
	}
	oldmap := make(map[string]Condition, len(old))
	for _, c := range old {
		oldmap[c.Type] = c
	}
	for _, c := range nu {
		e, ok := oldmap[c.Type]
		if !ok {
			return false
		}
		if e != c {
			return false
		}
	}
	return true
}

func StatusMapEqual(old, nu map[string]interface{}) bool {
	// optimization
	if len(old) != len(nu) {
		return false
	}

	for key, oldVal := range old {
		newVal, ok := nu[key]
		if !ok {
			return false
		}
		if key == "conditions" {
			// special case
			oldCond := make([]Condition, 0)
			if err := meta_util.DecodeObject(oldVal, &oldCond); err != nil {
				klog.Errorln(err)
				return false
			}
			nuCond := make([]Condition, 0)
			if err := meta_util.DecodeObject(newVal, &nuCond); err != nil {
				klog.Errorln(err)
				return false
			}
			if !condEqual(oldCond, nuCond) {
				return false
			}
		} else if !reflect.DeepEqual(oldVal, newVal) {
			return false
		}
	}

	for key := range nu {
		if _, ok := old[key]; !ok {
			return false
		}
	}
	return true
}
