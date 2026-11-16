package controller

import (
	"fmt"
	"sync"
	"time"

	tsv1alpha1 "github.com/akyriako/typesense-operator/api/v1alpha1"
)

// quorumDowngradeGracePeriod is how long an ElectionDeadlock must persist
// before downgradeQuorum is allowed to scale the StatefulSet down to 1 replica.
//
// A transient deadlock is normal during pod churn: when a pod is rescheduled its
// IP changes, the nodes ConfigMap is rewritten and every pod is force-patched,
// which briefly leaves the cluster with no leader. That is self-healing and must
// not be met with a destructive downgrade. Only a deadlock that outlasts this
// grace period is treated as genuine.
const quorumDowngradeGracePeriod = 3 * time.Minute

// deadlockObservations tracks, per cluster, when an ElectionDeadlock was first
// observed. The operator runs as a single leader-elected instance, so in-memory
// state is sufficient; losing it merely means the grace period restarts.
type deadlockObservations struct {
	mu        sync.Mutex
	grace     time.Duration
	firstSeen map[string]time.Time
}

func newDeadlockObservations(grace time.Duration) *deadlockObservations {
	return &deadlockObservations{
		grace:     grace,
		firstSeen: make(map[string]time.Time),
	}
}

// observe records an ElectionDeadlock for the cluster at the given time and
// reports whether the grace period has now elapsed, i.e. a downgrade may proceed.
func (d *deadlockObservations) observe(key string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	first, seen := d.firstSeen[key]
	if !seen {
		d.firstSeen[key] = now
		return false
	}

	return now.Sub(first) >= d.grace
}

// reset clears any recorded observation for the cluster.
func (d *deadlockObservations) reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.firstSeen, key)
}

// recordDeadlock records an ElectionDeadlock observation and reports whether the
// grace period has elapsed, i.e. a destructive downgrade may proceed.
func (r *TypesenseClusterReconciler) recordDeadlock(key string) bool {
	if r.deadlockObservations == nil {
		return true
	}

	return r.deadlockObservations.observe(key, time.Now())
}

// clearDeadlock forgets any recorded ElectionDeadlock observation.
func (r *TypesenseClusterReconciler) clearDeadlock(key string) {
	if r.deadlockObservations == nil {
		return
	}

	r.deadlockObservations.reset(key)
}

// quorumKey identifies a cluster for deadlock-observation tracking.
func quorumKey(ts *tsv1alpha1.TypesenseCluster) string {
	return fmt.Sprintf("%s/%s", ts.Namespace, ts.Name)
}
