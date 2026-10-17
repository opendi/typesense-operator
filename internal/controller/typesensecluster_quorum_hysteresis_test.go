package controller

import (
	"testing"
	"time"
)

const testKey = "typesense-system/typesense-global-2"

// TestDeadlockGracePeriodPreventsImmediateDowngrade documents the regression:
// a transient ElectionDeadlock during pod churn must not downgrade on the
// first observation.
func TestDeadlockGracePeriodPreventsImmediateDowngrade(t *testing.T) {
	grace := 3 * time.Minute
	obs := newDeadlockObservations(grace)
	start := time.Now()

	if obs.observe(testKey, start) {
		t.Fatal("first deadlock observation must not allow a downgrade")
	}

	if obs.observe(testKey, start.Add(grace-time.Second)) {
		t.Fatalf("downgrade allowed before grace period elapsed")
	}

	if !obs.observe(testKey, start.Add(grace)) {
		t.Fatal("downgrade not allowed once grace period elapsed")
	}

	if !obs.observe(testKey, start.Add(grace+time.Hour)) {
		t.Fatal("downgrade not allowed after grace period elapsed")
	}
}

// TestDeadlockObservationsArePerCluster ensures one cluster's deadlock does not
// affect another: cluster b expiring must not let cluster a downgrade early.
func TestDeadlockObservationsArePerCluster(t *testing.T) {
	grace := time.Minute
	obs := newDeadlockObservations(grace)
	start := time.Now()

	if obs.observe("ns/a", start) {
		t.Fatal("first observation for cluster a must not downgrade")
	}

	// Drive cluster b past its grace period; this must say nothing about a.
	expiredB := false
	for i := 1; i <= 3; i++ {
		if obs.observe("ns/b", start.Add(time.Duration(i)*grace)) {
			expiredB = true
			break
		}
	}
	if !expiredB {
		t.Fatal("cluster b should downgrade after its own grace period")
	}

	// Cluster a has been observed once, only just past its own first sighting.
	if obs.observe("ns/a", start.Add(time.Second)) {
		t.Fatal("cluster a must not downgrade because cluster b expired")
	}
}

// TestDeadlockResetRestartsGracePeriod covers the case where the cluster
// recovers and later deadlocks again: the grace period must restart.
func TestDeadlockResetRestartsGracePeriod(t *testing.T) {
	obs := newDeadlockObservations(time.Minute)
	start := time.Now()

	if obs.observe(testKey, start) {
		t.Fatal("first observation must not downgrade")
	}

	obs.reset(testKey)

	if obs.observe(testKey, start.Add(time.Hour)) {
		t.Fatal("after reset the grace period must restart from the new time")
	}

	if !obs.observe(testKey, start.Add(time.Hour+time.Minute)) {
		t.Fatal("grace period should elapse relative to the reset observation")
	}
}

// TestObserveIsConcurrencySafe guards the mutex, since reconciles can overlap.
func TestObserveIsConcurrencySafe(t *testing.T) {
	obs := newDeadlockObservations(time.Millisecond)
	start := time.Now()

	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 200; j++ {
				obs.observe(testKey, start)
				obs.reset(testKey)
			}
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
}
