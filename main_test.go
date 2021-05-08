package main

import (
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"gomodules.xyz/pointer"
)

var (
	d1 = &v1.Deployment{
		TypeMeta:   metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "demo",
			Name: "d1",
		},
		Spec:       v1.DeploymentSpec{
			Replicas: pointer.Int32P(3),
		},
		Status:     v1.DeploymentStatus{},
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("StatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
