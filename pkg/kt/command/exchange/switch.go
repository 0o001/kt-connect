package exchange

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"strings"
)

func BySwitchOver(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// 1. Get service to exchange
	svcName, err := general.GetServiceByResourceName(ctx, k, resourceName, opts)
	if err != nil {
		return err
	}

	// 2. Lock service to avoid conflict
	svc, err := general.LockAndFetchService(ctx, k, svcName, opts.Namespace, 0)
	if err != nil {
		return err
	}
	defer general.UnlockService(ctx, k, svcName, opts.Namespace)

	// 3. Create shadow pod
	shadowName := svc.Name + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))
	shadowLabels := map[string]string{
		common.KtRole: common.RoleExchangeShadow,
		common.KtName: shadowName,
	}
	if err = general.CreateShadowAndInbound(ctx, k, shadowName, util.CopyMap(shadowLabels), map[string]string{}, opts); err != nil {
		return err
	}

	// 4. Let target service select shadow pod
	opts.RuntimeOptions.Origin = svcName
	if err = general.UpdateServiceSelector(ctx, k, svc, shadowLabels); err != nil {
		return err
	}

	return nil
}