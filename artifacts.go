// Code generated by generate-artifacts. DO NOT EDIT.
//go:build !release

package neco

var CurrentArtifacts = ArtifactSet{
	Images: []ContainerImage{
		{Name: "coil", Repository: "ghcr.io/cybozu-go/coil", Tag: "2.6.0", Private: false},
		{Name: "bird", Repository: "ghcr.io/cybozu/bird", Tag: "2.15.0.1", Private: false},
		{Name: "chrony", Repository: "ghcr.io/cybozu/chrony", Tag: "4.5.0.1", Private: false},
		{Name: "etcd", Repository: "ghcr.io/cybozu/etcd", Tag: "3.5.13.1", Private: false},
		{Name: "promtail", Repository: "ghcr.io/cybozu/promtail", Tag: "2.9.5.1", Private: false},
		{Name: "sabakan", Repository: "ghcr.io/cybozu-go/sabakan", Tag: "3.1.1", Private: false},
		{Name: "serf", Repository: "ghcr.io/cybozu/serf", Tag: "0.10.1.5", Private: false},
		{Name: "setup-hw", Repository: "ghcr.io/cybozu-go/setup-hw", Tag: "1.16.1", Private: true},
		{Name: "squid", Repository: "ghcr.io/cybozu/squid", Tag: "6.9.0.1", Private: false},
		{Name: "squid-exporter", Repository: "ghcr.io/cybozu/squid-exporter", Tag: "1.0.5", Private: false},
		{Name: "vault", Repository: "ghcr.io/cybozu/vault", Tag: "1.16.0.1", Private: false},
		{Name: "cilium", Repository: "quay.io/cybozu/cilium", Tag: "1.13.7.2", Private: false},
		{Name: "cilium-operator-generic", Repository: "ghcr.io/cybozu/cilium-operator-generic", Tag: "1.13.7.3", Private: false},
		{Name: "hubble-relay", Repository: "ghcr.io/cybozu/hubble-relay", Tag: "1.13.7.3", Private: false},
		{Name: "cilium-certgen", Repository: "ghcr.io/cybozu/cilium-certgen", Tag: "0.1.9.2", Private: false},
	},
	Debs: []DebianPackage{
		{Name: "etcdpasswd", Owner: "cybozu-go", Repository: "etcdpasswd", Release: "v1.4.7"},
	},
	OSImage: OSImage{Channel: "stable", Version: "3815.2.1"},
}
