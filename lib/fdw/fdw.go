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
	return TableStats{Rows: uint(1), StartCost: 10, TotalCost: 1000}, nil
}

func (t *k8sTable) Scan(rel *Relation, opts *Options) (Iterator, error) {
	if t.k8sClient == nil {
		kubeconfig, ok := opts.ServerOptions["kubeconfig"]
		if !ok {
			return nil, errors.New("server option kubeconfig is mandatory")
		}

		client, err := client.GetClient(kubeconfig, true)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get kubeconfig")
		}
		t.k8sClient = client
	}

	scanState, err := t.k8sClient.GetTableScanner(getRelColumns(rel), opts.TableOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get scanner")
	}

	return &listIter{
		scanState: scanState,
	}, nil
}

func (it *k8sTable) Explain(e Explainer) {
	e.Property("Powered by", "Go FDW")
}

type listIter struct {
	sync.Mutex
	scanState *client.ScanState
}

var _ Explainable = (*k8sTable)(nil)

func (it *listIter) Next() ([]interface{}, error) {
	return it.scanState.Next()
}

func (it *listIter) Reset() {
	it.Lock()
	defer it.Unlock()
	it.scanState.Reset()
}
func (it *listIter) Close() error { return nil }

func getRelColumns(rel *Relation) []client.Column {
	cols := make([]client.Column, len(rel.Attr.Attrs))
	for i, attr := range rel.Attr.Attrs {
		cols[i].Name = attr.Name
		cols[i].Options = attr.Options
	}

	return cols
}
