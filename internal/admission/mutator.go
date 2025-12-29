package admission

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/config"
	corev1 "k8s.io/api/core/v1"
)

type Mutator struct {
	ndotsValue        string
	annotationChecker *AnnotationChecker
	namespaceFilter   *NamespaceFilter
	logger            *slog.Logger
}

func NewMutator(cfg *config.Config, logger *slog.Logger) *Mutator {
	return &Mutator{
		ndotsValue:        strconv.Itoa(cfg.NdotsValue),
		annotationChecker: NewAnnotationChecker(cfg.AnnotationKey, cfg.AnnotationMode),
		namespaceFilter:   NewNamespaceFilter(cfg.NamespaceInclude, cfg.NamespaceExclude, logger),
		logger:            logger,
	}
}

func (m *Mutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
	if !m.namespaceFilter.ShouldMutate(pod.Namespace) {
		m.logger.Debug("skipping mutation due to namespace filter",
			"namespace", pod.Namespace,
			"name", pod.Name,
		)
		return nil, nil
	}

	if !m.annotationChecker.ShouldMutate(pod.Annotations) {
		m.logger.Debug("skipping mutation due to annotation",
			"namespace", pod.Namespace,
			"name", pod.Name,
		)
		return nil, nil
	}

	if pod.Spec.DNSConfig == nil {
		return []PatchOperation{{
			Op:   "add",
			Path: "/spec/dnsConfig",
			Value: map[string]interface{}{
				"options": []map[string]interface{}{
					{"name": "ndots", "value": m.ndotsValue},
				},
			},
		}}, nil
	}

	if pod.Spec.DNSConfig.Options == nil {
		return []PatchOperation{{
			Op:   "add",
			Path: "/spec/dnsConfig/options",
			Value: []map[string]interface{}{
				{"name": "ndots", "value": m.ndotsValue},
			},
		}}, nil
	}

	idx := findNdotsIndex(pod.Spec.DNSConfig.Options)
	if idx == -1 {
		return []PatchOperation{{
			Op:   "add",
			Path: "/spec/dnsConfig/options/-",
			Value: map[string]interface{}{
				"name": "ndots", "value": m.ndotsValue,
			},
		}}, nil
	}

	// Check if update needed
	if pod.Spec.DNSConfig.Options[idx].Value != nil && *pod.Spec.DNSConfig.Options[idx].Value == m.ndotsValue {
		return nil, nil
	}

	return []PatchOperation{{
		Op:    "replace",
		Path:  fmt.Sprintf("/spec/dnsConfig/options/%d/value", idx),
		Value: m.ndotsValue,
	}}, nil
}

func findNdotsIndex(options []corev1.PodDNSConfigOption) int {
	for i, opt := range options {
		if opt.Name == "ndots" {
			return i
		}
	}
	return -1
}
