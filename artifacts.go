// Code generated by generate-artifacts. DO NOT EDIT.
// +build !release

package neco

var CurrentArtifacts = ArtifactSet{
	Images: []ContainerImage{
		{Name: "coil", Repository: "ghcr.io/cybozu-go/coil", Tag: "2.0.7", Private: false},
		{Name: "bird", Repository: "quay.io/cybozu/bird", Tag: "2.0.7.5", Private: false},
		{Name: "chrony", Repository: "quay.io/cybozu/chrony", Tag: "4.0.0.1", Private: false},
		{Name: "etcd", Repository: "quay.io/cybozu/etcd", Tag: "3.3.25.4", Private: false},
		{Name: "promtail", Repository: "quay.io/cybozu/promtail", Tag: "2.1.0.1", Private: false},
		{Name: "sabakan", Repository: "quay.io/cybozu/sabakan", Tag: "2.5.7", Private: false},
		{Name: "serf", Repository: "quay.io/cybozu/serf", Tag: "0.9.5.2", Private: false},
		{Name: "setup-hw", Repository: "quay.io/cybozu/setup-hw", Tag: "1.7.1", Private: true},
		{Name: "squid", Repository: "quay.io/cybozu/squid", Tag: "4.14.1", Private: false},
		{Name: "vault", Repository: "quay.io/cybozu/vault", Tag: "1.6.4.1", Private: false},
	},
	Debs: []DebianPackage{
		{Name: "etcdpasswd", Owner: "cybozu-go", Repository: "etcdpasswd", Release: "v1.1.2"},
	},
	CoreOS: CoreOSImage{Channel: "stable", Version: "2765.2.3"},
}
