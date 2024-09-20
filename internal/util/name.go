package util

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
)

func GetNamespacedNameString(name types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", name.Namespace, name.Name)
}

func GetNamespacedName(name string) (types.NamespacedName, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		log.Error().Msg("invalid namespaced name")
		return types.NamespacedName{}, errors.New("invalid namespaced name")
	}
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

func GetConfigMapNamespacedNameString(configMap *corev1.ConfigMap) string {
	return GetNamespacedNameString(types.NamespacedName{
		Namespace: configMap.Namespace,
		Name:      configMap.Name,
	})
}
