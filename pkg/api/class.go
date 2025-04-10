package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByClassName 根据类名搜索相关制品
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - class: 要搜索的类名（不含包名）
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchByClassName(ctx context.Context, class string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByClassName(ctx, class).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class)).SetLimit(limit)
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

// IteratorByClassName 返回一个类名搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - class: 要搜索的类名（不含包名）
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByClassName(ctx context.Context, class string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchClassesByMethod 搜索包含特定方法的类
// 注意：这个功能依赖于Maven索引中包含方法名信息，可能并非所有索引都支持
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - methodName: 方法名，如"toString"或"equals"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchClassesByMethod(ctx context.Context, methodName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByMethod(ctx, methodName).ToSlice()
	}

	// 使用自定义查询
	customQuery := "m:" + methodName

	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// IteratorByMethod 返回一个方法名搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - methodName: 方法名，如"toString"或"equals"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByMethod(ctx context.Context, methodName string) *SearchIterator[*response.Version] {
	customQuery := "m:" + methodName
	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchClassesWithClassHierarchy 搜索继承自特定基类的类
// 注意：这个功能可能需要进一步处理搜索结果以精确筛选符合继承关系的类
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - baseClassName: 基类名，如"Exception"或"AbstractList"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchClassesWithClassHierarchy(ctx context.Context, baseClassName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByClassHierarchy(ctx, baseClassName).ToSlice()
	}

	// Maven Central API可能不直接支持继承关系搜索，我们采用基于类名搜索+自定义过滤方案
	// 进行相关性搜索，找出可能相关的类

	// 首先使用普通类名搜索
	versions, err := c.SearchByClassName(ctx, baseClassName, limit*2) // 获取更多结果用于后续过滤
	if err != nil {
		return nil, err
	}

	// 这里我们直接返回结果，实际应用中可能需要额外处理来确定继承关系
	// 例如：可能需要下载JAR文件并解析class文件以验证继承关系

	// 限制返回数量
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}

	return versions, nil
}

// IteratorByClassHierarchy 返回一个继承关系搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - baseClassName: 基类名，如"Exception"或"AbstractList"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByClassHierarchy(ctx context.Context, baseClassName string) *SearchIterator[*response.Version] {
	// 使用简单的类名搜索来模拟继承关系搜索
	return c.IteratorByClassName(ctx, baseClassName)
}

// SearchInterfaceImplementations 尝试搜索指定接口的实现类
// 注意：这是一个近似搜索，可能需要进一步验证实现关系
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - interfaceName: 接口名，如"Listener"或"Handler"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchInterfaceImplementations(ctx context.Context, interfaceName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByInterfaceImplementation(ctx, interfaceName).ToSlice()
	}

	// 接口实现搜索策略：搜索类名+接口名组合
	// 例如：搜索MyListener接口的实现类，可以尝试搜索以"*Listener"结尾的类

	// 构造模式匹配搜索
	searchPattern := "*" + interfaceName
	customQuery := "c:" + searchPattern

	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// IteratorByInterfaceImplementation 返回一个接口实现搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - interfaceName: 接口名，如"Listener"或"Handler"
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByInterfaceImplementation(ctx context.Context, interfaceName string) *SearchIterator[*response.Version] {
	searchPattern := "*" + interfaceName
	customQuery := "c:" + searchPattern
	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchByClassSupertype 搜索具有特定父类或接口的类
// 这是SearchClassesWithClassHierarchy和SearchInterfaceImplementations的合并接口
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - supertypeName: 父类型名称（可以是类或接口）
//   - isInterface: 如果为true，表示搜索接口实现；否则搜索类继承
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 版本列表: 包含所有匹配的制品版本信息
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchByClassSupertype(ctx context.Context, supertypeName string, isInterface bool, limit int) ([]*response.Version, error) {
	if isInterface {
		return c.SearchInterfaceImplementations(ctx, supertypeName, limit)
	} else {
		return c.SearchClassesWithClassHierarchy(ctx, supertypeName, limit)
	}
}

// IteratorByClassSupertype 返回一个父类型搜索的迭代器
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - supertypeName: 父类型名称（可以是类或接口）
//   - isInterface: 如果为true，表示搜索接口实现；否则搜索类继承
//
// 返回:
//   - 搜索迭代器，用于逐个处理搜索结果
func (c *Client) IteratorByClassSupertype(ctx context.Context, supertypeName string, isInterface bool) *SearchIterator[*response.Version] {
	if isInterface {
		return c.IteratorByInterfaceImplementation(ctx, supertypeName)
	} else {
		return c.IteratorByClassHierarchy(ctx, supertypeName)
	}
}
