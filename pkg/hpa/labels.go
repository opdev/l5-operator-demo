/*
Copyright The L5 Operator Authors

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

package hpa

import (
	"regexp"
	"strings"

	v1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"

	truncate "github.com/aquilax/truncate"
)

func filterLabel(label string, filterLabels []string) bool {
	for _, pattern := range filterLabels {
		match, _ := regexp.MatchString(pattern, label)
		return match
	}

	return false
}

func Labels(bestie v1.Bestie, labels []string) map[string]string {
	labelSet := map[string]string{}
	if nil != bestie.Labels {
		for k, v := range bestie.Labels {
			if !filterLabel(k, labels) {
				labelSet[k] = v
			}
		}
	}
	instance := bestie.Namespace + "." + bestie.Name
	labelSet["app.kubernetes.io/managed-by"] = "L5-Operator"
	labelSet["app.kubernetes.io/instance"] = truncate.Truncate(instance, 63, ".", truncate.PositionEnd)
	labelSet["app.kubernetes.io/part-of"] = "L5"
	labelSet["app.kubernetes.io/component"] = "L5 Deployment"
	version := strings.Split(bestie.Spec.Image, ":")
	if len(version) > 1 {
		labelSet["app.kubernetes.io/version"] = version[len(version)-1]
	} else {
		labelSet["app.kubernetes.io/version"] = "latest"
	}

	return labelSet
}
