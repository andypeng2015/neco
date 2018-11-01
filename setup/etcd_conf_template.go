package setup

const etcdConfTemplate = `# This is the configuration file for the etcd server.

# Human-readable name for this member.
name: 'boot-{{.LRN}}'

# Path to the data directory.
data-dir: '/var/lib/etcd'

# List of comma separated URLs to listen on for peer traffic.
listen-peer-urls: https://0.0.0.0:2380

# List of comma separated URLs to listen on for client traffic.
listen-client-urls: https://0.0.0.0:2379

# List of this member's peer URLs to advertise to the rest of the cluster.
# The URLs needed to be a comma-separated list.
initial-advertise-peer-urls: {{.InitialAdvertisePeerURLs}}

# List of this member's client URLs to advertise to the public.
# The URLs needed to be a comma-separated list.
advertise-client-urls: {{.AdvertiseClientURLs}}

# Initial cluster configuration for bootstrapping.
initial-cluster: {{.InitialCluster}}

# Initial cluster token for the etcd cluster during bootstrap.
initial-cluster-token: 'boot-cluster'

# Initial cluster state ('new' or 'existing').
initial-cluster-state: '{{.InitialClusterState}}'

# Accept etcd V2 client requests
enable-v2: false

# Enable runtime profiling data via HTTP server
enable-pprof: true

# TLS certificates
client-transport-security:
  cert-file: /etc/neco/server.crt
  key-file: /etc/neco/server.key
  client-cert-auth: true
  trusted-ca-file: /etc/etcd/ca-client.crt

peer-transport-security:
  cert-file: /etc/etcd/peer.crt
  key-file: /etc/etcd/peer.key
  client-cert-auth: true
  trusted-ca-file: /etc/etcd/ca-peer.crt

# Specify 'stdout' or 'stderr' to skip journald logging even when running under systemd.
log-outputs: stderr

# auto compaction
auto-compaction-mode: periodic
auto-compaction-retention: "24"
`
