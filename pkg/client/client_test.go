package client

import (
	"os"
	"testing"

	"github.com/pkg/errors"
)

const validKubeconfig = "testdata/kubeconfig"

func TestGetClientNonexistentFile(t *testing.T) {
	file := "some nonexistent file"
	_, err := GetClient(file)
	if err == nil {
		t.Fatalf("expected err, got nil")
	}

	if !os.IsNotExist(errors.Cause(err)) {
		t.Errorf("expected file doesn't exist error, got %v", err)
	}
}

func TestGetClientExistentFile(t *testing.T) {
	kubecfg, err := GetClient(validKubeconfig)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if kubecfg.restcfg.Host != "https://localhost:6443" {
		t.Errorf("expected host to be localhost, was %q", kubecfg.restcfg.Host)
	}
}
