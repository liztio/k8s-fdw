package client

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/third_party/forked/golang/template"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/jsonpath"
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

// GetTableScanner returns a scanner appropriate for the given columns and table options.
// The scanner will not retrieve results until the first time Next is called
func (c *Client) GetTableScanner(columns []string, tableOpts map[string]string) (*ScanState, error) {
	apiVersion, ok := tableOpts["apiVersion"]
	if !ok {
		return nil, errors.New("apiVersion is mandatory")
	}
	kind, ok := tableOpts["kind"]
	if !ok {
		return nil, errors.New("kind is mandatory")
	}

	groupVersion, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse group version %q", apiVersion)
	}

	gvk := groupVersion.WithKind(kind)
	gvr, err := c.GetResourceForKind(&gvk)
	if err != nil {
		return nil, errors.Wrapf(err, "couldnt get resource for kind %v", gvk.String())
	}

	return c.makeTableScanner(gvr, columns, tableOpts)
}

func (c *Client) makeTableScanner(gvr *schema.GroupVersionResource, columns []string, tableOpts map[string]string) (*ScanState, error) {
	scanColumns := make([]string, len(columns))
	for i, col := range columns {
		alias, ok := tableOpts["@"+col]
		if ok {
			scanColumns[i] = alias
			continue
		}
		scanColumns[i] = col
	}

	return &ScanState{
		client:    c.dyn.Resource(*gvr),
		namespace: tableOpts["namespace"],
		fields:    scanColumns,
	}, nil
}

// GetResourceForKind turns a GroupVersionKind into a GroupVersionResource
func (c *Client) GetResourceForKind(gvk *schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get resource for %+v", gvk)
	}
	return &mapping.Resource, nil
}

// AsRows takes an object and turns it into a single k8s row.
func AsRows(obj *unstructured.Unstructured, columns []string) ([]interface{}, error) {
	row := make([]interface{}, len(columns))

	for i, col := range columns {
		// Is it JSONPath?
		if strings.HasPrefix(col, "{") && strings.HasSuffix(col, "}") {
			jp := jsonpath.New(col)
			// TODO(EKF): Cache these parse results
			if err := jp.Parse(col); err != nil {
				return []interface{}{}, errors.Wrapf(err, "couldn't parse template %q", col)
			}
			results, err := jp.FindResults(obj.Object)
			if err != nil {
				return []interface{}{}, errors.Wrapf(err, "failed to execute template %q", col)
			}

			if len(results) == 0 {
				continue
			}

			// extract results from the rat's nest of reflect.Value
			vals := make([]interface{}, len(results[0]))
			for i, resVal := range results[0] {
				val, ok := template.PrintableValue(resVal)
				if !ok {
					return []interface{}{}, fmt.Errorf("couldn't print result %v", results)
				}
				vals[i] = val
			}

			if len(vals) == 1 {
				row[i] = vals[0]
			} else {
				row[i] = vals
			}

			continue
		}

		val, found, err := unstructured.NestedFieldCopy(obj.Object, strings.Split(col, ".")...)
		if err != nil {
			return []interface{}{}, errors.Wrap(err, "couldn't traverse object")
		}
		if !found {
			return []interface{}{}, fmt.Errorf("no such field %q", col)
		}
		row[i] = val
	}

	return row, nil
}
