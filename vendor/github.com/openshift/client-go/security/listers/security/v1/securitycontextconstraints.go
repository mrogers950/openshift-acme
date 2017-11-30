// This file was automatically generated by lister-gen

package v1

import (
	v1 "github.com/openshift/api/security/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SecurityContextConstraintsLister helps list SecurityContextConstraintses.
type SecurityContextConstraintsLister interface {
	// List lists all SecurityContextConstraintses in the indexer.
	List(selector labels.Selector) (ret []*v1.SecurityContextConstraints, err error)
	// Get retrieves the SecurityContextConstraints from the index for a given name.
	Get(name string) (*v1.SecurityContextConstraints, error)
	SecurityContextConstraintsListerExpansion
}

// securityContextConstraintsLister implements the SecurityContextConstraintsLister interface.
type securityContextConstraintsLister struct {
	indexer cache.Indexer
}

// NewSecurityContextConstraintsLister returns a new SecurityContextConstraintsLister.
func NewSecurityContextConstraintsLister(indexer cache.Indexer) SecurityContextConstraintsLister {
	return &securityContextConstraintsLister{indexer: indexer}
}

// List lists all SecurityContextConstraintses in the indexer.
func (s *securityContextConstraintsLister) List(selector labels.Selector) (ret []*v1.SecurityContextConstraints, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.SecurityContextConstraints))
	})
	return ret, err
}

// Get retrieves the SecurityContextConstraints from the index for a given name.
func (s *securityContextConstraintsLister) Get(name string) (*v1.SecurityContextConstraints, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("securitycontextconstraints"), name)
	}
	return obj.(*v1.SecurityContextConstraints), nil
}
