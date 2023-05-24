package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownload(t *testing.T) {
	download, err := Download(context.Background(), "com/jolira/guice/3.0.0/guice-3.0.0.pom")
	assert.Nil(t, err)
	assert.NotEmpty(t, download)
}
