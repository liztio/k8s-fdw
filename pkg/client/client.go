package client

import (
	// include auth methods
	"io/ioutil"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents
type Client struct {
	restcfg *rest.Config
	dyn     dynamic.Interface
	mapper  meta.RESTMapper
}

// GetClient returns a client for the given kubectl path
// TODO: getMapper is a dirty hack to make tests work
func GetClient(path string, getMapper bool) (*Client, error) {
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

	client := &Client{
		restcfg: restcfg,
		dyn:     dyn,
	}

	if getMapper {
		mapper, err := newAPIHelperFromRESTConfig(restcfg)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't get resource mapper")
		}
		client.mapper = mapper
	}

	return client, nil
}

func newAPIHelperFromRESTConfig(cfg *rest.Config) (meta.RESTMapper, error) {
	discover, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not create discovery client")
	}
	groupResources, err := restmapper.GetAPIGroupResources(discover)
	if err != nil {
		return nil, errors.Wrap(err, "could not get api group resources")
	}
	return restmapper.NewDiscoveryRESTMapper(groupResources), nil
}

func (c *Client) GetResourceForKind(gvk *schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get resource for %+v", gvk)
	}
	return &mapping.Resource, nil
}
