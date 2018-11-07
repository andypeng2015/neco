package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/neco"
	"github.com/cybozu-go/neco/storage"
)

func proxyHTTPClient(proxyURL *url.URL) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   1 * time.Hour,
	}
}

func localHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Minute,
	}
}

// Worker implements Neco auto update worker process.
// This is a state machine.
type Worker struct {
	mylrn    int
	version  string
	ec       *clientv3.Client
	storage  storage.Storage
	operator Operator

	// internal states
	req     *neco.UpdateRequest
	step    int
	barrier Barrier
}

// NewWorker returns a *Worker.
func NewWorker(ec *clientv3.Client, op Operator, version string, mylrn int) *Worker {
	return &Worker{
		mylrn:    mylrn,
		version:  version,
		ec:       ec,
		storage:  storage.NewStorage(ec),
		operator: op,
	}
}

// Run waits for update request from neco-updater, then executes
// update process with other workers.  To communicate with neco-updater
// and other workers, etcd objects are used.
//
// Run works as follows:
//
// 1. Check the current request.  If the request is not found, go to 5.
//
// 2. If locally installed neco package is older than the requested version,
//    neco-worker updates the package, then exits to be restarted by systemd.
//
// 3. Check the status of request and workers; if the update process was aborted,
//    or if the update process has completed successfully, also go to 5.
//
// 4. Update programs for the requested version.
//
// 5. Wait for the new request.    If there is a new one, neco-worker updates
//    the package and exists to be restarted by systemd.
func (w *Worker) Run(ctx context.Context) error {
	req, rev, err := w.storage.GetRequestWithRev(ctx)

	for {
		if err == storage.ErrNotFound {
			req, rev, err = w.waitRequest(ctx, rev)
			continue
		}
		if err != nil {
			return err
		}

		if req.Stop {
			req, rev, err = w.waitRequest(ctx, rev)
			continue
		}

		if !req.IsMember(w.mylrn) {
			req, rev, err = w.waitRequest(ctx, rev)
			continue
		}

		if w.version != req.Version {
			// After update of "neco" package, old neco-worker should stop.
			return w.operator.UpdateNeco(ctx, req)
		}

		stMap, err := w.storage.GetStatuses(ctx)
		if err != nil {
			return err
		}
		if neco.UpdateAborted(req.Version, w.mylrn, stMap) {
			log.Info("previous update was aborted", nil)
			req, rev, err = w.waitRequest(ctx, rev)
			continue
		}
		if neco.UpdateCompleted(req.Version, req.Servers, stMap) {
			log.Info("previous update was completed successfully", nil)
			req, rev, err = w.waitRequest(ctx, rev)
			continue
		}

		w.req = req
		log.Info("update starts", map[string]interface{}{
			"version": req.Version,
		})
		err = w.update(ctx, rev)
		if err != nil {
			return err
		}

		log.Info("update finished", map[string]interface{}{
			"version": req.Version,
		})
		req, rev, err = w.waitRequest(ctx, rev)
	}
}

func (w *Worker) update(ctx context.Context, rev int64) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	st := neco.UpdateStatus{
		Version: w.req.Version,
		Step:    1,
		Cond:    neco.CondRunning,
	}
	err := w.storage.PutStatus(ctx, w.mylrn, st)
	if err != nil {
		return err
	}
	w.step = 1
	w.barrier = NewBarrier(w.req.Servers)

	ch := w.ec.Watch(ctx, storage.KeyStatusPrefix,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev+1),
		clientv3.WithFilterDelete())
	for wr := range ch {
		err := wr.Err()
		if err != nil {
			return err
		}

		for _, ev := range wr.Events {
			completed, err := w.dispatch(ctx, ev)
			if err != nil {
				return err
			}
			if completed {
				return nil
			}
		}
	}

	return nil
}

func (w *Worker) dispatch(ctx context.Context, ev *clientv3.Event) (bool, error) {
	key := string(ev.Kv.Key[len(storage.KeyStatusPrefix):])
	if key == "current" {
		return w.handleCurrent(ctx, ev)
	}

	lrn, err := strconv.Atoi(string(ev.Kv.Key[len(storage.KeyWorkerStatusPrefix):]))
	if err != nil {
		return false, err
	}

	return w.handleWorkerStatus(ctx, lrn, ev)
}

func (w *Worker) handleCurrent(ctx context.Context, ev *clientv3.Event) (bool, error) {
	req := new(neco.UpdateRequest)
	err := json.Unmarshal(ev.Kv.Value, req)
	if err != nil {
		return false, err
	}

	if req.Stop {
		log.Warn("request was canceled", map[string]interface{}{
			"version": req.Version,
		})
		return true, nil
	}

	log.Error("unexpected request", map[string]interface{}{
		"version":    req.Version,
		"servers":    req.Servers,
		"stop":       req.Stop,
		"started_at": req.StartedAt,
	})
	return false, errors.New("unexpected request")
}

func (w *Worker) handleWorkerStatus(ctx context.Context, lrn int, ev *clientv3.Event) (bool, error) {
	if !w.req.IsMember(lrn) {
		log.Warn("ignoring unexpected boot server", map[string]interface{}{
			"lrn":     lrn,
			"version": w.req.Version,
			"servers": w.req.Servers,
		})
		return false, nil
	}

	st := new(neco.UpdateStatus)
	err := json.Unmarshal(ev.Kv.Value, st)
	if err != nil {
		return false, err
	}
	if st.Version != w.req.Version {
		return false, errors.New("unexpected version in worker status: " + st.Version)
	}
	if st.Step != w.step {
		return false, fmt.Errorf("unexpected step in worker status: %d", st.Step)
	}
	if st.Cond == neco.CondAbort {
		return false, fmt.Errorf("other boot server failed to update: %d", lrn)
	}

	if w.barrier.Check(lrn) {
		return w.runStep(ctx)
	}

	return false, nil
}

func (w *Worker) runStep(ctx context.Context) (bool, error) {
	err := w.operator.RunStep(ctx, w.req, w.step)

	if err != nil {
		log.Error("update failed", map[string]interface{}{
			"version":   w.req.Version,
			"step":      w.step,
			log.FnError: err,
		})
		st := neco.UpdateStatus{
			Version: w.req.Version,
			Step:    w.step,
			Cond:    neco.CondAbort,
			Message: err.Error(),
		}

		err2 := w.storage.PutStatus(ctx, w.mylrn, st)
		if err2 != nil {
			log.Warn("failed to put status", map[string]interface{}{
				log.FnError: err2.Error(),
			})
		}

		return false, err
	}

	if w.step != w.operator.FinalStep() {
		w.step++
		w.barrier = NewBarrier(w.req.Servers)
		st := neco.UpdateStatus{
			Version: w.req.Version,
			Step:    w.step,
			Cond:    neco.CondRunning,
		}
		err = w.storage.PutStatus(ctx, w.mylrn, st)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	st := neco.UpdateStatus{
		Version: w.req.Version,
		Step:    w.step,
		Cond:    neco.CondComplete,
	}
	err = w.storage.PutStatus(ctx, w.mylrn, st)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (w *Worker) waitRequest(ctx context.Context, rev int64) (*neco.UpdateRequest, int64, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := w.ec.Watch(ctx, storage.KeyCurrent,
		clientv3.WithRev(rev+1),
		clientv3.WithFilterDelete())
	for wr := range ch {
		err := wr.Err()
		if err != nil {
			return nil, 0, err
		}

		if len(wr.Events) == 0 {
			continue
		}

		ev := wr.Events[0]
		req := new(neco.UpdateRequest)
		err = json.Unmarshal(ev.Kv.Value, req)
		if err != nil {
			return nil, 0, err
		}

		return req, ev.Kv.ModRevision, nil
	}

	return nil, 0, errors.New("waitRequest was interrupted")
}
