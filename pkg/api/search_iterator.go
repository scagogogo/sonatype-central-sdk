package api

import (
	"context"
	"errors"

	"github.com/golang-infrastructure/go-iterator"
	"github.com/golang-infrastructure/go-queue"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

var (
	// ErrQueryIteratorEnd 迭代器已经遍历完毕
	ErrQueryIteratorEnd = errors.New("query iterator ended")
)

// SearchIterator 把搜索结果以一个迭代器的形式返回
type SearchIterator[Doc any] struct {

	// 搜索参数
	search *request.SearchRequest

	// 缓存大小
	buff *queue.LinkedQueue[Doc]

	// 记录遍历位置
	total     int
	current   int
	nextStart int

	// 迭代器是否发生错误
	err error

	// 客户端引用
	client *Client
}

var _ iterator.ErrorableIterator[any] = &SearchIterator[any]{}

// NewSearchIterator 创建一个新的搜索迭代器实例
//
// 该方法用于创建一个SearchIterator实例，该迭代器能够高效地处理Maven Central搜索结果。
// 通过泛型参数Doc指定搜索结果文档的类型，如*response.Artifact、*response.Version等。
// 迭代器采用惰性加载模式，只有在需要数据时才会发起网络请求，并自动处理分页逻辑，
// 使开发者能够像处理内存中的集合一样处理大量的搜索结果。
//
// 参数:
//   - search: 搜索请求对象，包含查询条件、排序规则等参数
//
// 返回:
//   - *SearchIterator[Doc]: 一个新的搜索结果迭代器实例
//
// 使用示例:
//
//	// 创建一个查询条件
//	query := request.NewQuery().
//	    SetGroupId("org.apache.logging.log4j").
//	    SetArtifactId("log4j-core")
//
//	// 创建搜索请求，设置每页返回20条结果
//	searchReq := request.NewSearchRequest().
//	    SetQuery(query).
//	    SetLimit(20).
//	    SetSort("timestamp", false)  // 按时间降序排列
//
//	// 创建迭代器，处理Artifact类型的结果
//	iterator := api.NewSearchIterator[*response.Artifact](searchReq)
//
//	// 设置客户端并处理结果（推荐方式）
//	client := api.NewClient()
//	iterator.WithClient(client)
//
//	// 使用迭代器API处理结果
//	for iterator.Next() {
//	    artifact := iterator.Value()
//	    fmt.Printf("制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func NewSearchIterator[Doc any](search *request.SearchRequest) *SearchIterator[Doc] {
	return &SearchIterator[Doc]{
		search:    search,
		buff:      queue.NewLinkedQueue[Doc](),
		total:     -1,
		current:   0,
		nextStart: 0,
		err:       nil,
	}
}

// WithClient 设置迭代器使用的API客户端
//
// 该方法用于指定迭代器执行搜索请求时应使用的Client实例。如果不调用此方法，
// 迭代器将使用默认客户端配置执行请求。通过设置特定的客户端，可以控制重试策略、
// 超时设置、基础URL等参数，使迭代器适应不同的使用场景和环境配置。
//
// 参数:
//   - client: 要使用的API客户端实例，包含了连接参数、基础URL等配置
//
// 返回:
//   - *SearchIterator[Doc]: 返回迭代器本身，支持链式调用
//
// 使用示例:
//
//	// 创建一个自定义配置的客户端
//	client := api.NewClient().
//	    WithBaseURL("https://custom-maven-mirror.example.com/solrsearch").
//	    WithTimeout(60*time.Second)
//
//	// 创建搜索请求
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query)
//
//	// 创建迭代器并设置使用自定义客户端
//	iterator := NewSearchIterator[*response.Artifact](searchReq).
//	    WithClient(client)
//
//	// 使用迭代器处理结果
//	for iterator.Next() {
//	    artifact := iterator.Value()
//	    // 处理每个搜索结果...
//	}
func (x *SearchIterator[Doc]) WithClient(client *Client) *SearchIterator[Doc] {
	x.client = client
	return x
}

// ToSlice 将迭代器中的所有元素收集到一个切片中
//
// 该方法会遍历整个迭代器，将所有元素收集到一个切片中返回。这种方法适合处理数量可控的搜索结果，
// 当结果数量非常大时应谨慎使用，因为它会一次性将所有结果加载到内存中，可能导致内存占用过高。
// 方法会自动处理迭代和分页，直到获取所有结果或遇到错误为止。
//
// 参数: 无
//
// 返回:
//   - []Doc: 包含迭代器中所有元素的切片
//   - error: 如果在迭代过程中发生任何错误，将返回该错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个搜索对象，限制结果数量以避免内存问题
//	query := request.NewQuery().SetGroupId("org.apache.commons").SetArtifactId("commons-lang3")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(50)
//
//	// 创建迭代器
//	iterator := NewSearchIterator[*response.Artifact](searchReq).WithClient(client)
//
//	// 将所有结果收集到切片中
//	artifacts, err := iterator.ToSlice()
//	if err != nil {
//	    log.Fatalf("获取所有结果失败: %v", err)
//	}
//
//	// 使用收集的结果
//	fmt.Printf("找到 %d 个匹配的制品\n", len(artifacts))
//	for i, artifact := range artifacts {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
//
//	// 针对可能返回大量结果的查询，应当设置合理的限制或使用迭代器模式处理
func (x *SearchIterator[Doc]) ToSlice() ([]Doc, error) {
	slice := make([]Doc, 0)
	for {
		hasNext, err := x.NextE()
		if err != nil {
			return slice, err
		}
		if !hasNext {
			return slice, nil
		}
		artifact, err := x.ValueE()
		if err != nil {
			return slice, err
		}
		slice = append(slice, artifact)
	}
}

// Next 检查迭代器是否还有下一个元素
//
// 该方法用于判断迭代器中是否有更多元素可以获取。它是迭代器模式的核心方法之一，
// 通常在使用for循环处理迭代器时作为条件使用。方法内部调用NextE，但会忽略可能
// 的错误信息。如果需要错误处理，应使用NextE方法代替。
//
// 参数: 无
//
// 返回:
//   - bool: 如果迭代器中还有更多元素可以获取，返回true；否则返回false
//
// 使用示例:
//
//	// 创建一个查询
//	query := request.NewQuery().SetGroupId("com.google.guava")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(10)
//
//	// 创建并配置迭代器
//	iterator := NewSearchIterator[*response.Artifact](searchReq).WithClient(client)
//
//	// 使用Next()方法迭代处理结果
//	for iterator.Next() {
//	    // 获取当前元素并处理
//	    artifact := iterator.Value()
//	    fmt.Printf("制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
//
//	// 注意: 这种方式会忽略错误，如果需要错误处理，应使用NextE和ValueE方法
func (x *SearchIterator[Doc]) Next() bool {
	hasNext, _ := x.NextE()
	return hasNext
}

// Value 获取迭代器当前位置的元素值
//
// 该方法返回迭代器当前指向的元素。它应当在调用Next()方法确认迭代器有下一个元素后使用。
// 方法内部调用ValueE获取元素，但会忽略可能发生的错误。如果迭代器已经遍历完毕或发生错误，
// 将返回该类型的零值。为了获得更好的错误处理，建议使用ValueE方法。
//
// 参数: 无
//
// 返回:
//   - Doc: 当前位置的文档对象；如果迭代器已结束或发生错误，返回类型的零值
//
// 使用示例:
//
//	// 创建查询并配置迭代器
//	query := request.NewQuery().SetGroupId("org.springframework")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(10)
//	iterator := NewSearchIterator[*response.Artifact](searchReq).WithClient(client)
//
//	// 使用标准迭代器模式处理结果
//	for iterator.Next() {
//	    // 获取当前元素
//	    artifact := iterator.Value()
//
//	    // 处理当前元素
//	    fmt.Printf("Spring制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
//
//	// 注意：处理大量数据时，应考虑设置适当的限制，避免处理过多结果
func (x *SearchIterator[Doc]) Value() Doc {
	value, _ := x.ValueE()
	return value
}

// NextE 检查迭代器是否有下一个元素，并返回可能的错误
//
// 该方法是迭代器的核心功能之一，负责检查是否有更多元素可供处理，同时提供完整的错误处理。
// 方法内部管理迭代器状态、缓冲区和分页逻辑，使客户端代码无需关心底层的搜索实现细节。
// 首次调用时会初始化迭代器并发送第一次网络请求；后续调用会管理缓冲区和分页，
// 必要时自动获取下一页数据。
//
// 参数: 无
//
// 返回:
//   - bool: 如果迭代器中还有更多元素可以获取，返回true；否则返回false
//   - error: 如果在确定是否有下一元素的过程中发生错误，返回相应错误；成功时返回nil
//
// 使用示例:
//
//	// 创建查询并配置迭代器
//	query := request.NewQuery().SetGroupId("org.apache.logging.log4j")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(20)
//	iterator := NewSearchIterator[*response.Artifact](searchReq).WithClient(client)
//
//	// 使用带错误处理的迭代方式
//	for {
//	    // 检查是否有下一个元素，同时处理可能的错误
//	    hasNext, err := iterator.NextE()
//	    if err != nil {
//	        log.Fatalf("迭代过程中发生错误: %v", err)
//	    }
//
//	    // 如果没有更多元素，结束迭代
//	    if !hasNext {
//	        break
//	    }
//
//	    // 获取当前元素并处理可能的错误
//	    artifact, err := iterator.ValueE()
//	    if err != nil {
//	        log.Fatalf("获取元素时发生错误: %v", err)
//	    }
//
//	    // 处理获取到的元素
//	    fmt.Printf("日志组件: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (x *SearchIterator[Doc]) NextE() (bool, error) {
	// 如果迭代器之前遇到过错误，避免继续执行，直接返回该错误
	if x.err != nil {
		return false, x.err
	}

	if x.total < 0 {
		// total为负值表示这是首次调用，需要初始化迭代器
		var r *response.Response[Doc]
		var err error

		// 根据是否设置了客户端决定如何执行请求
		if x.client != nil {
			// 使用指定的客户端发送请求
			r, err = SearchRequestJsonDoc[Doc](x.client, context.Background(), x.search)
		} else {
			// 使用默认客户端发送请求
			r, err = SearchRequestJsonDoc[Doc](nil, context.Background(), x.search)
		}

		// 处理网络请求可能发生的错误
		if err != nil {
			// 保存错误状态，避免后续重复请求
			x.err = err
			return false, err
		}

		// 验证响应结果是否合法
		if r == nil || r.ResponseBody == nil {
			x.err = errors.New("empty response body")
			return false, x.err
		}

		// 获取查询结果的总数量，用于确定何时结束迭代
		x.total = r.ResponseBody.NumFound

		// 计算下一次查询的起始位置
		x.nextStart = x.nextStart + len(r.ResponseBody.Docs)

		// 将结果文档批量添加到内部缓冲区
		err = x.buff.Put(r.ResponseBody.Docs...)
		if err != nil {
			x.err = err
			return false, err
		}

		// 判断是否还有更多结果（当前已处理数量与总数比较）
		return x.current < x.total, nil
	} else {
		// 已经初始化过，需要判断是否继续获取下一批数据

		// 如果缓冲区中还有未处理的文档，直接返回true
		if x.buff.IsNotEmpty() {
			return true, nil
		}

		// 如果当前已处理的文档数量达到或超过总数，表示迭代结束
		if x.current >= x.total {
			return false, nil
		}

		// 更新搜索请求的起始位置，准备获取下一页数据
		x.search.Start = x.nextStart

		// 执行获取下一页数据的请求
		var r *response.Response[Doc]
		var err error
		if x.client != nil {
			r, err = SearchRequestJsonDoc[Doc](x.client, context.Background(), x.search)
		} else {
			r, err = SearchRequestJsonDoc[Doc](nil, context.Background(), x.search)
		}

		// 处理请求过程中可能发生的错误
		if err != nil {
			x.err = err
			return false, err
		}

		// 验证响应结果是否有效
		if r == nil || r.ResponseBody == nil {
			x.err = errors.New("empty response body")
			return false, x.err
		}

		// 更新下一次请求的起始位置
		x.nextStart = x.nextStart + len(r.ResponseBody.Docs)

		// 将新获取的文档添加到缓冲区
		err = x.buff.Put(r.ResponseBody.Docs...)
		if err != nil {
			x.err = err
			return false, err
		}

		// 返回是否还有更多结果需要处理
		return x.current < x.total, nil
	}
}

// ValueE 获取迭代器当前位置的元素值，并返回可能发生的错误
//
// 该方法是迭代器的核心实现之一，负责获取迭代器当前指向的元素，并将迭代器位置前进一步。
// 方法会从内部缓冲区中取出并移除一个元素，确保每次调用都能获取不同的元素。在调用此方法前，
// 应当先调用NextE方法确认迭代器是否还有下一个元素，否则可能会收到ErrQueryIteratorEnd错误。
//
// 参数: 无
//
// 返回:
//   - Doc: 当前位置的文档对象，如果缓冲区为空则返回类型的零值
//   - error: 如果缓冲区为空返回ErrQueryIteratorEnd，如果获取元素过程中发生其他错误则返回相应错误
//
// 使用示例:
//
//	// 创建并初始化迭代器
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query)
//	iterator := NewSearchIterator[*response.Artifact](searchReq).WithClient(client)
//
//	// 使用迭代器模式处理结果
//	for iterator.Next() {
//	    artifact := iterator.Value()
//	    fmt.Printf("找到制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
//
//	// 或者使用带错误处理的方式
//	for {
//	    hasNext, err := iterator.NextE()
//	    if err != nil {
//	        log.Fatalf("迭代出错: %v", err)
//	    }
//	    if !hasNext {
//	        break
//	    }
//
//	    artifact, err := iterator.ValueE()
//	    if err != nil {
//	        log.Fatalf("获取元素出错: %v", err)
//	    }
//	    // 处理artifact...
//	}
func (x *SearchIterator[Doc]) ValueE() (Doc, error) {
	// 检查缓冲区是否为空，如果为空则无法获取元素
	if x.buff.IsEmpty() {
		// 返回类型的零值和迭代结束错误
		var zero Doc
		return zero, ErrQueryIteratorEnd
	}

	// 递增已处理元素计数
	x.current++

	// 从缓冲区中取出并移除一个元素
	return x.buff.TakeE()
}
