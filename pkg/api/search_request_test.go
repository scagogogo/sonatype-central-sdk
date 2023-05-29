package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchRequest(t *testing.T) {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId("net.virtual-void"))
	r, err := SearchRequest[*response.Artifact](context.Background(), search)
	assert.Nil(t, err)
	assert.NotNil(t, r)
}
