package api

import (
	"context"
	"fmt"
	"github.com/crawler-go-go-go/go-requests"
	"strings"
)

// Download 从Maven中央仓库下载文件
// filepath比如： com/jolira/guice/3.0.0/guice-3.0.0.pom
func Download(ctx context.Context, filepath string) ([]byte, error) {
	targetUrl := "https://search.maven.org/remotecontent?filepath=" + filepath
	return requests.GetBytes(ctx, targetUrl)
}

func DownloadPom(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s/%s/%s-%s.pom", strings.ReplaceAll(groupId, ".", "/"), artifactId, version, artifactId, version)
	return Download(ctx, path)
}
