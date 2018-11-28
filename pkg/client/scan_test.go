package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestNext(t *testing.T) {
	pod := yamlToUnstructured(t, testPod)
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), pod)

	tests := []struct {
		name     string
		state    *ScanState
		expected []interface{}
	}{
		{
			name: "empty list",
			state: &ScanState{
				list: &unstructured.UnstructuredList{},
			},
		},
		{
			name: "end of list",
			state: &ScanState{
				list: &unstructured.UnstructuredList{
					Items: []unstructured.Unstructured{
						unstructured.Unstructured{},
					},
				},
				curRow: 1,
			},
		},
		{
			name: "one item",
			state: &ScanState{
				list: &unstructured.UnstructuredList{
					Items: []unstructured.Unstructured{
						*pod,
					},
				},
				fields: []string{"metadata.name", "metadata.namespace", "{.spec.containers[0].image}"},
			},
			expected: []interface{}{"myapp-pod", "test-namespace", "busybox"},
		},
		{
			name: "retrieve options",
			state: &ScanState{
				client: client.Resource(schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				}),
				fields:    []string{"metadata.name", "metadata.namespace", "{.spec.containers[0].image}"},
				namespace: "test-namespace",
			},
			expected: []interface{}{"myapp-pod", "test-namespace", "busybox"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := assert.New(t)
			vals, err := test.state.Next()
			if a.NoError(err) {
				if test.expected == nil {
					a.Nil(vals)
				} else {
					a.Equal(test.expected, vals)
				}
			}
		})
	}
}
