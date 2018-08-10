/*
Copyright 2017 Google Inc.
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

package retrieval

import (
	"reflect"
	"testing"

	"github.com/prometheus/prometheus/pkg/labels"
)

func TestTranslate(t *testing.T) {
	r := ResourceMap{
		Type: "my_type",
		LabelMap: map[string]labelTranslation{
			"__target1": constValue("sdt1"),
			"__target2": constValue("sdt2"),
		},
	}
	// This target is missing label "__target1".
	noMatchTarget := labels.Labels{
		{"ignored", "x"},
		{"__target2", "y"},
	}
	matchTarget := labels.Labels{
		{"ignored", "x"},
		{"__target2", "y"},
		{"__target1", "z"},
	}
	if labels := r.Translate(noMatchTarget); labels != nil {
		t.Errorf("Expected no match, matched %v", labels)
	}

	expectedLabels := map[string]string{
		"sdt1": "z",
		"sdt2": "y",
	}
	if labels := r.Translate(matchTarget); labels == nil {
		t.Errorf("Expected %v, actual nil", expectedLabels)
	} else if !reflect.DeepEqual(labels, expectedLabels) {
		t.Errorf("Expected %v, actual %v", expectedLabels, labels)
	}
}

func TestTranslateGceInstance(t *testing.T) {
	target := labels.Labels{
		{"__meta_gce_project", "my-project"},
		{"__meta_gce_zone", "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a"},
		{"__meta_gce_instance_id", "1234110975759588"},
	}
	expectedLabels := map[string]string{
		"project_id":  "my-project",
		"zone":        "us-central1-a",
		"instance_id": "1234110975759588",
	}
	if labels := GCEResourceMap.Translate(target); labels == nil {
		t.Errorf("Expected %v, actual nil", expectedLabels)
	} else if !reflect.DeepEqual(labels, expectedLabels) {
		t.Errorf("Expected %v, actual %v", expectedLabels, labels)
	}
}

func BenchmarkTranslate(b *testing.B) {
	r := ResourceMap{
		Type: "gke_container",
		LabelMap: map[string]labelTranslation{
			ProjectIDLabel:                   constValue("project_id"),
			KubernetesLocationLabel:          constValue("zone"),
			KubernetesClusterNameLabel:       constValue("cluster_name"),
			"_kubernetes_namespace":          constValue("namespace_id"),
			"_kubernetes_pod_name":           constValue("pod_id"),
			"_kubernetes_pod_node_name":      constValue("instance_id"),
			"_kubernetes_pod_container_name": constValue("container_name"),
		},
	}
	targetLabels := labels.Labels{
		{ProjectIDLabel, "1:anoeuh oeusoeh uasoeuh"},
		{KubernetesLocationLabel, "2:anoeuh oeusoeh uasoeuh"},
		{KubernetesClusterNameLabel, "3:anoeuh oeusoeh uasoeuh"},
		{"_kubernetes_namespace", "4:anoeuh oeusoeh uasoeuh"},
		{"_kubernetes_pod_name", "5:anoeuh oeusoeh uasoeuh"},
		{"_kubernetes_pod_node_name", "6:anoeuh oeusoeh uasoeuh"},
		{"_kubernetes_pod_container_name", "7:anoeuh oeusoeh uasoeuh"},
		{"ignored", "8:anoeuh oeusoeh uasoeuh"},
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if labels := r.Translate(targetLabels); labels == nil {
			b.Fail()
		}
	}
}
