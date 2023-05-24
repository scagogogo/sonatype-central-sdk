package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByTag(t *testing.T) {
	tag, err := SearchByTag(context.Background(), "scalaVersion-2.9", -1)
	assert.Nil(t, err)
	assert.True(t, len(tag) > 0)
}
