package v1alpha1

import (
corev1 "k8s.io/api/core/v1"
"k8s.io/apimachinery/pkg/api/resource"
)

func (s *TypesenseClusterSpec) GetImagePullSecrets() []corev1.LocalObjectReference {
if s.ImagePullSecrets != nil {
return s.ImagePullSecrets
}

return []corev1.LocalObjectReference{}
}

func (s *TypesenseClusterSpec) GetPriorityClassName() string {
if s.PriorityClassName != nil {
return *s.PriorityClassName
}

return ""
}

func (s *TypesenseClusterSpec) GetStatefulSetAnnotations() map[string]string {
if s.StatefulSetAnnotations != nil {
return s.StatefulSetAnnotations
}

return map[string]string{}
}

func (s *TypesenseClusterSpec) GetTopologySpreadConstraints() []corev1.TopologySpreadConstraint {
tscs := make([]corev1.TopologySpreadConstraint, 0)

if s.TopologySpreadConstraints != nil {
for _, tsc := range s.TopologySpreadConstraints {
tscs = append(tscs, tsc)
}
}

return tscs
}

func (s *TypesenseClusterSpec) GetMetricsExporterSpecs() MetricsExporterSpec {
if s.Metrics != nil {
return *s.Metrics
}

return MetricsExporterSpec{
Release:           "promstack",
Image:             "akyriako78/typesense-prometheus-exporter:0.1.9",
IntervalInSeconds: 15,
}
}

func (s *TypesenseClusterSpec) GetMetricsExporterResources() corev1.ResourceRequirements {
if s.Metrics != nil && s.Metrics.Resources != nil {
return *s.Metrics.Resources
}

return corev1.ResourceRequirements{
Limits: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("100m"),
corev1.ResourceMemory: resource.MustParse("64Mi"),
},
Requests: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("100m"),
corev1.ResourceMemory: resource.MustParse("32Mi"),
},
}
}

func (s *TypesenseClusterSpec) GetPodManagementPolicy() string {
if s.PodManagementPolicy != nil {
return *s.PodManagementPolicy
}
return "Parallel"
}

func (s *TypesenseClusterSpec) GetTerminationGracePeriodSeconds() int64 {
if s.TerminationGracePeriodSeconds != nil {
return *s.TerminationGracePeriodSeconds
}
return 5
}

func (s *TypesenseClusterSpec) GetHealthCheckSidecarSpecs() HealthCheckSpec {
if s.HealthCheck != nil {
return *s.HealthCheck
}

return HealthCheckSpec{
Image: "akyriako78/typesense-healthcheck:0.1.8",
}
}

func (s *TypesenseClusterSpec) GetHealthCheckSidecarResources() corev1.ResourceRequirements {
if s.HealthCheck != nil && s.HealthCheck.Resources != nil {
return *s.HealthCheck.Resources
}

return corev1.ResourceRequirements{
Limits: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("100m"),
corev1.ResourceMemory: resource.MustParse("64Mi"),
},
Requests: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("100m"),
corev1.ResourceMemory: resource.MustParse("32Mi"),
},
}
}

func (s *IngressSpec) GetReverseProxyResources() corev1.ResourceRequirements {
if s.Resources != nil {
return *s.Resources
}

return corev1.ResourceRequirements{
Limits: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("150m"),
corev1.ResourceMemory: resource.MustParse("64Mi"),
},
Requests: corev1.ResourceList{
corev1.ResourceCPU:    resource.MustParse("100m"),
corev1.ResourceMemory: resource.MustParse("32Mi"),
},
}
}
