package api

const (
	TlsAcmeAnnotation                   = "kubernetes.io/tls-acme"
	TlsAcmePausedAnnotation             = "kubernetes.io/tls-acme-paused"
	TlsAcmeLabel                        = TlsAcmeAnnotation
	AcmeAwaitingAuthzUrlAnnotation      = "kubernetes.io/tls-acme-awaiting-authorization-at-url"
	AcmeAwaitingAuthzUrlOwnerAnnotation = "kubernetes.io/tls-acme-awaiting-authorization-owner"
)
