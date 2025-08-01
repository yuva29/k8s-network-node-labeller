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
	"regexp"
	"strings"
)

// NormalizeString normalizes a string by converting to be able to use it as a label key or value.
func NormalizeString(s string) string {
	// Convert to lower case
	s = strings.ToLower(s)
	// Replace spaces and underscores with dashes
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove any non-alphanumeric characters except dashes and dots
	re := regexp.MustCompile(`[^a-z0-9-.]`)
	s = re.ReplaceAllString(s, "")
	return s
}
