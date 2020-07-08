package ingresswatcher

import (
	"io"

	"github.com/cybozu-go/neco"
)

// GenerateService generates systemd service unit contents.
func GenerateService(w io.Writer) error {
	return serviceTmpl.Execute(w, struct {
		ConfigFile string
	}{
		ConfigFile: neco.IngressWatcherConfFile,
	})
}
