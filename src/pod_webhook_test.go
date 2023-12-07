package main

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPodName(t *testing.T) {
	podFixture := NewPodFixture("test-pod", "default", nil)

	podName := getPodName(podFixture)
	if podName != "test-pod" {
		t.Errorf("Expected pod name to be test-pod, got: %v", podName)
	}
}

func TestDefaultPodMutatorWithoutAnnotation(t *testing.T) {
	podFixture := NewPodFixture("test-pod", "default", nil)

	podMutator := &podMutator{}
	err := podMutator.Default(context.Background(), podFixture)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	cpuLimit, exist := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceCPU]
	if exist {
		t.Errorf("Expected CPU limit to not exist, value: %v", cpuLimit)
	}

	_, exists := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceMemory]
	if !exists {
		t.Errorf("Expected memory limit to exist")
	}
}

func TestDefaultPodMutatorWithAnnotation(t *testing.T) {
	annotations := map[string]string{}

	for _, val := range []string{"true", "false"} {
		annotations[KeepLimitsAnnotation] = val

		podFixture := NewPodFixture("test-pod", "default", annotations)

		podMutator := &podMutator{}
		err := podMutator.Default(context.Background(), podFixture)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		cpuLimit, exist := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceCPU]
		if exist && val == "false" {
			t.Errorf("Expected CPU limit to not exist, value: %v", cpuLimit)
		} else if !exist && val == "true" {
			t.Errorf("Expected CPU limit to exist")
		}

		_, exists := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceMemory]
		if !exists {
			t.Errorf("Expected memory limit to exist")
		}
	}
}

func TestRemoveContainerLimits(t *testing.T) {
	podFixture := NewPodFixture("test-pod", "default", nil)

	removeContainerLimits(&podFixture.Spec.Containers[0], v1.ResourceCPU, podFixture)

	cpuLimit, exist := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceCPU]
	if exist {
		t.Errorf("Expected CPU limit to not exist, value: %v", cpuLimit)
	}

	_, exists := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceMemory]
	if !exists {
		t.Errorf("Expected memory limit to exist")
	}
}

func NewPodFixture(name string, namespace string, annotations map[string]string) *v1.Pod {
	if annotations == nil {
		annotations = map[string]string{}
	}

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "myapp",
			},
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container-1",
					Image: "nginx:latest",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("1"),
							v1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}
}
