// This file was automatically generated by lister-gen

package v1

import (
	v1 "github.com/openshift/api/oauth/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// OAuthClientAuthorizationLister helps list OAuthClientAuthorizations.
type OAuthClientAuthorizationLister interface {
	// List lists all OAuthClientAuthorizations in the indexer.
	List(selector labels.Selector) (ret []*v1.OAuthClientAuthorization, err error)
	// Get retrieves the OAuthClientAuthorization from the index for a given name.
	Get(name string) (*v1.OAuthClientAuthorization, error)
	OAuthClientAuthorizationListerExpansion
}

// oAuthClientAuthorizationLister implements the OAuthClientAuthorizationLister interface.
type oAuthClientAuthorizationLister struct {
	indexer cache.Indexer
}

// NewOAuthClientAuthorizationLister returns a new OAuthClientAuthorizationLister.
func NewOAuthClientAuthorizationLister(indexer cache.Indexer) OAuthClientAuthorizationLister {
	return &oAuthClientAuthorizationLister{indexer: indexer}
}

// List lists all OAuthClientAuthorizations in the indexer.
func (s *oAuthClientAuthorizationLister) List(selector labels.Selector) (ret []*v1.OAuthClientAuthorization, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.OAuthClientAuthorization))
	})
	return ret, err
}

// Get retrieves the OAuthClientAuthorization from the index for a given name.
func (s *oAuthClientAuthorizationLister) Get(name string) (*v1.OAuthClientAuthorization, error) {
	key := &v1.OAuthClientAuthorization{ObjectMeta: meta_v1.ObjectMeta{Name: name}}
	obj, exists, err := s.indexer.Get(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("oauthclientauthorization"), name)
	}
	return obj.(*v1.OAuthClientAuthorization), nil
}
