// Code generated by generate-artifacts. DO NOT EDIT.
//go:build !release

package neco

var CurrentArtifacts = ArtifactSet{
	Images: []ContainerImage{
		{Name: "coil", Repository: "ghcr.io/cybozu-go/coil", Tag: "2.4.0", Private: false},
		{Name: "bird", Repository: "quay.io/cybozu/bird", Tag: "2.13.1.1", Private: false},
		{Name: "chrony", Repository: "quay.io/cybozu/chrony", Tag: "4.3.0.3", Private: false},
		{Name: "etcd", Repository: "quay.io/cybozu/etcd", Tag: "3.5.9.1", Private: false},
		{Name: "promtail", Repository: "quay.io/cybozu/promtail", Tag: "2.9.1.1", Private: false},
		{Name: "sabakan", Repository: "quay.io/cybozu/sabakan", Tag: "2.13.2", Private: false},
		{Name: "serf", Repository: "quay.io/cybozu/serf", Tag: "0.10.1.3", Private: false},
		{Name: "setup-hw", Repository: "quay.io/cybozu/setup-hw", Tag: "1.14.3", Private: true},
		{Name: "squid", Repository: "quay.io/cybozu/squid", Tag: "5.8.0.2", Private: false},
		{Name: "vault", Repository: "quay.io/cybozu/vault", Tag: "1.14.3.1", Private: false},
		{Name: "cilium", Repository: "quay.io/cybozu/cilium", Tag: "1.13.7.1", Private: false},
		{Name: "cilium-operator-generic", Repository: "quay.io/cybozu/cilium-operator-generic", Tag: "1.13.7.1", Private: false},
		{Name: "hubble-relay", Repository: "quay.io/cybozu/hubble-relay", Tag: "1.13.7.1", Private: false},
		{Name: "cilium-certgen", Repository: "quay.io/cybozu/cilium-certgen", Tag: "0.1.9.1", Private: false},
	},
	Debs: []DebianPackage{
		{Name: "etcdpasswd", Owner: "cybozu-go", Repository: "etcdpasswd", Release: "v1.4.3"},
	},
	OSImage: OSImage{Channel: "stable", Version: "3510.2.8"},
}
