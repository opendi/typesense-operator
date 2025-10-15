package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"

	tsv1alpha1 "github.com/akyriako/typesense-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	letters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	debugLevel = 1
)

func generateToken() (string, error) {
	token := make([]byte, 256)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	base64EncodedToken := base64.StdEncoding.EncodeToString(token)
	return base64EncodedToken, nil
}

func generateSecureRandomString(length int) (string, error) {
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}

	for i := range result {
		result[i] = letters[int(result[i])%len(letters)]
	}
	return string(result), nil
}

func getLabels(ts *tsv1alpha1.TypesenseCluster) map[string]string {
	return map[string]string{
		"app": fmt.Sprintf(ClusterAppLabel, ts.Name),
	}
}

func getObjectMeta(ts *tsv1alpha1.TypesenseCluster, name *string, annotations map[string]string) metav1.ObjectMeta {
	if name == nil {
		name = &ts.Name
	}

	return metav1.ObjectMeta{
		Name:        *name,
		Namespace:   ts.Namespace,
		Labels:      getLabels(ts),
		Annotations: annotations,
	}
}

func getReverseProxyLabels(ts *tsv1alpha1.TypesenseCluster) map[string]string {
	return map[string]string{
		"app": fmt.Sprintf(ClusterReverseProxyAppLabel, ts.Name),
	}
}

func getReverseProxyObjectMeta(ts *tsv1alpha1.TypesenseCluster, name *string, annotations map[string]string) metav1.ObjectMeta {
	if name == nil {
		name = &ts.Name
	}

	return metav1.ObjectMeta{
		Name:        *name,
		Namespace:   ts.Namespace,
		Labels:      getReverseProxyLabels(ts),
		Annotations: annotations,
	}
}

func getPodMonitorLabels(ts *tsv1alpha1.TypesenseCluster) map[string]string {
	return map[string]string{
		"app": fmt.Sprintf(ClusterMetricsPodMonitorAppLabel, ts.Name),
	}
}

func getPodMonitorObjectMeta(ts *tsv1alpha1.TypesenseCluster, name *string, annotations map[string]string) metav1.ObjectMeta {
	if name == nil {
		name = &ts.Name
	}

	return metav1.ObjectMeta{
		Name:        *name,
		Namespace:   ts.Namespace,
		Labels:      getPodMonitorLabels(ts),
		Annotations: annotations,
	}
}

const (
	minDelayPerReplicaFactor = 1
	maxDelayPerReplicaFactor = 3
)

func getDelayPerReplicaFactor(size int) int64 {
	if size != 0 {
		if size <= maxDelayPerReplicaFactor {
			return int64(size)
		} else {
			return maxDelayPerReplicaFactor
		}
	}
	return minDelayPerReplicaFactor
}

func contains(values []string, value string) (int, bool) {
	//sort.Strings(values)

	for i, v := range values {
		if v == value {
			return i, true
		}
	}

	return -1, false
}

func normalizeVolumes(vols []corev1.Volume) []corev1.Volume {
	if vols == nil {
		vols = []corev1.Volume{}
	}

	vcopy := append([]corev1.Volume(nil), vols...)
	for i := range vcopy {
		if cm := vcopy[i].VolumeSource.ConfigMap; cm != nil {
			cm.DefaultMode = nil
		}
	}

	sort.Slice(vcopy, func(i, j int) bool {
		return vcopy[i].Name < vcopy[j].Name
	})

	return vcopy
}

func normalizeVolumeMounts(mounts []corev1.VolumeMount) []corev1.VolumeMount {
	if mounts == nil {
		mounts = []corev1.VolumeMount{}
	}
	copyMounts := append([]corev1.VolumeMount(nil), mounts...)
	sort.Slice(copyMounts, func(i, j int) bool {
		return copyMounts[i].Name < copyMounts[j].Name
	})
	return copyMounts
}

// needsSyncVolumes returns true if the desired vols differ from what's in the pod.
func needsSyncVolumes(desired, existing []corev1.Volume) bool {
	return !equality.Semantic.DeepEqual(
		normalizeVolumes(desired),
		normalizeVolumes(existing),
	)
}

// needsSyncMounts returns true if the desired mounts differ from what's in the container.
func needsSyncMounts(desired, existing []corev1.VolumeMount) bool {
	return !equality.Semantic.DeepEqual(
		normalizeVolumeMounts(desired),
		normalizeVolumeMounts(existing),
	)
}

var ip4Prefix = regexp.MustCompile(
	`^((25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}` +
		`(25[0-5]|2[0-4]\d|[01]?\d?\d)`,
)

func hasIP4Prefix(s string) bool {
	return ip4Prefix.MatchString(s)
}
