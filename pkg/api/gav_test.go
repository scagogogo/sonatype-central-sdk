package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByGAV(t *testing.T) {
	gav, err := SearchByGAV(context.Background(), "com.google.inject", "guice", "3.0")
	assert.Nil(t, err)
	assert.True(t, len(gav) > 0)
}
