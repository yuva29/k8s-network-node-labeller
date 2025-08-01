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

package label

import (
	"testing"
)

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uppercase to lowercase",
			input:    "UPPERCASE",
			expected: "uppercase",
		},
		{
			name:     "spaces to dashes",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "underscores to dashes",
			input:    "hello_world",
			expected: "hello-world",
		},
		{
			name:     "mixed spaces and underscores",
			input:    "hello world_test string",
			expected: "hello-world-test-string",
		},
		{
			name:     "remove special characters",
			input:    "hello@world#test$string%",
			expected: "helloworldteststring",
		},
		{
			name:     "preserve dots",
			input:    "hello.world.test",
			expected: "hello.world.test",
		},
		{
			name:     "preserve dashes",
			input:    "hello-world-test",
			expected: "hello-world-test",
		},
		{
			name:     "preserve alphanumeric",
			input:    "hello123world456",
			expected: "hello123world456",
		},
		{
			name:     "complex string with all transformations",
			input:    "Hello_World Test@123.Example#String$",
			expected: "hello-world-test123.examplestring",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with only special characters",
			input:    "@#$%^&*()",
			expected: "",
		},
		{
			name:     "string with only spaces and underscores",
			input:    "_ _ _ _",
			expected: "-------",
		},
		{
			name:     "numbers only",
			input:    "123456",
			expected: "123456",
		},
		{
			name:     "dots only",
			input:    "...",
			expected: "...",
		},
		{
			name:     "dashes only",
			input:    "---",
			expected: "---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeString(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
