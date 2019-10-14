package internal

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"strings"
)

func GetCurrentContextName(ctx context.Context) (string, error) {
	contextutils.LoggerFrom(ctx).Infow("Getting current context name")
	out, err := cmd.Kubectl().CurrentContext().Output()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error getting current context name",
			zap.Error(err), zap.String("out", out))
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func CreateDomain(ctx context.Context, appName, hostedZone string) (string, error) {
	currentContext, err := GetCurrentContextName(ctx)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write([]byte(currentContext))
	bs := h.Sum(nil)
	fullHash := fmt.Sprintf("%x", bs)
	domain := fmt.Sprintf("valet-%s-%s.%s", appName, fullHash[:10], hostedZone)
	domain = strings.TrimSuffix(domain, ".")
	return domain, nil
}

func CreateCert(name, namespace, domain string) string {
	return fmt.Sprintf(`
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: %s
  namespace: %s
spec:
  secretName: %s
  dnsNames:
    - %s
  acme:
    config:
      - dns01:
          provider: route53
        domains:
          - %s
  issuerRef:
    name: letsencrypt-dns-prod
    kind: ClusterIssuer`, name, namespace, domain, domain, domain)
}