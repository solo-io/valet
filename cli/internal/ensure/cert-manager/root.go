package cert_manager

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
	"os"
)

var (
	NoAccessKeyId = errors.Errorf("AWS_ACCESS_KEY_ID not found")
	NoSecretAccessKey = errors.Errorf("AWS_SECRET_ACCESS_KEY not found")
)

func EnsureCertManager(ctx context.Context, certManager *api.CertManager) error {
	contextutils.LoggerFrom(ctx).Infow("Creating cert-manager resources")
	if _, err := internal.ExecuteCmd("kubectl", "apply", "-f", "https://github.com/jetstack/cert-manager/releases/download/v0.10.0/cert-manager-no-webhook.yaml"); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating cert-manager resources", zap.Error(err))
		return err
	}

	contextutils.LoggerFrom(ctx).Infow("Creating amazon secret for cert-manager")
	if err := createAmazonSecret(); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating amazon secret", zap.Error(err))
		return err
	}

	contextutils.LoggerFrom(ctx).Infow("Creating lets encrypt issuer")
	issuer, err := getIssuer()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error determining issuer", zap.Error(err))
		return err
	}
	_, err = internal.ExecuteCmdStdIn(issuer, "kubectl", "apply", "-f", "-")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error apply issuer manifest", zap.Error(err))
		return err
	}

	return nil
}

func createAmazonSecret() error {
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return NoSecretAccessKey
	}
	accessKeyId, err := getAccessKeyId()
	if err != nil {
		return err
	}
	accessKeyOpt := fmt.Sprintf("--from-literal=access_key_id=%s", accessKeyId)
	secretKeyOpt := fmt.Sprintf("--from-literal=secret_access_key=%s", secretAccessKey)

	out, err := internal.ExecuteCmd("kubectl", "create", "secret", "generic", "cert-manager", "-n", "cert-manager", accessKeyOpt, secretKeyOpt, "--dry-run", "-oyaml")
	if err != nil {
		return err
	}

	_, err = internal.ExecuteCmdStdIn(out, "kubectl", "apply", "-f", "-")
	return err
}

func getAccessKeyId() (string, error) {
	accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyId == "" {
		return "", NoAccessKeyId
	}
	return accessKeyId, nil
}

func getIssuer() (string, error) {
	accessKeyId, err := getAccessKeyId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-dns-prod
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
`, accessKeyId), nil
}