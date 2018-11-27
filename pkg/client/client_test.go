package client

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

const validKubeconfig = "testdata/kubeconfig"

var (
	testJob = `apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
  backoffLimit: 4
`
	testPod = `apiVersion: v1
kind: Pod
metadata:
  namespace: test-namespace
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox
    command: ['sh', '-c', 'echo Hello Kubernetes! && sleep 3600']
  - name: myapp-container2
    image: postgresql
`
)

func TestGetClientNonexistentFile(t *testing.T) {
	file := "some nonexistent file"
	_, err := GetClient(file, false)
	if err == nil {
		t.Fatalf("expected err, got nil")
	}

	if !os.IsNotExist(errors.Cause(err)) {
		t.Errorf("expected file doesn't exist error, got %v", err)
	}
}

func TestGetClientExistentFile(t *testing.T) {
	kubecfg, err := GetClient(validKubeconfig, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if kubecfg.restcfg.Host != "https://localhost:6443" {
		t.Errorf("expected host to be localhost, was %q", kubecfg.restcfg.Host)
	}
}

func TestGetResourceForKind(t *testing.T) {
	t.Skip("needs a valid kubeconfig")
	a := assert.New(t)

	tests := []struct {
		name     string
		kind     *schema.GroupVersionKind
		resource *schema.GroupVersionResource
	}{
		{
			name: "deployment",
			kind: &schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1beta1",
				Kind:    "Deployment",
			},
			resource: &schema.GroupVersionResource{
				Group:    "extensions",
				Version:  "v1beta1",
				Resource: "deployments",
			},
		},
	}

	client, err := GetClient(validKubeconfig, true)
	if !a.NoError(err) {
		t.FailNow()
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := assert.New(t)
			resource, err := client.GetResourceForKind(test.kind)

			if a.NoError(err) {
				a.Equal(resource.Group, test.resource.Group, "group did not match")
				a.Equal(resource.Version, test.resource.Version, "version did not match")
				a.Equal(resource.Resource, test.resource.Resource, "resource did not match")

			}
		})
	}
}

func yamlToUnstructured(t *testing.T, yaml string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), []byte(yaml), obj); err != nil {
		t.Fatalf("couldn't decode template: %v", err)
	}
	return obj
}

func TestAsRows(t *testing.T) {
	pod := yamlToUnstructured(t, testPod)
	job := yamlToUnstructured(t, testJob)

	tests := []struct {
		name      string
		obj       *unstructured.Unstructured
		rows      []string
		expected  []interface{}
		expectErr bool
	}{
		{
			name:     "basic pod",
			obj:      pod,
			rows:     []string{"metadata.name", "metadata.namespace"},
			expected: []interface{}{"myapp-pod", "test-namespace"},
		},
		{
			name:     "basic job",
			obj:      job,
			rows:     []string{"metadata.name", "spec.backoffLimit"},
			expected: []interface{}{"pi", int64(4)},
		},
		{
			name:      "unknown field",
			obj:       job,
			rows:      []string{"nonsense"},
			expectErr: true,
		},
		{
			name:     "jsonpath",
			obj:      pod,
			rows:     []string{"{.spec.containers[0].image}"},
			expected: []interface{}{"busybox"},
		},
		{
			name:      "malformed jsonpath",
			obj:       pod,
			rows:      []string{"{.spec.container[}"},
			expectErr: true,
		},
		{
			name:      "jsonpath for unknown field",
			obj:       job,
			rows:      []string{"{.spec.somethingelse}"},
			expectErr: true,
		},
		{
			name:     "jsonpath returning list",
			obj:      pod,
			rows:     []string{"{.spec.containers[*].image}"},
			expected: []interface{}{[]interface{}{"busybox", "postgresql"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := assert.New(t)

			result, err := AsRows(test.obj, test.rows)
			if test.expectErr {
				a.Error(err)
				return
			}

			if a.NoError(err) {
				a.Equal(test.expected, result)
			}
		})
	}
}
