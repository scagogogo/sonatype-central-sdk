package examples

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

// HighlightSearchExample 展示如何使用高亮搜索功能
func HighlightSearchExample() {
	// 初始化客户端
	client := api.NewClient()

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行基本高亮搜索示例
	basicHighlightSearch(ctx, client)

	// 执行通配符搜索示例
	wildcardSearch(ctx, client)

	// 提取高亮信息示例
	extractHighlights(ctx, client)
}

// basicHighlightSearch 展示基本的高亮搜索功能
func basicHighlightSearch(ctx context.Context, client *api.Client) {
	fmt.Println("\n1. 基本高亮搜索示例")
	fmt.Println("---------------------------")

	// 构建搜索请求
	className := "org.apache.commons.io.FileUtils"

	// 执行搜索
	resp, err := client.SearchClassesWithHighlighting(ctx, className, 10)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	// 输出结果
	fmt.Printf("找到 %d 个结果\n", resp.ResponseBody.NumFound)

	// 显示高亮信息
	if len(resp.ResponseBody.Docs) > 0 {
		fmt.Println("\n高亮结果:")
		for i, doc := range resp.ResponseBody.Docs {
			if i >= 3 {
				fmt.Printf("... 还有 %d 个结果\n", len(resp.ResponseBody.Docs)-i)
				break
			}

			fmt.Printf("\n文档 #%d:\n", i+1)
			fmt.Printf("  组ID: %s\n", doc.GroupId)
			fmt.Printf("  构件ID: %s\n", doc.ArtifactId)
			fmt.Printf("  版本: %s\n", doc.Version)

			// 获取该文档的高亮信息
			if resp.Highlighting != nil {
				if hl, exists := resp.Highlighting[doc.ID]; exists && len(hl["fch"]) > 0 {
					fmt.Printf("  类名高亮: %s\n", strings.Join(hl["fch"], ", "))
				}
			}
		}
	}
}

// wildcardSearch 展示使用通配符进行搜索
func wildcardSearch(ctx context.Context, client *api.Client) {
	fmt.Println("\n2. 通配符搜索示例")
	fmt.Println("---------------------------")

	// 使用通配符搜索Spring Controller
	className := "org.springframework.web.*Controller"

	// 执行搜索
	resp, err := client.SearchClassesWithHighlighting(ctx, className, 50)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	// 分析结果
	fmt.Printf("找到 %d 个Spring Controller类\n", resp.ResponseBody.NumFound)

	// 统计不同类型的控制器
	controllerTypes := make(map[string]int)
	for _, doc := range resp.ResponseBody.Docs {
		// 由于Version中没有直接存储类名，我们从高亮信息中提取
		if resp.Highlighting != nil {
			if hl, exists := resp.Highlighting[doc.ID]; exists && len(hl["fch"]) > 0 {
				for _, c := range hl["fch"] {
					// 移除高亮标签以获取纯文本类名
					plainText := strings.Replace(strings.Replace(c, "<em>", "", -1), "</em>", "", -1)
					parts := strings.Split(plainText, ".")
					if len(parts) > 0 {
						className := parts[len(parts)-1]
						if strings.HasSuffix(className, "Controller") {
							controllerTypes[className]++
						}
					}
				}
			}
		}
	}

	// 显示统计结果
	fmt.Println("\n控制器类型统计:")
	for controller, count := range controllerTypes {
		if count > 1 {
			fmt.Printf("  %s: 出现 %d 次\n", controller, count)
		}
	}
}

// extractHighlights 展示如何提取和使用高亮信息
func extractHighlights(ctx context.Context, client *api.Client) {
	fmt.Println("\n3. 提取高亮信息示例")
	fmt.Println("---------------------------")

	// 搜索JUnit Assert类
	className := "org.junit.Assert"

	// 执行搜索
	resp, err := client.SearchClassesWithHighlighting(ctx, className, 10)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	// 显示版本和高亮
	if len(resp.ResponseBody.Docs) > 0 {
		fmt.Println("JUnit Assert类的版本和使用示例:")

		// 提取版本
		versions := make(map[string]bool)
		for _, doc := range resp.ResponseBody.Docs {
			versions[doc.Version] = true
		}

		fmt.Printf("\n可用版本: %s\n", strings.Join(getMapKeys(versions), ", "))

		// 提取高亮信息
		if resp.Highlighting != nil {
			fmt.Println("\n代码使用示例:")

			exampleCount := 0
			for id, hl := range resp.Highlighting {
				if exampleCount >= 3 {
					break
				}

				if len(hl["fch"]) > 0 {
					for _, doc := range resp.ResponseBody.Docs {
						if doc.ID == id {
							fmt.Printf("\n示例 #%d (来自 %s:%s):\n", exampleCount+1, doc.GroupId, doc.ArtifactId)
							fmt.Printf("  %s\n", strings.Join(hl["fch"], "\n  "))
							exampleCount++
							break
						}
					}
				}
			}
		}

		// 提供使用建议
		fmt.Println("\n推荐用法:")
		fmt.Println(`
  // 在Maven项目中添加依赖
  <dependency>
      <groupId>junit</groupId>
      <artifactId>junit</artifactId>
      <version>4.13.2</version>
      <scope>test</scope>
  </dependency>
  
  // 示例代码
  import org.junit.Assert;
  import org.junit.Test;
  
  public class MyTest {
      @Test
      public void testSomething() {
          // 使用Assert中的各种断言方法
          Assert.assertEquals("期望值与实际值应相等", expected, actual);
          Assert.assertTrue("条件应为true", condition);
          Assert.assertNotNull("对象不应为null", object);
      }
  }
  `)
	}
}

// getMapKeys 辅助函数，获取map的所有键作为字符串切片
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
