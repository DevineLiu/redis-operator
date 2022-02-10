package cert

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	v12 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CertificateRequest struct {
	Name       string
	Namespace  string
	Duration   *metav1.Duration
	DNSNames   []string
	SecretName string
}

type ICertificateControl interface {
	Create(*certv1.Certificate) error
	GetCertificate(namespace, name string) (*certv1.Certificate, error)
	CreateIfNotExists(*certv1.Certificate) error
}

type certificateController struct {
	client client.Client
}

func NewCertificateController(client client.Client) ICertificateControl {
	return &certificateController{
		client: client,
	}
}

func GenerateCert(certReq CertificateRequest, labels map[string]string) *certv1.Certificate {
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certReq.Name,
			Namespace: certReq.Namespace,
			Labels:    labels,
		},
		Spec: certv1.CertificateSpec{
			Duration:   certReq.Duration,
			DNSNames:   certReq.DNSNames,
			IssuerRef:  v12.ObjectReference{Kind: certv1.ClusterIssuerKind, Name: "cpaas-ca"},
			SecretName: certReq.SecretName,
			SecretTemplate: &certv1.CertificateSecretTemplate{
				Labels: labels,
			},
		},
	}
}

func (s *certificateController) GetCertificate(name, namespace string) (*certv1.Certificate, error) {
	cert := &certv1.Certificate{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cert)
	return cert, err
}

func (s *certificateController) Create(cert *certv1.Certificate) error {
	return s.client.Create(context.TODO(), cert)
}

func (s *certificateController) CreateIfNotExists(cert *certv1.Certificate) error {
	if _, err := s.GetCertificate(cert.Namespace, cert.Name); err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.Create(cert)
		}
		return err
	}
	return nil
}
