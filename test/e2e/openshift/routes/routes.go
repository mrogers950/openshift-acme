package routes

import (
	"context"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/tnozicka/openshift-acme/pkg/acme/client/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/tnozicka/openshift-acme/pkg/util"
	"github.com/tnozicka/openshift-acme/test/e2e/framework"
	exutil "github.com/tnozicka/openshift-acme/test/e2e/openshift/util"
)

const (
	RouteAdmissionTimeout          = 5 * time.Second
	CertificateProvisioningTimeout = 60 * time.Second
)

func DeleteACMEAccountIfRequested(f *framework.Framework) error {
	namespace := exutil.DeleteAccountBetweenStepsInNamespace()
	if namespace == "" {
		return nil
	}
	name := "acme-account"

	// We need to deactivate account first because controller uses informer and might have it cached
	secret, err := f.KubeAdminClientSet().CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	client, err := builder.BuildClientFromSecret(secret)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.DeactivateAccount(ctx, client.Account)

	var grace int64 = 0
	propagation := metav1.DeletePropagationForeground
	framework.Logf("Deleting account Secret %s/%s", namespace, name)
	err = f.KubeAdminClientSet().CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy:  &propagation,
		GracePeriodSeconds: &grace,
	})
	if err != nil {
		return err
	}

	return nil
}

var _ = g.Describe("Routes", func() {
	defer g.GinkgoRecover()
	f := framework.NewFramework("routes")

	g.It("should provision certificates for annotated Routes", func() {
		namespace := f.Namespace()

		g.By("creating new Route without TLS")
		route := &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					"kubernetes.io/tls-acme": "true",
				},
			},
			Spec: routev1.RouteSpec{
				Host: exutil.Domain(),
				To: routev1.RouteTargetReference{
					Name: "non-existing",
				},
			},
		}
		route, err := f.RouteClientset().RouteV1().Routes(namespace).Create(route)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for Route to be admitted by the router")
		w, err := f.RouteClientset().RouteV1().Routes(namespace).Watch(metav1.SingleObject(route.ObjectMeta))
		o.Expect(err).NotTo(o.HaveOccurred())
		event, err := watch.Until(RouteAdmissionTimeout, w, util.RouteAdmittedFunc())
		o.Expect(err).NotTo(o.HaveOccurred(), "Failed to wait for Route to be admitted by the router!")

		g.By("waiting for initial certificate to be be provisioned")
		w, err = f.RouteClientset().RouteV1().Routes(namespace).Watch(metav1.SingleObject(route.ObjectMeta))
		o.Expect(err).NotTo(o.HaveOccurred())
		event, err = watch.Until(CertificateProvisioningTimeout, w, util.RouteTLSChangedFunc(route.Spec.TLS))
		o.Expect(err).NotTo(o.HaveOccurred(), "Failed to wait for certificate to be provisioned!")

		route = event.Object.(*routev1.Route)
		o.Expect(route.Spec.TLS).NotTo(o.BeNil())

		o.Expect(route.Spec.TLS.Termination).To(o.Equal(routev1.TLSTerminationEdge))

		crt, err := util.CertificateFromPEM([]byte(route.Spec.TLS.Certificate))
		o.Expect(err).NotTo(o.HaveOccurred())

		now := time.Now()
		o.Expect(now.Before(crt.NotBefore)).To(o.BeFalse())
		o.Expect(now.After(crt.NotAfter)).To(o.BeFalse())

		// ACME server will likely cache the validation for our domain and won't retry it so soon.
		err = DeleteACMEAccountIfRequested(f)
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("deleting the initial certificate and waiting for new one to be provisioned")
		routeCopy := route.DeepCopy()
		routeCopy.Spec.TLS = nil
		route, err = f.RouteClientset().RouteV1().Routes(namespace).Patch(route.Name, types.StrategicMergePatchType, []byte(`{"spec":{"tls":{"certificate":"","key":""}}}`))
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(route.Spec.TLS).NotTo(o.BeNil())
		o.Expect(route.Spec.TLS.Certificate).To(o.BeEmpty())
		o.Expect(route.Spec.TLS.Key).To(o.BeEmpty())

		w, err = f.RouteClientset().RouteV1().Routes(namespace).Watch(metav1.SingleObject(route.ObjectMeta))
		o.Expect(err).NotTo(o.HaveOccurred())
		event, err = watch.Until(CertificateProvisioningTimeout, w, util.RouteTLSChangedFunc(route.Spec.TLS))
		o.Expect(err).NotTo(o.HaveOccurred(), "Failed to wait for certificate to be re-provisioned!")

		g.By("validating the initial certificate")
		route = event.Object.(*routev1.Route)
		o.Expect(route.Spec.TLS).NotTo(o.BeNil())

		o.Expect(route.Spec.TLS.Termination).To(o.Equal(routev1.TLSTerminationEdge))

		framework.Logf("route: %#v", route)
		framework.Logf("route tls: %#v", route.Spec.TLS)
		crt, err = util.CertificateFromPEM([]byte(route.Spec.TLS.Certificate))
		o.Expect(err).NotTo(o.HaveOccurred())

		now = time.Now()
		o.Expect(now.Before(crt.NotBefore)).To(o.BeFalse())
		o.Expect(now.After(crt.NotAfter)).To(o.BeFalse())

		// ACME server will likely cache the validation for our domain and won't retry it so soon.
		//err = DeleteACMEAccountIfRequested(f)
		//o.Expect(err).NotTo(o.HaveOccurred())
		// FIXME: test also by replacing with old certificate

	})
})
