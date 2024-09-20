package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestNamespacedNameCacheKey(t *testing.T) {

	key := &NamespacedNameCacheKey{}
	data := key.Marshal(types.NamespacedName{Namespace: "foo", Name: "bar"})
	assert.Equal(t, "foo/bar", data)

	namespacedName, err := key.Unmarshal(data)
	assert.Nil(t, err)
	assert.Equal(t, "foo", namespacedName.Namespace)
	assert.Equal(t, "bar", namespacedName.Name)

}
