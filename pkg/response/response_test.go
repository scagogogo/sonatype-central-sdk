package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWithHighlighting(t *testing.T) {
	// 构造一个包含高亮信息的JSON响应
	jsonData := `{
		"responseHeader": {
			"status": 0,
			"QTime": 617,
			"params": {
				"q": "fc:org.specs.runner.JUnit",
				"hl.snippets": "3",
				"hl": "true",
				"indent": "off",
				"fl": "id,g,a,v,p,ec,timestamp,tags",
				"hl.fl": "fch",
				"sort": "score desc,timestamp desc,g asc,a asc,v desc",
				"rows": "20",
				"wt": "json",
				"version": "2.2"
			}
		},
		"response": {
			"numFound": 54,
			"start": 0,
			"docs": [
				{
					"id": "org.specs:specs:1.2.3",
					"g": "org.specs",
					"a": "specs",
					"v": "1.2.3",
					"p": "jar",
					"timestamp": 1227569516000,
					"ec": ["-sources.jar", ".jar", "-tests.jar", ".pom"],
					"tags": ["behaviour", "driven", "framework", "design", "specs"]
				},
				{
					"id": "org.specs:specs:1.2.4",
					"g": "org.specs",
					"a": "specs",
					"v": "1.2.4",
					"p": "jar",
					"timestamp": 1227569513000,
					"ec": ["-sources.jar", ".jar", "-tests.jar", ".pom"],
					"tags": ["behaviour", "driven", "framework", "design", "specs"]
				}
			]
		},
		"highlighting": {
			"org.specs:specs:1.2.3": {
				"fch": ["<em>org</em>.<em>specs</em>.<em>runner</em>.<em>JUnit</em>"]
			},
			"org.specs:specs:1.2.4": {
				"fch": ["<em>org</em>.<em>specs</em>.<em>runner</em>.<em>JUnit</em>"]
			}
		}
	}`

	// 解析JSON到Response结构体
	var response Response[*Version]
	err := json.Unmarshal([]byte(jsonData), &response)

	// 验证解析是否成功
	assert.NoError(t, err)
	assert.NotNil(t, response.ResponseHeader)
	assert.NotNil(t, response.ResponseBody)
	assert.NotNil(t, response.Highlighting)

	// 验证响应头信息
	assert.Equal(t, 0, response.ResponseHeader.Status)
	assert.Equal(t, 617, response.ResponseHeader.QTime)

	// 验证响应体信息
	assert.Equal(t, 54, response.ResponseBody.NumFound)
	assert.Equal(t, 0, response.ResponseBody.Start)
	assert.Equal(t, 2, len(response.ResponseBody.Docs))

	// 验证文档内容
	assert.Equal(t, "org.specs:specs:1.2.3", response.ResponseBody.Docs[0].ID)
	assert.Equal(t, "org.specs", response.ResponseBody.Docs[0].GroupId)
	assert.Equal(t, "specs", response.ResponseBody.Docs[0].ArtifactId)
	assert.Equal(t, "1.2.3", response.ResponseBody.Docs[0].Version)

	// 验证高亮信息
	assert.Equal(t, 2, len(response.Highlighting))

	// 检查第一个文档的高亮
	highlight1, exists := response.Highlighting["org.specs:specs:1.2.3"]
	assert.True(t, exists)
	assert.Contains(t, highlight1, "fch")
	assert.Equal(t, 1, len(highlight1["fch"]))
	assert.Equal(t, "<em>org</em>.<em>specs</em>.<em>runner</em>.<em>JUnit</em>", highlight1["fch"][0])

	// 检查第二个文档的高亮
	highlight2, exists := response.Highlighting["org.specs:specs:1.2.4"]
	assert.True(t, exists)
	assert.Contains(t, highlight2, "fch")
	assert.Equal(t, 1, len(highlight2["fch"]))
	assert.Equal(t, "<em>org</em>.<em>specs</em>.<em>runner</em>.<em>JUnit</em>", highlight2["fch"][0])
}

func TestFacetCounts(t *testing.T) {
	// 构造一个包含聚合信息的JSON响应
	jsonData := `{
		"responseHeader": {
			"status": 0,
			"QTime": 10,
			"params": {
				"q": "*:*",
				"facet": "true",
				"facet.field": ["g", "a"]
			}
		},
		"response": {
			"numFound": 100,
			"start": 0,
			"docs": []
		},
		"facet_counts": {
			"facet_fields": {
				"g": ["org.apache", 30, "com.google", 25],
				"a": ["commons-io", 15, "guava", 10]
			}
		}
	}`

	// 解析JSON到Response结构体
	var response Response[*Artifact]
	err := json.Unmarshal([]byte(jsonData), &response)

	// 验证解析是否成功
	assert.NoError(t, err)
	assert.NotNil(t, response.ResponseHeader)
	assert.NotNil(t, response.ResponseBody)
	assert.NotNil(t, response.FacetCounts)

	// 验证聚合信息
	assert.NotNil(t, response.FacetCounts.FacetFields)
	assert.Equal(t, 2, len(response.FacetCounts.FacetFields))

	// 验证groupId聚合
	gField, exists := response.FacetCounts.FacetFields["g"]
	assert.True(t, exists)
	assert.Equal(t, 4, len(gField))

	// 验证artifactId聚合
	aField, exists := response.FacetCounts.FacetFields["a"]
	assert.True(t, exists)
	assert.Equal(t, 4, len(aField))
}
