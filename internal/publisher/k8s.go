/*
Copyright (c) 2025 Advanced Micro Devices, Inc. All rights reserved.

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

package publisher

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ROCm/k8s-network-node-labeller/internal/label"
	"k8s.io/client-go/util/retry"
)

// KubernetesPublisher implements Publisher interface for writing labels to Kubernetes nodes
type KubernetesPublisher struct {
	clientset kubernetes.Interface
	nodeName  string
}

// NewKubernetesPublisher creates a new Kubernetes publisher with an existing clientset
func NewKubernetesPublisher(clientset kubernetes.Interface, nodeName string) (*KubernetesPublisher, error) {
	if clientset == nil {
		return nil, fmt.Errorf("clientset cannot be nil")
	}
	if nodeName == "" {
		return nil, fmt.Errorf("node name cannot be empty")
	}

	return &KubernetesPublisher{
		clientset: clientset,
		nodeName:  nodeName,
	}, nil
}

// Publish applies labels to the Kubernetes node
func (kp *KubernetesPublisher) Publish(ctx context.Context, labels []label.Label) error {
	if len(labels) == 0 {
		return nil
	}

	// Get the current node
	var node *v1.Node
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		n, err := kp.clientset.CoreV1().Nodes().Get(ctx, kp.nodeName, metav1.GetOptions{})
		node = n
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", kp.nodeName, err)
	}

	// Update node labels
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	// Add/update labels
	for _, l := range labels {
		node.Labels[l.Key] = l.Value
	}

	// Update the node
	err = retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		_, err := kp.clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to update node %s with labels: %w", kp.nodeName, err)
	}

	return nil
}

// GetName returns the name of the publisher for logging purposes
func (kp *KubernetesPublisher) GetName() string {
	return "KubernetesPublisher"
}
