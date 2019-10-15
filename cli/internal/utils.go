package internal

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"strings"
)

func GetCurrentContextName(ctx context.Context) (string, error) {
	out, err := cmd.Kubectl().CurrentContext().Output(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func CreateDomainString(ctx context.Context, appName, hostedZone string) (string, error) {
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

func CreateCertString(name, namespace, domain string) string {
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
