/*
Copyright 2023 bjw-s.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type podMutator struct{}

func (m *podMutator) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.Log
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected a Pod but got a %T", obj)
	}

	if val, ok := pod.Annotations[KeepLimitsAnnotation]; ok {
		if val == "true" {
			log.Info("Skipping Pod because annotation is explicitly true", "pod", getPodName(pod))
			return nil
		}
	}

	for _, container := range pod.Spec.InitContainers {
		removeContainerLimits(&container, corev1.ResourceCPU, pod)
	}

	for _, container := range pod.Spec.Containers {
		removeContainerLimits(&container, corev1.ResourceCPU, pod)
	}
	return nil
}

func removeContainerLimits(container *corev1.Container, limitType corev1.ResourceName, pod *corev1.Pod) {
	log := logf.Log
	limits := container.Resources.Limits
	_, cpuLimitExists := limits[limitType]
	if cpuLimitExists {
		delete(limits, limitType)
		log.Info("Removed resource limit",
			"namespace", pod.Namespace,
			"pod", getPodName(pod),
			"container", container.Name,
			"limit", limitType,
		)
	}
}

func getPodName(pod *corev1.Pod) string {
	podName := pod.GetName()
	if podName != "" {
		return podName
	}
	return pod.GetGenerateName()
}
