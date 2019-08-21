package server

import (
	"github.com/cybozu-go/cke"
	"github.com/cybozu-go/cke/op"
	"github.com/cybozu-go/cke/op/clusterdns"
	"github.com/cybozu-go/cke/op/etcd"
	"github.com/cybozu-go/cke/op/etcdbackup"
	"github.com/cybozu-go/cke/op/k8s"
	"github.com/cybozu-go/cke/op/nodedns"
	"github.com/cybozu-go/cke/static"
	"github.com/cybozu-go/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DecideOps returns the next operations to do.
// This returns nil when no operation need to be done.
func DecideOps(c *cke.Cluster, cs *cke.ClusterStatus, resources []cke.ResourceDefinition) []cke.Operator {

	nf := NewNodeFilter(c, cs)

	// 1. Run or restart rivers.  This guarantees:
	// - CKE tools image is pulled on all nodes.
	// - Rivers runs on all nodes and will proxy requests only to control plane nodes.
	if ops := riversOps(c, nf); len(ops) > 0 {
		return ops
	}

	// 2. Bootstrap etcd cluster, if not yet.
	if !nf.EtcdBootstrapped() {
		return []cke.Operator{etcd.BootOp(nf.ControlPlane(), c.Options.Etcd, c.Options.Kubelet.Domain)}
	}

	// 3. Start etcd containers.
	if nodes := nf.EtcdStoppedMembers(); len(nodes) > 0 {
		return []cke.Operator{etcd.StartOp(nodes, c.Options.Etcd, c.Options.Kubelet.Domain)}
	}

	// 4. Wait for etcd cluster to become ready
	if !cs.Etcd.IsHealthy {
		return []cke.Operator{etcd.WaitClusterOp(nf.ControlPlane())}
	}

	// 5. Run or restart kubernetes components.
	if ops := k8sOps(c, nf); len(ops) > 0 {
		return ops
	}

	// 6. Maintain etcd cluster.
	if o := etcdMaintOp(c, nf); o != nil {
		return []cke.Operator{o}
	}

	// 7. Maintain k8s resources.
	if ops := k8sMaintOps(c, cs, resources, nf); len(ops) > 0 {
		return ops
	}

	// 8. Stop and delete control plane services running on non control plane nodes.
	if ops := cleanOps(c, nf); len(ops) > 0 {
		return ops
	}

	return nil
}

func riversOps(c *cke.Cluster, nf *NodeFilter) (ops []cke.Operator) {
	if nodes := nf.RiversStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, op.RiversBootOp(nodes, nf.ControlPlane(), c.Options.Rivers, op.RiversContainerName, op.RiversUpstreamPort, op.RiversListenPort))
	}
	if nodes := nf.RiversOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, op.RiversRestartOp(nodes, nf.ControlPlane(), c.Options.Rivers, op.RiversContainerName, op.RiversUpstreamPort, op.RiversListenPort))
	}
	if nodes := nf.EtcdRiversStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, op.RiversBootOp(nodes, nf.ControlPlane(), c.Options.EtcdRivers, op.EtcdRiversContainerName, op.EtcdRiversUpstreamPort, op.EtcdRiversListenPort))
	}
	if nodes := nf.EtcdRiversOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, op.RiversRestartOp(nodes, nf.ControlPlane(), c.Options.EtcdRivers, op.EtcdRiversContainerName, op.EtcdRiversUpstreamPort, op.EtcdRiversListenPort))
	}
	return ops
}

func k8sOps(c *cke.Cluster, nf *NodeFilter) (ops []cke.Operator) {
	if nodes := nf.APIServerStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.APIServerBootOp(nodes, nf.ControlPlane(), c.ServiceSubnet, c.Options.Kubelet.Domain, c.Options.APIServer))
	}
	if nodes := nf.APIServerOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.APIServerRestartOp(nodes, nf.ControlPlane(), c.ServiceSubnet, c.Options.Kubelet.Domain, c.Options.APIServer))
	}
	if nodes := nf.ControllerManagerStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.ControllerManagerBootOp(nodes, c.Name, c.ServiceSubnet, c.Options.ControllerManager))
	}
	if nodes := nf.ControllerManagerOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.ControllerManagerRestartOp(nodes, c.Name, c.ServiceSubnet, c.Options.ControllerManager))
	}
	if nodes := nf.SchedulerStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.SchedulerBootOp(nodes, c.Name, c.Options.Scheduler))
	}
	if nodes := nf.SchedulerOutdatedNodes(c.Options.Scheduler.Extenders); len(nodes) > 0 {
		ops = append(ops, k8s.SchedulerRestartOp(nodes, c.Name, c.Options.Scheduler))
	}
	if nodes := nf.KubeletUnrecognizedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.KubeletRestartOp(nodes, c.Name, c.ServiceSubnet, c.Options.Kubelet))
	}
	if nodes := nf.KubeletStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.KubeletBootOp(nodes, nf.KubeletStoppedRegisteredNodes(), nf.HealthyAPIServer(), c.Name, c.PodSubnet, c.Options.Kubelet))
	}
	if nodes := nf.KubeletOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.KubeletRestartOp(nodes, c.Name, c.ServiceSubnet, c.Options.Kubelet))
	}
	if nodes := nf.ProxyStoppedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.KubeProxyBootOp(nodes, c.Name, c.Options.Proxy))
	}
	if nodes := nf.ProxyOutdatedNodes(); len(nodes) > 0 {
		ops = append(ops, k8s.KubeProxyRestartOp(nodes, c.Name, c.Options.Proxy))
	}
	return ops
}

func etcdMaintOp(c *cke.Cluster, nf *NodeFilter) cke.Operator {
	if members := nf.EtcdNonClusterMembers(false); len(members) > 0 {
		return etcd.RemoveMemberOp(nf.ControlPlane(), members)
	}
	if nodes, ids := nf.EtcdNonCPMembers(false); len(nodes) > 0 {
		return etcd.DestroyMemberOp(nf.ControlPlane(), nodes, ids)
	}
	if nodes := nf.EtcdUnstartedMembers(); len(nodes) > 0 {
		return etcd.AddMemberOp(nf.ControlPlane(), nodes[0], c.Options.Etcd, c.Options.Kubelet.Domain)
	}

	if !nf.EtcdIsGood() {
		log.Warn("etcd is not good for maintenance", nil)
		// return nil to proceed to k8s maintenance.
		return nil
	}

	// Adding members or removing/restarting healthy members is done only when
	// all members are in sync.

	if nodes := nf.EtcdNewMembers(); len(nodes) > 0 {
		return etcd.AddMemberOp(nf.ControlPlane(), nodes[0], c.Options.Etcd, c.Options.Kubelet.Domain)
	}
	if members := nf.EtcdNonClusterMembers(true); len(members) > 0 {
		return etcd.RemoveMemberOp(nf.ControlPlane(), members)
	}
	if nodes, ids := nf.EtcdNonCPMembers(true); len(ids) > 0 {
		return etcd.DestroyMemberOp(nf.ControlPlane(), nodes, ids)
	}
	if nodes := nf.EtcdOutdatedMembers(); len(nodes) > 0 {
		return etcd.RestartOp(nf.ControlPlane(), nodes[0], c.Options.Etcd)
	}

	return nil
}

func k8sMaintOps(c *cke.Cluster, cs *cke.ClusterStatus, resources []cke.ResourceDefinition, nf *NodeFilter) (ops []cke.Operator) {
	ks := cs.Kubernetes
	apiServer := nf.HealthyAPIServer()

	if !ks.IsControlPlaneReady {
		return []cke.Operator{op.KubeWaitOp(apiServer)}
	}

	ops = append(ops, decideResourceOps(apiServer, ks, resources, ks.IsReady(c))...)

	ops = append(ops, decideClusterDNSOps(apiServer, c, ks)...)

	ops = append(ops, decideNodeDNSOps(apiServer, c, ks)...)

	cpAddresses := make([]corev1.EndpointAddress, len(nf.ControlPlane()))
	for i, cp := range nf.ControlPlane() {
		cpAddresses[i] = corev1.EndpointAddress{
			IP: cp.Address,
		}
	}

	masterEP := &corev1.Endpoints{}
	masterEP.Namespace = metav1.NamespaceDefault
	masterEP.Name = "kubernetes"
	masterEP.Subsets = []corev1.EndpointSubset{
		{
			Addresses: cpAddresses,
			Ports: []corev1.EndpointPort{
				{
					Name:     "https",
					Port:     6443,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	epOp := decideEpOp(masterEP, ks.MasterEndpoints, apiServer)
	if epOp != nil {
		ops = append(ops, epOp)
	}

	// Endpoints needs a corresponding Service.
	// If an Endpoints lacks such a Service, it will be removed.
	// https://github.com/kubernetes/kubernetes/blob/b7c2d923ef4e166b9572d3aa09ca72231b59b28b/pkg/controller/endpoint/endpoints_controller.go#L392-L397
	svcOp := decideEtcdServiceOps(apiServer, ks.EtcdService)
	if svcOp != nil {
		ops = append(ops, svcOp)
	}

	etcdEP := &corev1.Endpoints{}
	etcdEP.Namespace = metav1.NamespaceSystem
	etcdEP.Name = op.EtcdEndpointsName
	etcdEP.Subsets = []corev1.EndpointSubset{
		{
			Addresses: cpAddresses,
			Ports: []corev1.EndpointPort{
				{
					Port:     2379,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	epOp = decideEpOp(etcdEP, ks.EtcdEndpoints, apiServer)
	if epOp != nil {
		ops = append(ops, epOp)
	}

	if nodes := nf.OutdatedAttrsNodes(); len(nodes) > 0 {
		ops = append(ops, op.KubeNodeUpdateOp(apiServer, nodes))
	}

	if nodes := nf.NonClusterNodes(); len(nodes) > 0 {
		ops = append(ops, op.KubeNodeRemoveOp(apiServer, nodes))
	}

	ops = append(ops, decideEtcdBackupOps(apiServer, c, ks)...)

	return ops
}

func decideClusterDNSOps(apiServer *cke.Node, c *cke.Cluster, ks cke.KubernetesClusterStatus) (ops []cke.Operator) {
	desiredDNSServers := c.DNSServers
	if ks.DNSService != nil {
		switch ip := ks.DNSService.Spec.ClusterIP; ip {
		case "", "None":
		default:
			desiredDNSServers = []string{ip}
		}
	}
	desiredClusterDomain := c.Options.Kubelet.Domain

	if len(desiredClusterDomain) == 0 {
		panic("Options.Kubelet.Domain is empty")
	}

	if ks.ClusterDNS.ConfigMap == nil {
		ops = append(ops, clusterdns.CreateConfigMapOp(apiServer, desiredClusterDomain, desiredDNSServers))
	} else {
		actualConfigData := ks.ClusterDNS.ConfigMap.Data
		expectedConfig := clusterdns.ConfigMap(desiredClusterDomain, desiredDNSServers)
		if actualConfigData["Corefile"] != expectedConfig.Data["Corefile"] {
			ops = append(ops, clusterdns.UpdateConfigMapOp(apiServer, expectedConfig))
		}
	}

	return ops
}

func decideNodeDNSOps(apiServer *cke.Node, c *cke.Cluster, ks cke.KubernetesClusterStatus) (ops []cke.Operator) {
	if len(ks.ClusterDNS.ClusterIP) == 0 {
		return nil
	}

	desiredDNSServers := c.DNSServers
	if ks.DNSService != nil {
		switch ip := ks.DNSService.Spec.ClusterIP; ip {
		case "", "None":
		default:
			desiredDNSServers = []string{ip}
		}
	}

	if ks.NodeDNS.ConfigMap == nil {
		ops = append(ops, nodedns.CreateConfigMapOp(apiServer, ks.ClusterDNS.ClusterIP, c.Options.Kubelet.Domain, desiredDNSServers))
	} else {
		actualConfigData := ks.NodeDNS.ConfigMap.Data
		expectedConfig := nodedns.ConfigMap(ks.ClusterDNS.ClusterIP, c.Options.Kubelet.Domain, desiredDNSServers)
		if actualConfigData["unbound.conf"] != expectedConfig.Data["unbound.conf"] {
			ops = append(ops, nodedns.UpdateConfigMapOp(apiServer, expectedConfig))
		}
	}

	return ops
}

func decideEpOp(expect, actual *corev1.Endpoints, apiServer *cke.Node) cke.Operator {
	if actual == nil {
		return op.KubeEndpointsCreateOp(apiServer, expect)
	}

	updateOp := op.KubeEndpointsUpdateOp(apiServer, expect)
	if len(actual.Subsets) != 1 {
		return updateOp
	}

	subset := actual.Subsets[0]
	if len(subset.Ports) != 1 || subset.Ports[0].Port != expect.Subsets[0].Ports[0].Port {
		return updateOp
	}

	if len(subset.Addresses) != len(expect.Subsets[0].Addresses) {
		return updateOp
	}

	endpoints := make(map[string]bool)
	for _, a := range expect.Subsets[0].Addresses {
		endpoints[a.IP] = true
	}
	for _, a := range subset.Addresses {
		if !endpoints[a.IP] {
			return updateOp
		}
	}

	return nil
}

func decideEtcdServiceOps(apiServer *cke.Node, svc *corev1.Service) cke.Operator {
	if svc == nil {
		return op.KubeEtcdServiceCreateOp(apiServer)
	}

	updateOp := op.KubeEtcdServiceUpdateOp(apiServer)

	if len(svc.Spec.Ports) != 1 {
		return updateOp
	}
	if svc.Spec.Ports[0].Port != 2379 {
		return updateOp
	}
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return updateOp
	}
	if svc.Spec.ClusterIP != corev1.ClusterIPNone {
		return updateOp
	}

	return nil
}

func decideEtcdBackupOps(apiServer *cke.Node, c *cke.Cluster, ks cke.KubernetesClusterStatus) (ops []cke.Operator) {
	if c.EtcdBackup.Enabled == false {
		if ks.EtcdBackup.ConfigMap != nil {
			ops = append(ops, etcdbackup.ConfigMapRemoveOp(apiServer))
		}
		if ks.EtcdBackup.Secret != nil {
			ops = append(ops, etcdbackup.SecretRemoveOp(apiServer))
		}
		if ks.EtcdBackup.CronJob != nil {
			ops = append(ops, etcdbackup.CronJobRemoveOp(apiServer))
		}
		if ks.EtcdBackup.Service != nil {
			ops = append(ops, etcdbackup.ServiceRemoveOp(apiServer))
		}
		if ks.EtcdBackup.Pod != nil {
			ops = append(ops, etcdbackup.PodRemoveOp(apiServer))
		}
		return ops
	}

	if ks.EtcdBackup.ConfigMap == nil {
		ops = append(ops, etcdbackup.ConfigMapCreateOp(apiServer, c.EtcdBackup.Rotate))
	} else {
		actual := ks.EtcdBackup.ConfigMap.Data["config.yml"]
		expected := etcdbackup.RenderConfigMap(c.EtcdBackup.Rotate).Data["config.yml"]
		if actual != expected {
			ops = append(ops, etcdbackup.ConfigMapUpdateOp(apiServer, c.EtcdBackup.Rotate))
		}
	}
	if ks.EtcdBackup.Secret == nil {
		ops = append(ops, etcdbackup.SecretCreateOp(apiServer))
	}
	if ks.EtcdBackup.Service == nil {
		ops = append(ops, etcdbackup.ServiceCreateOp(apiServer))
	}
	if ks.EtcdBackup.Pod == nil {
		ops = append(ops, etcdbackup.PodCreateOp(apiServer, c.EtcdBackup.PVCName))
	} else if needUpdateEtcdBackupPod(c, ks) {
		ops = append(ops, etcdbackup.PodUpdateOp(apiServer, c.EtcdBackup.PVCName))
	}

	if ks.EtcdBackup.CronJob == nil {
		ops = append(ops, etcdbackup.CronJobCreateOp(apiServer, c.EtcdBackup.Schedule))
	} else if ks.EtcdBackup.CronJob.Spec.Schedule != c.EtcdBackup.Schedule {
		ops = append(ops, etcdbackup.CronJobUpdateOp(apiServer, c.EtcdBackup.Schedule))
	}

	return ops
}

func needUpdateEtcdBackupPod(c *cke.Cluster, ks cke.KubernetesClusterStatus) bool {
	volumes := ks.EtcdBackup.Pod.Spec.Volumes
	vol := new(corev1.Volume)
	for _, v := range volumes {
		if v.Name == "etcdbackup" {
			vol = &v
			break
		}
	}
	if vol == nil {
		return true
	}

	if vol.PersistentVolumeClaim == nil {
		return true
	}
	if vol.PersistentVolumeClaim.ClaimName != c.EtcdBackup.PVCName {
		return true
	}
	return false
}

func decideResourceOps(apiServer *cke.Node, ks cke.KubernetesClusterStatus, resources []cke.ResourceDefinition, isReady bool) (ops []cke.Operator) {
	for _, res := range static.Resources {
		// To avoid thundering herd problem. Deployments need to be created only after enough nodes become ready.
		if res.Kind == cke.KindDeployment && !isReady {
			continue
		}
		annotations, ok := ks.ResourceStatuses[res.Key]
		if !ok || res.NeedUpdate(annotations) {
			ops = append(ops, op.ResourceApplyOp(apiServer, res))
		}
	}
	for _, res := range resources {
		if res.Kind == cke.KindDeployment && !isReady {
			continue
		}
		annotations, ok := ks.ResourceStatuses[res.Key]
		if !ok || res.NeedUpdate(annotations) {
			ops = append(ops, op.ResourceApplyOp(apiServer, res))
		}
	}
	return ops
}

func cleanOps(c *cke.Cluster, nf *NodeFilter) (ops []cke.Operator) {
	var apiServers, controllerManagers, schedulers, etcds, etcdRivers []*cke.Node

	for _, n := range c.Nodes {
		if n.ControlPlane {
			continue
		}

		st := nf.nodeStatus(n)
		if st.Etcd.Running && nf.EtcdIsGood() {
			etcds = append(etcds, n)
		}
		if st.APIServer.Running {
			apiServers = append(apiServers, n)
		}
		if st.ControllerManager.Running {
			controllerManagers = append(controllerManagers, n)
		}
		if st.Scheduler.Running {
			schedulers = append(schedulers, n)
		}
		if st.EtcdRivers.Running {
			etcdRivers = append(etcdRivers, n)
		}
	}

	if len(apiServers) > 0 {
		ops = append(ops, op.APIServerStopOp(apiServers))
	}
	if len(controllerManagers) > 0 {
		ops = append(ops, op.ControllerManagerStopOp(controllerManagers))
	}
	if len(schedulers) > 0 {
		ops = append(ops, op.SchedulerStopOp(schedulers))
	}
	if len(etcds) > 0 {
		ops = append(ops, op.EtcdStopOp(etcds))
	}
	if len(etcdRivers) > 0 {
		ops = append(ops, op.EtcdRiversStopOp(etcdRivers))
	}
	return ops
}