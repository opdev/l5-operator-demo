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
	"reflect"
	"testing"

	v1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
)

func TestLabels(t *testing.T) {
	type args struct {
		bestie v1.Bestie
		labels []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Labels(tt.args.bestie, tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Labels() = %v, want %v", got, tt.want)
			}
		})
	}
}
