package main

import (
	"testing"
	"time"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var (
	a1 = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:03:45Z'
    lastTransitionTime: '2021-05-08T19:03:45Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'True'
    lastUpdateTime: '2021-05-08T19:03:45Z'
    lastTransitionTime: '2021-05-08T19:03:45Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
	a1MissingCondition = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
`
	a1ConditionTimeUpdated = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:11:21Z'
    lastTransitionTime: '2021-05-08T19:11:21Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'True'
    lastUpdateTime: '2021-05-08T19:11:21Z'
    lastTransitionTime: '2021-05-08T19:11:21Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
	a1ConditionStatusUpdated = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:12:51Z'
    lastTransitionTime: '2021-05-08T19:12:51Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'False'
    lastUpdateTime: '2021-05-08T19:12:51Z'
    lastTransitionTime: '2021-05-08T19:12:51Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
)

var (
	d1 = &apps.Deployment{
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
					LastUpdateTime:     metav1.NewTime(time.Now()),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability.",
				},
				{
					Type:               "Progressing",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now()),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet \"d1\" has successfully progressed.",
				},
			},
		},
	}
	d1MissingCondition = &apps.Deployment{
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
			Conditions:          nil,
		},
	}
	d1ConditionTimeUpdated = &apps.Deployment{
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
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
					LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet \"d1\" has successfully progressed.",
				},
			},
		},
	}
	d1ConditionStatusUpdated = &apps.Deployment{
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
)

func TestStatusEqual(t *testing.T) {
	type args struct {
		old interface{}
		new interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Map Same",
			args: args{
				old: toJSON(a1),
				new: toJSON(a1),
			},
			want: true,
		},
		{
			name: "Map Missing Conditions",
			args: args{
				old: toJSON(a1MissingCondition),
				new: toJSON(a1MissingCondition),
			},
			want: true,
		},
		{
			name: "Map Condition Time Modified",
			args: args{
				old: toJSON(a1),
				new: toJSON(a1ConditionTimeUpdated),
			},
			want: true,
		},
		{
			name: "Map Condition Status Modified",
			args: args{
				old: toJSON(a1),
				new: toJSON(a1ConditionStatusUpdated),
			},
			want: false,
		},
		{
			name: "Struct Same",
			args: args{
				old: d1,
				new: d1,
			},
			want: true,
		},
		{
			name: "Struct Missing Conditions",
			args: args{
				old: d1MissingCondition,
				new: d1MissingCondition,
			},
			want: true,
		},
		{
			name: "Struct Condition Time Modified",
			args: args{
				old: d1,
				new: d1ConditionTimeUpdated,
			},
			want: true,
		},
		{
			name: "Struct Condition Status Modified",
			args: args{
				old: d1,
				new: d1ConditionStatusUpdated,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("StatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toJSON(s string) runtime.Object {
	data, err := yaml.YAMLToJSON([]byte(s))
	if err != nil {
		panic(err)
	}
	out, _, err := unstructured.UnstructuredJSONScheme.Decode(data, nil, nil)
	if err != nil {
		panic(err)
	}
	return out
}
