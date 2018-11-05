package updater

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/neco"
	"github.com/cybozu-go/neco/storage"
	version "github.com/hashicorp/go-version"
)

// ReleaseChecker checks newer GitHub releases by polling
type ReleaseChecker struct {
	storage storage.Storage
	github  ReleaseInterface

	pkg      PackageManager
	latest   *atomic.Value
	updated  *sync.WaitGroup
	notfound bool
}

func NewReleaseChecker(st storage.Storage, github ReleaseInterface, pkg PackageManager) ReleaseChecker {
	c := ReleaseChecker{
		storage: st,
		github:  github,
		pkg:     pkg,

		latest:  new(atomic.Value),
		updated: new(sync.WaitGroup),
	}
	c.latest.Store("")
	return c
}

// Run runs newer release at bDaduring periodic intervals
func (c *ReleaseChecker) Run(ctx context.Context) error {
	c.updated = new(sync.WaitGroup)

	for {
		interval, err := c.storage.GetCheckUpdateInterval(ctx)
		if err != nil {
			return err
		}
		c.updated.Done()

		err = c.update(ctx)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			break
		case <-time.After(interval):
		}
	}
	return nil
}

// GetLatest returns latest version in GitHub Releases, or returns empty if no
// release are available
func (c *ReleaseChecker) GetLatest() string {
	if c.notfound {
		return ""
	}
	c.updated.Wait()
	return c.latest.Load().(string)
}

func (c *ReleaseChecker) update(ctx context.Context) error {
	env, err := c.storage.GetEnvConfig(ctx)
	if err == storage.ErrNotFound {
		return nil
	} else if err != nil {
		return err
	}

	var latest string
	if env == neco.StagingEnv {
		latest, err = c.github.GetLatestPreReleaseTag(ctx)
	} else if env == neco.ProdEnv {
		latest, err = c.github.GetLatestReleaseTag(ctx)
	} else {
		log.Warn("Unknown env: "+env, map[string]interface{}{})
		c.latest.Store("")
		return nil
	}
	if err == ErrNoReleases {
		c.notfound = true
		return nil
	}
	if err != nil {
		return err
	}

	current, err := c.pkg.GetVersion(ctx, "neco")
	if err != nil {
		return err
	}

	latestVer, err := version.NewVersion(latest)
	if err != nil {
		return err
	}

	currentVer, err := version.NewVersion(current)
	if err != nil {
		return err
	}

	if !latestVer.GreaterThan(currentVer) {
		return nil
	}

	log.Warn("New neco release is found ", map[string]interface{}{
		"env":     env,
		"version": latest,
	})
	c.latest.Store(latest)
	return nil
}
