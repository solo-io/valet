package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	CertManagerManifest      = "https://github.com/jetstack/cert-manager/releases/download/v0.10.0/cert-manager-no-webhook.yaml"
	CertManagerNamespace     = "cert-manager"
	CertManagerAwsSecretName = "cert-manager"
	IssuerName               = "letsencrypt-dns-prod"
)

var (
	_ Resource = new(CertManager)

	manifest      = &Manifest{Path: CertManagerManifest}
	awsSecret     = AwsSecret(CertManagerNamespace, CertManagerAwsSecretName)
	clusterIssuer = &ClusterIssuer{}
)

type CertManager struct {
}

func (c *CertManager) Ensure(ctx context.Context, command cmd.Factory) error {
	return EnsureAll(ctx, command, manifest, awsSecret, clusterIssuer)
}

func (c *CertManager) Teardown(ctx context.Context, command cmd.Factory) error {
	return TeardownAll(ctx, command, manifest, awsSecret, clusterIssuer)
}

type ClusterIssuer struct {
}

func (c *ClusterIssuer) Ensure(ctx context.Context, command cmd.Factory) error {
	issuer, err := getIssuer()
	if err != nil {
		return err
	}
	return command.Kubectl().ApplyStdIn(issuer).Run(ctx)
}

func (c *ClusterIssuer) Teardown(ctx context.Context, command cmd.Factory) error {
	issuer, err := getIssuer()
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(issuer).Run(ctx)
}

func getIssuer() (string, error) {
	accessKeyId, err := GetEnvVar(AwsAccessKeyIdEnvVar)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: %s
spec:
  acme:
    dns01:
      providers:
        - name: route53
          route53:
            accessKeyID: %s
            region: us-east-1
            secretAccessKeySecretRef:
              key: secret_access_key
              name: cert-manager
    email: yuval@solo.io
    privateKeySecretRef:
      name: letsencrypt-dns-prod
    server: https://acme-v02.api.letsencrypt.org/directory
`, IssuerName, accessKeyId), nil
}
