package repository

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

type NamespacedNameCacheKey struct {
}

func (k *NamespacedNameCacheKey) Marshal(key types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", key.Namespace, key.Name)
}

func (k *NamespacedNameCacheKey) Unmarshal(data string) (types.NamespacedName, error) {
	var key types.NamespacedName
	nameComponents := strings.Split(data, "/")
	if len(nameComponents) != 2 {
		return key, fmt.Errorf("invalid data format")
	}
	key.Namespace = nameComponents[0]
	key.Name = nameComponents[1]
	return key, nil
}

type RepositoryEventType string

const (
	RepositoryEventCreated RepositoryEventType = "created"
	RepositoryEventUpdated RepositoryEventType = "updated"
	RepositoryEventDeleted RepositoryEventType = "deleted"
)

type RepositoryEvent[K any] struct {
	Type    RepositoryEventType
	Name    types.NamespacedName
	Element *K
}

type RepositoryEventListener[K any] interface {
	Handle(event *RepositoryEvent[K])
}
