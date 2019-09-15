package gloo

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"strings"
)

type UiVirtualServiceCreator interface {
	Create(ctx context.Context, glooui options.UiVirtualService) error
}

var _ UiVirtualServiceCreator = new(kubectlUiVirtualServiceCreator)

func NewKubectlUiVirtualServiceCreator() *kubectlUiVirtualServiceCreator {
	return &kubectlUiVirtualServiceCreator{}
}

type kubectlUiVirtualServiceCreator struct {}

const (
	GlooUiVirtualService = `
apiVersion: gateway.solo.io/v1
kind: VirtualService
metadata:
  name: glooui
  namespace: gloo-system
spec:
  virtualHost:
    domains:
      - "*"
    name: gloo-system.glooui
    routes:
      - matcher:
          prefix: /
        routeAction:
          single:
            upstream:
              name: gloo-system-apiserver-ui-8080
              namespace: gloo-system`
)

func (k *kubectlUiVirtualServiceCreator) Create(ctx context.Context, glooui options.UiVirtualService) error {
	contextutils.LoggerFrom(ctx).Infow("Creating ui virtual service")
	out, err := internal.ExecuteCmdStdIn(GlooUiVirtualService, "kubectl", "apply", "-f", "-")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating glooui virtualservice",
			zap.Error(err), zap.String("out", out))
		return err
	}

	if glooui.DNS == nil || glooui.DNS.HostedZone == "" {
		contextutils.LoggerFrom(ctx).Infow("No DNS config provided")
		return nil
	}

	client, err := NewAwsDnsClient()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating aws dns client", zap.Error(err))
		return err
	}

	proxyIp, err := GetGlooProxyExternalIp(ctx)
	if err != nil {
		return err
	}
	domain := glooui.DNS.Domain
	if domain == "" {
		domain, err = internal.CreateDomain(ctx, "glooui", glooui.DNS.HostedZone)
		if err != nil {
			return err
		}
	}
	err = client.CreateMapping(ctx, glooui.DNS.HostedZone, domain, proxyIp)
	if err != nil {
		return err
	}

	return patchGloouiWithDomain(ctx, domain)
}

func patchGloouiWithDomain(ctx context.Context, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("Patching glooui domain")
	patchStr := fmt.Sprintf("-p=[{\"op\":\"add\",\"path\":\"/spec/virtualHost/domains\",\"value\":[\"%s\"]}]", domain)
	out, err := internal.ExecuteCmd("kubectl", "patch", "vs", "glooui", "-n", "gloo-system", "--type=json", patchStr)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error patching glooui virtualservice",
			zap.Error(err), zap.String("out", out), zap.String("domain", domain))
		return err
	}
	return nil
}

func GetGlooProxyExternalIp(ctx context.Context) (string, error) {
	contextutils.LoggerFrom(ctx).Infow("Getting Gloo proxy ip")
	out, err := internal.ExecuteCmd("kubectl", "get", "svc", "-n", "gloo-system", "gateway-proxy-v2", "-o=jsonpath={.status.loadBalancer.ingress[0].ip}")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating glooui virtualservice",
			zap.Error(err), zap.String("out", out))
		return "", err
	}
	return strings.TrimSpace(out), nil
}







