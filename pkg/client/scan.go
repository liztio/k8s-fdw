package client

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// ScanState represents the state of one particular scan
type ScanState struct {
	client    dynamic.NamespaceableResourceInterface
	namespace string
	list      *unstructured.UnstructuredList
	curRow    int
	fields    []string
}

// Next returns the next value, or an empty list if there's no more left
func (s *ScanState) Next() ([]interface{}, error) {
	if s.list == nil {
		if err := s.retrieve(); err != nil {
			return []interface{}{}, err
		}
	}

	// TODO(EKF): handle continue
	if len(s.list.Items) <= s.curRow {
		return []interface{}{}, nil
	}

	row, err := AsRows(&s.list.Items[s.curRow], s.fields)
	if err != nil {
		return []interface{}{}, errors.Wrap(err, "couldn't construct row")
	}

	s.curRow++
	return row, nil
}

func (s *ScanState) retrieve() error {
	var err error
	if s.namespace == "" {
		s.list, err = s.client.List(metav1.ListOptions{})
	} else {
		s.list, err = s.client.Namespace(s.namespace).List(metav1.ListOptions{})
	}
	return errors.Wrap(err, "couldn't retrieve objects")
}

func (s *ScanState) Reset() {
	s.curRow = 0
}
