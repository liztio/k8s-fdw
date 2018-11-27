package client

import (
	// include auth methods
	"io/ioutil"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents
type Client struct {
	restcfg *rest.Config
	dyn     dynamic.Interface
}

// GetClient returns a client for the given kubectl path
func GetClient(path string) (*Client, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read file")
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse kubeconfig %s", path)
	}
	dyn, err := dynamic.NewForConfig(restcfg)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create dynamic client from %s", path)
	}

	return &Client{
		restcfg: restcfg,
		dyn:     dyn,
	}, nil
}
