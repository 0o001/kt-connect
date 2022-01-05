package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func setupDump2Host(k cluster.KubernetesInterface, currentNamespace, targetNamespaces, clusterDomain string) bool {
	namespacesToDump := []string{currentNamespace}
	if targetNamespaces != "" {
		namespacesToDump = []string{}
		for _, ns := range strings.Split(targetNamespaces, ",") {
			namespacesToDump = append(namespacesToDump, ns)
		}
	}
	hosts := map[string]string{}
	for _, namespace := range namespacesToDump {
		log.Debug().Msgf("Search service in %s namespace ...", namespace)
		host := getServiceHosts(k, namespace)
		for svc, ip := range host {
			if ip == "" || ip == "None" {
				continue
			}
			log.Debug().Msgf("Service found: %s.%s %s", svc, namespace, ip)
			if namespace == currentNamespace {
				hosts[svc] = ip
			}
			hosts[svc+"."+namespace] = ip
			hosts[svc+"."+namespace+"."+clusterDomain] = ip
		}
	}
	return util.DumpHosts(hosts)
}

func getServiceHosts(k cluster.KubernetesInterface, namespace string) map[string]string {
	hosts := map[string]string{}
	services, err := k.GetAllServiceInNamespace(context.TODO(), namespace)
	if err == nil {
		for _, service := range services.Items {
			hosts[service.Name] = service.Spec.ClusterIP
		}
	}
	return hosts
}

func getOrCreateShadow(kubernetes cluster.KubernetesInterface, options *options.DaemonOptions) (string, string, *util.SSHCredential, error) {
	shadowPodName := fmt.Sprintf("kt-connect-shadow-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.SharedShadow {
		shadowPodName = fmt.Sprintf("kt-connect-shadow-daemon")
	}

	endPointIP, podName, credential, err := cluster.GetOrCreateShadow(context.TODO(), kubernetes,
		shadowPodName, options, getLabels(shadowPodName), make(map[string]string), getEnvs())
	if err != nil {
		return "", "", nil, err
	}

	return endPointIP, podName, credential, nil
}

func getEnvs() map[string]string {
	envs := make(map[string]string)
	localDomains := util.GetLocalDomains()
	if localDomains != "" {
		log.Debug().Msgf("Found local domains: %s", localDomains)
		envs[common.EnvVarLocalDomains] = localDomains
	}
	return envs
}

func getLabels(workload string) map[string]string {
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RoleConnectShadow,
		common.KtName:    workload,
	}
	return labels
}
