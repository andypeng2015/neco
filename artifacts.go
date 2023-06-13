// Code generated by generate-artifacts. DO NOT EDIT.
//go:build !release

package neco

var CurrentArtifacts = ArtifactSet{
	Images: []ContainerImage{
		{Name: "coil", Repository: "ghcr.io/cybozu-go/coil", Tag: "2.3.0", Private: false},
		{Name: "bird", Repository: "quay.io/cybozu/bird", Tag: "2.0.12.1", Private: false},
		{Name: "chrony", Repository: "quay.io/cybozu/chrony", Tag: "4.3.0.2", Private: false},
		{Name: "etcd", Repository: "quay.io/cybozu/etcd", Tag: "3.5.7.2", Private: false},
		{Name: "promtail", Repository: "quay.io/cybozu/promtail", Tag: "2.8.0.2", Private: false},
		{Name: "sabakan", Repository: "quay.io/cybozu/sabakan", Tag: "2.13.2", Private: false},
		{Name: "serf", Repository: "quay.io/cybozu/serf", Tag: "0.10.1.1", Private: false},
		{Name: "setup-hw", Repository: "quay.io/cybozu/setup-hw", Tag: "1.13.2", Private: true},
		{Name: "squid", Repository: "quay.io/cybozu/squid", Tag: "5.8.0.2", Private: false},
		{Name: "vault", Repository: "quay.io/cybozu/vault", Tag: "1.13.0.1", Private: false},
		{Name: "cilium", Repository: "quay.io/cybozu/cilium", Tag: "1.12.8.4", Private: false},
		{Name: "cilium-operator-generic", Repository: "quay.io/cybozu/cilium-operator-generic", Tag: "1.12.8.1", Private: false},
		{Name: "hubble-relay", Repository: "quay.io/cybozu/hubble-relay", Tag: "1.12.8.1", Private: false},
		{Name: "cilium-certgen", Repository: "quay.io/cybozu/cilium-certgen", Tag: "0.1.8.1", Private: false},
	},
	Debs: []DebianPackage{
		{Name: "etcdpasswd", Owner: "cybozu-go", Repository: "etcdpasswd", Release: "v1.4.2"},
	},
	OSImage: OSImage{Channel: "stable", Version: "3510.2.1"},
}
