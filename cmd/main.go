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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ROCm/k8s-network-node-labeller/internal/cleaner"
	"github.com/ROCm/k8s-network-node-labeller/internal/discoverer"
	"github.com/ROCm/k8s-network-node-labeller/internal/nicctl"
	"github.com/ROCm/k8s-network-node-labeller/internal/nodelabeller"
	"github.com/ROCm/k8s-network-node-labeller/internal/publisher"
	"github.com/ROCm/k8s-network-node-labeller/internal/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func createNodeLabeller() (*nodelabeller.NodeLabeller, error) {
	nodeName, err := utils.GetNodeNameFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get node name: %w", err)
	}
	log.Printf("Node name: %s", nodeName)

	// Create shared Kubernetes client
	clientset, err := createKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create nicctl client
	nicctlClient, err := createNicctlClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create nicctl client: %w", err)
	}

	// Create cleaner with shared kubernetes client
	kubeCleaner, err := cleaner.NewKubernetesCleaner(clientset, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes cleaner: %w", err)
	}

	// Create publisher with shared kubernetes client
	kubePublisher, err := publisher.NewKubernetesPublisher(clientset, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes publisher: %w", err)
	}

	// Create discoverer with the nicctl client
	nicctlDiscoverer := discoverer.NewNicctlDiscoverer(nicctlClient)

	discoverers := []discoverer.Discoverer{
		nicctlDiscoverer,
	}
	publishers := []publisher.Publisher{
		kubePublisher,
	}
	cleaners := []cleaner.Cleaner{
		kubeCleaner,
	}

	return nodelabeller.NewNodeLabeller(
		&cleaners,
		&discoverers,
		&publishers,
	), nil
}

// createKubernetesClient creates a Kubernetes client
func createKubernetesClient() (kubernetes.Interface, error) {
	// Try to create in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return clientset, nil
}

// createNicctlClient creates a nicctl client based on the environment
func createNicctlClient() (nicctl.NicctlClient, error) {
	// If SIM_ENABLE environment variable is set, create a mock client
	if _, present := os.LookupEnv("SIM_ENABLE"); present {
		log.Println("SIM_ENABLE is set, using MockNicctlClient")
		mockClient := nicctl.NewMockNicctlClient()
		mockClient.GetIonicDriverVersionFunc = func() (string, error) {
			return "", nil
		}
		return mockClient, nil
	}

	nicctlClient, err := nicctl.NewNicctlCommandClient(discoverer.NicctlBinaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create nicctl client: %w", err)
	}
	return nicctlClient, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start NodeLabeller
	labeller, err := createNodeLabeller()
	if err != nil {
		log.Fatalf("Failed to create NodeLabeller: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the labeller in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := labeller.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
		labeller.Stop()
	case err := <-errChan:
		log.Printf("NodeLabeller error: %v", err)
		cancel()
		labeller.Stop()
	}

	log.Println("NodeLabeller shutdown complete")
}
