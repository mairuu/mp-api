package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectMatch(t *testing.T) {
	assert.True(t, subjectMatch("*", "user"))
	assert.True(t, subjectMatch("user", "user"))
	assert.False(t, subjectMatch("user", "admin"))
}

func TestScopeMatch(t *testing.T) {
	assert.True(t, scopeMatch("manga:*", "manga:owner"))
	assert.True(t, scopeMatch("manga:owner", "manga:owner"))
	assert.False(t, scopeMatch("manga:owner", "manga:other"))
	assert.False(t, scopeMatch("chapter:*", "manga:owner"))
}

func TestActionMatch(t *testing.T) {
	assert.True(t, actionMatch("read", "read"))
	assert.True(t, actionMatch("*", "read"))
	assert.False(t, actionMatch("write", "read"))
	assert.False(t, actionMatch("read", "write"))
}
