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

package nodelabeller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ROCm/k8s-network-node-labeller/internal/cleaner"
	"github.com/ROCm/k8s-network-node-labeller/internal/discoverer"
	"github.com/ROCm/k8s-network-node-labeller/internal/label"
	"github.com/ROCm/k8s-network-node-labeller/internal/publisher"
)

const (
	// DefaultInterval is the default interval for labelling cycles
	DefaultInterval = time.Minute
)

// NodeLabeller coordinates the labelling process
type NodeLabeller struct {
	cleaners    *[]cleaner.Cleaner
	discoverers *[]discoverer.Discoverer
	publishers  *[]publisher.Publisher
}

// NewNodeLabeller creates a new NodeLabeller
func NewNodeLabeller(
	cleaners *[]cleaner.Cleaner,
	discoverers *[]discoverer.Discoverer,
	publishers *[]publisher.Publisher,
) *NodeLabeller {
	return &NodeLabeller{
		cleaners:    cleaners,
		discoverers: discoverers,
		publishers:  publishers,
	}
}

// Run starts the periodic labelling process
func (nl *NodeLabeller) Run(ctx context.Context) error {
	log.Printf("Starting NodeLabeller with interval %v", DefaultInterval)

	ticker := time.NewTicker(DefaultInterval)
	defer ticker.Stop()

	// Run once immediately, exit if there's an error in the initial run
	if err := nl.runOnce(ctx); err != nil {
		log.Printf("Error during initial labelling run: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("NodeLabeller stopping due to context cancellation")
			return ctx.Err()
		case <-ticker.C:
			if err := nl.runOnce(ctx); err != nil {
				log.Printf("Error during labelling run: %v", err)
				// Continue running even if there's an error
			}
		}
	}
}

// runOnce performs a single labelling cycle
func (nl *NodeLabeller) runOnce(ctx context.Context) error {

	// Step 1: Clean existing labels that match managed label patterns
	regexps := label.DefaultNICPrefixedLabelRegexps
	log.Printf("Cleaning labels matching managed label regexps")
	for _, cleaner := range *nl.cleaners {
		if err := cleaner.CleanByRegexps(ctx, regexps); err != nil {
			return fmt.Errorf("failed to clean labels by regexps: %w", err)
		}
	}
	log.Printf("Successfully cleaned labels matching managed label regexps")

	// Step 2: Discover new labels
	var allLabels []label.Label
	for i, disc := range *nl.discoverers {
		log.Printf("Running discoverer %d...", i+1)
		labels, err := disc.Discover()
		if err != nil {
			return fmt.Errorf("error discovering from discoverer %d: %w", i+1, err)
		}
		log.Printf("Discoverer %s[%d] found %d labels", disc.GetName(), i+1, len(labels))
		allLabels = append(allLabels, labels...)
	}

	log.Printf("Generated %d labels total", len(allLabels))

	// Print discovered labels for debugging
	printDiscoveredLabels(allLabels)

	// Step 3: Publish labels
	if len(allLabels) == 0 {
		log.Println("No labels to publish")
		return nil
	}

	for i, pub := range *nl.publishers {
		log.Printf("Publishing %d labels with publisher %s", len(allLabels), pub.GetName())
		if err := pub.Publish(ctx, allLabels); err != nil {
			return fmt.Errorf("failed to publish labels with publisher %d: %w", i+1, err)
		}
		log.Printf("Successfully published %d labels with publisher %s", len(allLabels), pub.GetName())
	}

	log.Printf("Labelling cycle completed")
	return nil
}

// Stop gracefully stops the NodeLabeller
func (nl *NodeLabeller) Stop() {
	log.Printf("Cleaning up managed labels when stopping NodeLabeller")
	regexps := label.DefaultNICPrefixedLabelRegexps
	for _, cleaner := range *nl.cleaners {
		if err := cleaner.CleanByRegexps(context.TODO(), regexps); err != nil {
			log.Printf("failed to clean labels by regexps: %v", err)
		}
	}
	log.Printf("Successfully cleaned up managed labels")
	log.Println("NodeLabeller stopped")
}

// printDiscoveredLabels logs the discovered labels
func printDiscoveredLabels(labels []label.Label) {
	if len(labels) == 0 {
		log.Println("No labels discovered")
		return
	}

	log.Println("Discovered labels:")
	for _, l := range labels {
		log.Printf(" - %s: %s", l.Key, l.Value)
	}
}
