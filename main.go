package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"time"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"

	"github.com/fatih/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	meta_util "kmodules.xyz/client-go/meta"
)

var d11 = &apps.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: apps.SchemeGroupVersion.String(),
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "demo",
		Name:      "d1",
	},
	Spec: apps.DeploymentSpec{
		Replicas: pointer.Int32P(3),
	},
	Status: apps.DeploymentStatus{
		ObservedGeneration:  2,
		Replicas:            3,
		UpdatedReplicas:     3,
		ReadyReplicas:       3,
		AvailableReplicas:   3,
		UnavailableReplicas: 0,
		Conditions: []apps.DeploymentCondition{
			{
				Type:               "Available",
				Status:             core.ConditionTrue,
				LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
				LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
				Reason:             "MinimumReplicasAvailable",
				Message:            "Deployment has minimum availability.",
			},
			{
				Type:               "Progressing",
				Status:             core.ConditionFalse,
				LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
				LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
				Reason:             "NewReplicaSetAvailable",
				Message:            "ReplicaSet \"d1\" has successfully progressed.",
			},
		},
	},
}

func main() {
	data, err := json.MarshalIndent(d11, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc := dynamic.NewForConfigOrDie(config)

	// kube-system          coredns
	gvrDeploy := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	d1, err := dc.Resource(gvrDeploy).Namespace("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(StatusEqual(d1, d1))

	kc := kubernetes.NewForConfigOrDie(config)
	d2, err := kc.AppsV1().Deployments("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(StatusEqual(d2, d2))
}

type Condition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
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
			result = statusMapEqual(oldStatus.(map[string]interface{}), newStatus.(map[string]interface{}))
		} else {
			oldStruct := structs.New(oldStatus)
			oldStruct.TagName = "json"

			newStruct := structs.New(newStatus)
			newStruct.TagName = "json"

			result = statusMapEqual(oldStruct.Map(), newStruct.Map())
		}
		if !result && klog.V(8).Enabled() {
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

func conditionsEqual(old, nu []Condition) bool {
	// optimization
	if len(old) != len(nu) {
		return false
	}
	oldMap := make(map[Condition]bool, len(old))
	for _, c := range old {
		oldMap[c] = true
	}
	for _, c := range nu {
		if !oldMap[c] {
			return false
		}
	}
	return true
}

func statusMapEqual(old, nu map[string]interface{}) bool {
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
			if !conditionsEqual(oldCond, nuCond) {
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
