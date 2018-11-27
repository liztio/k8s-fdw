package client

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const validKubeconfig = "testdata/kubeconfig"

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
			resource, err := client.GetResourceForKind(test.kind)

			if a.NoError(err) {
				a.Equal(resource.Group, test.resource.Group, "group did not match")
				a.Equal(resource.Version, test.resource.Version, "version did not match")
				a.Equal(resource.Resource, test.resource.Resource, "resource did not match")

			}
		})
	}
}
