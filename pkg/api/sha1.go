package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func (c *Client) SearchBySha1(ctx context.Context, sha1 string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorBySha1(ctx, sha1).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1)).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

func (c *Client) IteratorBySha1(ctx context.Context, sha1 string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// GetFirstBySha1 返回与给定SHA1匹配的第一个版本信息，如果不存在则返回nil
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1: 要搜索的SHA1哈希值
//
// 返回:
//   - *response.Version: 找到的第一个版本，如果未找到则为nil
//   - error: 如果搜索过程中发生错误
func (c *Client) GetFirstBySha1(ctx context.Context, sha1 string) (*response.Version, error) {
	results, err := c.SearchBySha1(ctx, sha1, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// ExistsSha1 检查是否存在具有给定SHA1的构件
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1: 要检查的SHA1哈希值
//
// 返回:
//   - bool: 如果存在匹配的构件则为true，否则为false
//   - error: 如果检查过程中发生错误
func (c *Client) ExistsSha1(ctx context.Context, sha1 string) (bool, error) {
	version, err := c.GetFirstBySha1(ctx, sha1)
	if err != nil {
		return false, err
	}
	return version != nil, nil
}

// SearchExactSha1 执行精确的SHA1搜索，使用自定义查询语法
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1: 要搜索的SHA1哈希值
//
// 返回:
//   - []*response.Version: 与SHA1完全匹配的版本列表
//   - error: 如果搜索过程中发生错误
func (c *Client) SearchExactSha1(ctx context.Context, sha1 string) ([]*response.Version, error) {
	// 直接使用SHA1查询，Maven Central API会执行精确匹配
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1))
	// 添加自定义参数以确保精确匹配
	search.AddCustomParam("exact", "true")

	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// CountBySha1 计算匹配给定SHA1哈希的构件数量
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1: 要搜索的SHA1哈希值
//
// 返回:
//   - int: 匹配的构件数量
//   - error: 如果计数过程中发生错误
func (c *Client) CountBySha1(ctx context.Context, sha1 string) (int, error) {
	// 设置limit为0表示我们只关心总数而不需要实际返回数据
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1)).SetLimit(0)
	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return 0, err
	}
	if result == nil || result.ResponseBody == nil {
		return 0, errors.New("empty response body")
	}
	return result.ResponseBody.NumFound, nil
}

// SearchBySha1Prefix 使用SHA1前缀进行模糊搜索
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1Prefix: SHA1哈希的前缀（可以是任意长度）
//   - limit: 最大返回结果数，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Version: 与SHA1前缀匹配的版本列表
//   - error: 如果搜索过程中发生错误
func (c *Client) SearchBySha1Prefix(ctx context.Context, sha1Prefix string, limit int) ([]*response.Version, error) {
	if len(sha1Prefix) == 0 {
		return nil, errors.New("SHA1前缀不能为空")
	}

	if len(sha1Prefix) == 40 {
		// 如果提供了完整的SHA1，使用精确搜索
		return c.SearchBySha1(ctx, sha1Prefix, limit)
	}

	if limit <= 0 {
		return c.IteratorBySha1Prefix(ctx, sha1Prefix).ToSlice()
	}

	// 使用自定义查询构建SHA1前缀搜索
	customQuery := "1:" + sha1Prefix + "*"
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetCustomQuery(customQuery)).SetLimit(limit)

	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// IteratorBySha1Prefix 使用SHA1前缀进行模糊搜索，返回迭代器
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - sha1Prefix: SHA1哈希的前缀（可以是任意长度）
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器
func (c *Client) IteratorBySha1Prefix(ctx context.Context, sha1Prefix string) *SearchIterator[*response.Version] {
	if len(sha1Prefix) == 40 {
		// 如果提供了完整的SHA1，使用精确搜索
		return c.IteratorBySha1(ctx, sha1Prefix)
	}

	// 使用自定义查询构建SHA1前缀搜索
	customQuery := "1:" + sha1Prefix + "*"
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetCustomQuery(customQuery))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}
