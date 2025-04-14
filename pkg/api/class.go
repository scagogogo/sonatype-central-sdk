package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByClassName 根据类名搜索相关制品
//
// 该方法在Maven Central仓库中搜索包含指定类名的所有制品并返回详细信息。
// 搜索只根据类名进行，不考虑包名，因此同名但不同包的类都会被找到。
// 如果需要搜索特定包中的类，应使用完全限定类名搜索功能。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时时间
//   - class: 要搜索的类名（不含包名），如"ArrayList"或"HttpClient"
//   - limit: 最大返回结果数量，建议值为10-50；如果小于等于0则返回所有结果(使用迭代器内部实现)
//
// 返回:
//   - []*response.Version: 包含所有匹配的制品版本信息的数组
//   - error: 如果搜索过程中发生错误，如网络错误、参数错误等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// 搜索ArrayList类
//	versions, err := client.SearchByClassName(ctx, "ArrayList", 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	fmt.Printf("找到 %d 个包含ArrayList类的制品:\n", len(versions))
//	for i, version := range versions {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, version.GroupId, version.ArtifactId, version.Version)
//	}
//
//	// 获取所有结果(不限制数量)
//	allVersions, err := client.SearchByClassName(ctx, "Logger", 0)
//	if err == nil {
//	    fmt.Printf("共找到 %d 个包含Logger类的制品\n", len(allVersions))
//	}
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
//
// 该方法创建一个迭代器，用于高效处理大规模类名搜索结果。
// 迭代器模式适合处理可能返回大量数据的搜索，通过分批获取数据
// 减少内存占用，特别适合需要处理全部结果但又不希望一次性加载所有数据的场景。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置较长的超时时间
//   - class: 要搜索的类名（不含包名），如"HashMap"或"Executor"
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器，提供逐批获取结果的能力
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器搜索所有包含Connection类的制品
//	iterator := client.IteratorByClassName(ctx, "Connection")
//
//	// 使用迭代器处理所有结果
//	count := 0
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    // 处理当前批次的结果
//	    for _, version := range versions {
//	        count++
//	        fmt.Printf("找到第 %d 个结果: %s:%s:%s\n",
//	            count, version.GroupId, version.ArtifactId, version.Version)
//	    }
//
//	    // 可以设置处理数量限制
//	    if count >= 100 {
//	        fmt.Println("已处理100个结果，停止迭代")
//	        break
//	    }
//	}
//
//	// 或者直接转换为切片(适用于确认结果数量不会太大的情况)
//	allVersions, err := iterator.ToSlice()
//	if err == nil {
//	    fmt.Printf("共找到 %d 个包含Connection类的制品\n", len(allVersions))
//	}
func (c *Client) IteratorByClassName(ctx context.Context, class string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchClassesByMethod 搜索包含特定方法的类
//
// 该方法在Maven Central仓库中搜索包含指定方法名的Java类，返回相关制品信息。
// 这是一个高级搜索功能，可以帮助开发者查找特定功能的实现，尤其适合寻找API和工具类。
// 注意此功能依赖于Maven索引中的方法信息，在某些情况下可能不够完整或准确。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时时间
//   - methodName: 要搜索的方法名，如"parse"、"encode"或"transform"
//   - limit: 最大返回结果数量，建议值为10-20；如果小于等于0则返回所有结果(使用迭代器实现)
//
// 返回:
//   - []*response.Version: 包含所有匹配的制品版本信息的数组
//   - error: 如果搜索过程中发生错误，如网络错误、参数错误等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// 搜索实现了parse方法的类
//	versions, err := client.SearchClassesByMethod(ctx, "parse", 15)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	fmt.Printf("找到 %d 个包含parse方法的制品:\n", len(versions))
//	for i, version := range versions {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, version.GroupId, version.ArtifactId, version.Version)
//	    // 可能需要进一步检查类文件确认方法的确切签名和功能
//	}
//
//	// 查找所有包含serialize方法的类(不限制数量)
//	allSerializers, err := client.SearchClassesByMethod(ctx, "serialize", 0)
//	if err == nil {
//	    fmt.Printf("共找到 %d 个包含serialize方法的制品\n", len(allSerializers))
//	}
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
//
// 该方法创建一个迭代器对象，用于高效地处理包含特定方法名的类搜索结果。
// 迭代器模式特别适合处理可能返回大量数据的方法搜索，通过分批获取结果数据
// 减少内存占用，避免一次性加载大量数据导致的性能问题。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置较长的超时时间
//   - methodName: 要搜索的方法名，如"execute"、"connect"或"parse"
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器，提供Next()、HasNext()等方法遍历结果
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器搜索所有包含convert方法的类
//	iterator := client.IteratorByMethod(ctx, "convert")
//
//	// 逐批处理搜索结果
//	processedCount := 0
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    // 处理当前批次的结果
//	    for _, version := range versions {
//	        processedCount++
//	        fmt.Printf("找到包含convert方法的制品 #%d: %s:%s:%s\n",
//	            processedCount, version.GroupId, version.ArtifactId, version.Version)
//
//	        // 可以对感兴趣的制品进行进一步处理
//	        if strings.Contains(version.GroupId, "apache") {
//	            fmt.Println("  - 这是Apache开源项目的一部分")
//	        }
//	    }
//
//	    // 可以设置一个处理上限，避免处理过多结果
//	    if processedCount >= 200 {
//	        fmt.Println("已处理200个结果，停止后续处理")
//	        break
//	    }
//	}
func (c *Client) IteratorByMethod(ctx context.Context, methodName string) *SearchIterator[*response.Version] {
	customQuery := "m:" + methodName
	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchClassesWithClassHierarchy 搜索继承自特定基类的类
//
// 该方法在Maven Central仓库中搜索可能继承自指定基类的Java类。
// 这是一个高级搜索功能，可以帮助开发者查找特定基类的子类实现，有助于研究类继承关系。
// 由于Maven索引中可能不直接包含完整的继承关系信息，因此该方法基于类名相似性进行初步过滤，
// 可能需要进一步下载和分析类文件以确认真实的继承关系。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时时间
//   - baseClassName: 基类名称，如"AbstractList"、"Exception"或"Thread"
//   - limit: 最大返回结果数量，建议值为20-50；如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Version: 包含所有可能继承自指定基类的制品版本信息
//   - error: 如果搜索过程中发生错误，如网络错误、参数错误等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// 搜索可能继承自Exception类的所有类
//	versions, err := client.SearchClassesWithClassHierarchy(ctx, "Exception", 25)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	fmt.Printf("找到 %d 个可能继承自Exception的类:\n", len(versions))
//	for i, version := range versions {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, version.GroupId, version.ArtifactId, version.Version)
//	}
//
//	// 注意：要确认真实的继承关系，可能需要额外步骤
//	fmt.Println("\n注意: 这些结果是基于类名相似性的初步筛选，")
//	fmt.Println("要确认真实的继承关系，需要下载并分析.class文件")
//
//	// 进一步验证特定制品是否真正继承自基类的伪代码:
//	// 1. 下载JAR文件: client.Download(ctx, jarPath)
//	// 2. 解压JAR文件并查找特定类的.class文件
//	// 3. 使用Java反射或字节码分析工具分析类的继承结构
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
//
// 该方法创建一个迭代器对象，用于高效地处理可能继承自指定基类的Java类搜索结果。
// 与SearchClassesWithClassHierarchy方法类似，该方法也基于类名相似性进行初步筛选，
// 但采用迭代器模式，适合处理大量的搜索结果，降低内存占用并提供更灵活的处理方式。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置较长的超时时间
//   - baseClassName: 基类名称，如"Exception"、"AbstractCollection"或"Component"
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器，提供对搜索结果的流式访问
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器搜索所有可能继承自Thread类的实现
//	iterator := client.IteratorByClassHierarchy(ctx, "Thread")
//
//	// 使用迭代器逐批处理搜索结果
//	count := 0
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    // 处理当前批次的结果
//	    for _, version := range versions {
//	        count++
//	        // 打印前50个结果
//	        if count <= 50 {
//	            fmt.Printf("%d. 可能的Thread子类: %s:%s:%s\n",
//	                count, version.GroupId, version.ArtifactId, version.Version)
//	        }
//	    }
//
//	    // 可以设置处理上限
//	    if count >= 200 {
//	        fmt.Println("已处理200个结果，停止处理")
//	        break
//	    }
//	}
//
//	fmt.Printf("共找到 %d 个可能继承自Thread的类\n", count)
//
//	// 注意: 要确认真实的继承关系，需要额外的验证步骤
func (c *Client) IteratorByClassHierarchy(ctx context.Context, baseClassName string) *SearchIterator[*response.Version] {
	// 使用简单的类名搜索来模拟继承关系搜索
	return c.IteratorByClassName(ctx, baseClassName)
}

// SearchInterfaceImplementations 尝试搜索指定接口的实现类
//
// 该方法在Maven Central仓库中搜索可能实现了指定接口的Java类，使用类名模式匹配策略。
// 由于Maven索引可能不直接包含接口实现关系的信息，该方法基于一个启发式策略：
// 搜索类名中包含或以接口名结尾的类，如搜索"Listener"接口的实现类时，会查找诸如
// "EventListener"、"MouseListener"等类名。这种方法虽然不够精确，但可以提供一个
// 合理的初步结果集，帮助开发者发现潜在的接口实现。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时时间
//   - interfaceName: 接口名称，如"Listener"、"Handler"或"Factory"
//   - limit: 最大返回结果数量，建议值为20-50；如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Version: 包含所有可能实现了指定接口的制品版本信息
//   - error: 如果搜索过程中发生错误，如网络错误、参数错误等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// 搜索Serializable接口的可能实现类
//	versions, err := client.SearchInterfaceImplementations(ctx, "Serializable", 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	fmt.Printf("找到 %d 个可能实现了Serializable接口的类:\n", len(versions))
//	for i, version := range versions {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, version.GroupId, version.ArtifactId, version.Version)
//	}
//
//	// 注意：此结果需要进一步验证
//	fmt.Println("\n注意: 搜索结果是基于命名规则的初步推断，")
//	fmt.Println("要确认接口实现关系，需要下载制品并分析字节码")
//
//	// 如需进一步验证，可以:
//	// 1. 下载相关JAR文件
//	// 2. 使用反射或者字节码分析工具检查类是否真正实现了接口
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
//
// 该方法创建一个迭代器对象，用于高效地处理可能实现了指定接口的Java类搜索结果。
// 与SearchInterfaceImplementations方法类似，该方法也使用类名模式匹配策略进行搜索，
// 但采用迭代器模式，适合处理大规模搜索结果，尤其是当搜索常见接口如"Serializable"
// 或"Comparable"等可能有成千上万个实现类的情况。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置较长的超时时间
//   - interfaceName: 接口名称，如"Closeable"、"Runnable"或"Comparable"
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器，提供对搜索结果的分批访问能力
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器搜索所有可能实现了Comparable接口的类
//	iterator := client.IteratorByInterfaceImplementation(ctx, "Comparable")
//
//	// 使用迭代器逐批处理搜索结果
//	processedCount := 0
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    // 处理当前批次的结果
//	    for _, version := range versions {
//	        processedCount++
//	        // 仅处理前100个结果
//	        if processedCount <= 100 {
//	            fmt.Printf("%d. 可能实现了Comparable的类: %s:%s:%s\n",
//	                processedCount, version.GroupId, version.ArtifactId, version.Version)
//	        }
//	    }
//
//	    // 设置一个合理的处理上限，避免处理过多结果
//	    if processedCount >= 500 {
//	        fmt.Println("已处理500个结果，停止后续处理")
//	        break
//	    }
//	}
//
//	// 分析搜索结果的分布情况（示例）
//	fmt.Printf("共找到 %d 个可能实现了Comparable接口的制品\n", processedCount)
func (c *Client) IteratorByInterfaceImplementation(ctx context.Context, interfaceName string) *SearchIterator[*response.Version] {
	searchPattern := "*" + interfaceName
	customQuery := "c:" + searchPattern
	query := request.NewQuery().SetCustomQuery(customQuery)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// SearchByClassSupertype 搜索具有特定父类或接口的类
//
// 该方法根据提供的父类型名称搜索制品，可以搜索实现了特定接口的类或者继承自特定基类的类。
// 通过isInterface参数控制搜索类型：true表示搜索接口实现，false表示搜索类继承关系。
// 这是SearchClassesWithClassHierarchy和SearchInterfaceImplementations的统一接口，
// 便于调用者在不确定是搜索接口还是类的情况下使用。
//
// 参数:
//   - ctx: 上下文，用于控制请求的超时和取消
//   - supertypeName: 父类型名称，可能是接口名如"Serializable"或类名如"Exception"
//   - isInterface: 搜索模式标志，true表示搜索接口实现，false表示搜索类继承
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Version: 包含所有匹配的制品版本信息的列表
//   - error: 如果搜索过程中发生错误，如网络问题、参数无效等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索实现了Serializable接口的类
//	interfaceResults, err := client.SearchByClassSupertype(ctx, "Serializable", true, 10)
//	if err != nil {
//	    log.Fatalf("搜索接口实现失败: %v", err)
//	}
//	fmt.Printf("找到 %d 个可能实现了Serializable接口的制品\n", len(interfaceResults))
//
//	// 搜索继承自Exception类的类
//	classResults, err := client.SearchByClassSupertype(ctx, "Exception", false, 10)
//	if err != nil {
//	    log.Fatalf("搜索类继承失败: %v", err)
//	}
//	fmt.Printf("找到 %d 个可能继承自Exception类的制品\n", len(classResults))
func (c *Client) SearchByClassSupertype(ctx context.Context, supertypeName string, isInterface bool, limit int) ([]*response.Version, error) {
	if isInterface {
		return c.SearchInterfaceImplementations(ctx, supertypeName, limit)
	} else {
		return c.SearchClassesWithClassHierarchy(ctx, supertypeName, limit)
	}
}

// IteratorByClassSupertype 返回一个用于父类型搜索的迭代器
//
// 该方法创建一个迭代器用于按需获取搜索结果，适合处理大量数据的情况。
// 与SearchByClassSupertype类似，它可以根据isInterface参数搜索接口实现或类继承关系。
// 使用迭代器模式可以避免一次性加载所有结果，降低内存使用并提高性能。
//
// 参数:
//   - ctx: 上下文，用于控制请求的超时和取消
//   - supertypeName: 父类型名称，可能是接口名如"Runnable"或类名如"Thread"
//   - isInterface: 搜索模式标志，true表示搜索接口实现，false表示搜索类继承
//
// 返回:
//   - *SearchIterator[*response.Version]: 搜索结果迭代器，提供Next()、HasNext()等方法逐个访问结果
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器搜索实现了Comparable接口的类
//	iterator := client.IteratorByClassSupertype(ctx, "Comparable", true)
//
//	// 逐个处理搜索结果
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    for _, version := range versions {
//	        fmt.Printf("找到可能实现了Comparable的类: %s:%s:%s\n",
//	            version.GroupId, version.ArtifactId, version.Version)
//	    }
//	}
//
//	// 或者直接将所有结果转换为切片
//	allResults, err := iterator.ToSlice()
//	if err != nil {
//	    log.Fatalf("获取所有结果失败: %v", err)
//	}
func (c *Client) IteratorByClassSupertype(ctx context.Context, supertypeName string, isInterface bool) *SearchIterator[*response.Version] {
	if isInterface {
		return c.IteratorByInterfaceImplementation(ctx, supertypeName)
	} else {
		return c.IteratorByClassHierarchy(ctx, supertypeName)
	}
}
