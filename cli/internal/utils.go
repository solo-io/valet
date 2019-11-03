package internal

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

func GetCurrentContextName(ctx context.Context, command cmd.Factory) (string, error) {
	out, err := command.Kubectl().CurrentContext().Cmd().Output(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func CreateDomainString(ctx context.Context, command cmd.Factory, appName, hostedZone string) (string, error) {
	currentContext, err := GetCurrentContextName(ctx, command)
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
