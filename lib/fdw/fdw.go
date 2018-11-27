package main

import (
	"sync"

	"github.com/liztio/k8s-fdw/pkg/client"
	"github.com/pkg/errors"
)

// TODO: This file should be replaced by your actual implementation. Call SetTable in init to start serving queries.

func init() {
	SetTable(&k8sTable{})
}

type k8sTable struct {
	k8sClient *client.Client
}

func (t *k8sTable) Stats(opts *Options) (TableStats, error) {
	return TableStats{Rows: uint(t.Rows), StartCost: 10, TotalCost: 1000}, nil
}

func (t *k8sTable) Scan(rel *Relation, opts *Options) (Iterator, error) {
	if t.k8sClient == nil {
		kubeconfig, ok := opts.ServerOptions["kubeconfig"]
		if !ok {
			return nil, errors.New("server option kubeconfig is mandatory")
		}

		client, err := GetClient(kubeconfig, true)
		if err != nil {
			return errors.Wrap(err, "couldn't get kubeconfig")
		}
		k8sTable.k8sClient = client
	}

	scanState, err := t.k8sClient.GetTableScanner(rel.Attrs, opts.TableOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get scanner")
	}

	return &listIter{
		scanState: scanState,
	}
}

type listIter struct {
	sync.Mutex
	scanState *client.ScanState
}

var _ Explainable = (*k8sTable)(nil)

func (it *listIter) Explain(e Explainer) {
	e.Property("Powered by", "Go FDW")
}

func (it *listIter) Next() ([]interface{}, error) {
	return it.scanState.Next()
}

func (it *listIter) Reset() {
	it.Lock()
	defer it.Unlock()
	it.scanState.Reset()
}
func (it *listIter) Close() error { return nil }
