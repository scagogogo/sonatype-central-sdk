package api

import (
	"context"
	"errors"
	"strings"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// 本文件包含与全限定类名相关的搜索方法

// SearchByFullyQualifiedClassName 根据全限定类名搜索制品
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - fullyQualifiedClassName: 全限定类名，如"org.apache.commons.lang3.StringUtils"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchByFullyQualifiedClassName(ctx context.Context, fullyQualifiedClassName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByFullyQualifiedClassName(ctx, fullyQualifiedClassName).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName)).SetLimit(limit)
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

// IteratorByFullyQualifiedClassName 返回一个全限定类名搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - fullyQualifiedClassName: 全限定类名，如"org.apache.commons.lang3.StringUtils"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByFullyQualifiedClassName(ctx context.Context, fullyQualifiedClassName string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchByPackageAndClassName 根据包名和类名组合搜索
// 例如，包名为"org.apache.commons.lang3"和类名为"StringUtils"将查找该包中的StringUtils类
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - packageName: 包名，如"org.apache.commons.lang3"
//   - className: 类名，如"StringUtils"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchByPackageAndClassName(ctx context.Context, packageName, className string, limit int) ([]*response.Version, error) {
	// 组合成全限定类名
	fullyQualifiedClassName := packageName
	if !strings.HasSuffix(fullyQualifiedClassName, ".") {
		fullyQualifiedClassName += "."
	}
	fullyQualifiedClassName += className

	return c.SearchByFullyQualifiedClassName(ctx, fullyQualifiedClassName, limit)
}

// IteratorByPackageAndClassName 返回一个包名和类名组合搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - packageName: 包名，如"org.apache.commons.lang3"
//   - className: 类名，如"StringUtils"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByPackageAndClassName(ctx context.Context, packageName, className string) *SearchIterator[*response.Version] {
	// 组合成全限定类名
	fullyQualifiedClassName := packageName
	if !strings.HasSuffix(fullyQualifiedClassName, ".") {
		fullyQualifiedClassName += "."
	}
	fullyQualifiedClassName += className

	return c.IteratorByFullyQualifiedClassName(ctx, fullyQualifiedClassName)
}

// SearchByJavaPackage 根据Java包名搜索制品
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - packageName: 包名，如"org.apache.commons.lang3"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchByJavaPackage(ctx context.Context, packageName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByJavaPackage(ctx, packageName).ToSlice()
	} else {
		// Maven Central API并不直接支持按包名搜索
		// 我们使用多种查询方式提高匹配成功率

		// 方法1: 使用通配符搜索包内所有类
		wildcardQuery := packageName + ".*"
		search1 := request.NewSearchRequest().SetQuery(request.NewQuery().SetFullyQualifiedClassName(wildcardQuery)).SetLimit(limit)
		result1, err1 := SearchRequestJsonDoc[*response.Version](c, ctx, search1)

		// 如果方法1成功，直接返回结果
		if err1 == nil && result1 != nil && result1.ResponseBody != nil && len(result1.ResponseBody.Docs) > 0 {
			return result1.ResponseBody.Docs, nil
		}

		// 方法2: 使用groupId匹配（有些包名与groupId可能匹配）
		groupParts := strings.Split(packageName, ".")
		if len(groupParts) >= 2 {
			potentialGroupId := strings.Join(groupParts[:2], ".")
			search2 := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(potentialGroupId)).SetLimit(limit)
			result2, err2 := SearchRequestJsonDoc[*response.Version](c, ctx, search2)

			if err2 == nil && result2 != nil && result2.ResponseBody != nil && len(result2.ResponseBody.Docs) > 0 {
				return result2.ResponseBody.Docs, nil
			}
		}

		// 方法3: 使用自定义查询包含文本搜索
		customQuery := packageName
		search3 := request.NewSearchRequest().SetQuery(request.NewQuery().SetCustomQuery(customQuery)).SetLimit(limit)
		result3, err3 := SearchRequestJsonDoc[*response.Version](c, ctx, search3)

		if err3 == nil && result3 != nil && result3.ResponseBody != nil && len(result3.ResponseBody.Docs) > 0 {
			return result3.ResponseBody.Docs, nil
		}

		// 如果所有方法都失败，返回最后一个尝试的结果
		if result3 != nil && result3.ResponseBody != nil {
			return result3.ResponseBody.Docs, nil
		}

		// 如果没有任何结果，返回空切片和一个解释性错误
		return nil, errors.New("无法找到匹配包名的制品，Maven Central可能不支持直接按包名搜索")
	}
}

// IteratorByJavaPackage 返回一个Java包名搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - packageName: 包名，如"org.apache.commons.lang3"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByJavaPackage(ctx context.Context, packageName string) *SearchIterator[*response.Version] {
	// 使用改进的搜索方式 - 按包名+通配符搜索
	wildcardQuery := packageName + ".*"
	query := request.NewQuery().SetFullyQualifiedClassName(wildcardQuery)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Version](search).WithClient(c)
}
