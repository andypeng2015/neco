// Code generated by generate-artifacts. DO NOT EDIT.
// +build !release

package neco

var CurrentArtifacts = ArtifactSet{
	Images: []ContainerImage{
		{Name: "cke", Repository: "quay.io/cybozu/cke", Tag: "1.17.3", Private: false},
		{Name: "etcd", Repository: "quay.io/cybozu/etcd", Tag: "3.3.20.1", Private: false},
		{Name: "setup-hw", Repository: "quay.io/cybozu/setup-hw", Tag: "1.6.9", Private: true},
		{Name: "sabakan", Repository: "quay.io/cybozu/sabakan", Tag: "2.5.1", Private: false},
		{Name: "serf", Repository: "quay.io/cybozu/serf", Tag: "0.8.5.2", Private: false},
		{Name: "vault", Repository: "quay.io/cybozu/vault", Tag: "1.2.2.2", Private: false},
		{Name: "coil", Repository: "quay.io/cybozu/coil", Tag: "1.1.9", Private: false},
		{Name: "squid", Repository: "quay.io/cybozu/squid", Tag: "3.5.27.1.8", Private: false},
		{Name: "teleport", Repository: "quay.io/cybozu/teleport", Tag: "4.2.8.2", Private: false},
	},
	Debs: []DebianPackage{
		{Name: "etcdpasswd", Owner: "cybozu-go", Repository: "etcdpasswd", Release: "v1.0.0"},
	},
	CoreOS: CoreOSImage{Channel: "stable", Version: "2345.3.0"},
}
