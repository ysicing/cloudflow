/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/ysicing/cloudflow/apis/jobs/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// GenTLSLister helps list GenTLSs.
// All objects returned here must be treated as read-only.
type GenTLSLister interface {
	// List lists all GenTLSs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.GenTLS, err error)
	// GenTLSs returns an object that can list and get GenTLSs.
	GenTLSs(namespace string) GenTLSNamespaceLister
	GenTLSListerExpansion
}

// genTLSLister implements the GenTLSLister interface.
type genTLSLister struct {
	indexer cache.Indexer
}

// NewGenTLSLister returns a new GenTLSLister.
func NewGenTLSLister(indexer cache.Indexer) GenTLSLister {
	return &genTLSLister{indexer: indexer}
}

// List lists all GenTLSs in the indexer.
func (s *genTLSLister) List(selector labels.Selector) (ret []*v1beta1.GenTLS, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.GenTLS))
	})
	return ret, err
}

// GenTLSs returns an object that can list and get GenTLSs.
func (s *genTLSLister) GenTLSs(namespace string) GenTLSNamespaceLister {
	return genTLSNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// GenTLSNamespaceLister helps list and get GenTLSs.
// All objects returned here must be treated as read-only.
type GenTLSNamespaceLister interface {
	// List lists all GenTLSs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.GenTLS, err error)
	// Get retrieves the GenTLS from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.GenTLS, error)
	GenTLSNamespaceListerExpansion
}

// genTLSNamespaceLister implements the GenTLSNamespaceLister
// interface.
type genTLSNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all GenTLSs in the indexer for a given namespace.
func (s genTLSNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.GenTLS, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.GenTLS))
	})
	return ret, err
}

// Get retrieves the GenTLS from the indexer for a given namespace and name.
func (s genTLSNamespaceLister) Get(name string) (*v1beta1.GenTLS, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("gentls"), name)
	}
	return obj.(*v1beta1.GenTLS), nil
}
