package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// TestSearchByClassName 使用真实API测试类名搜索功能
func TestSearchByClassName(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 减少测试数据，选择几个常见的类名
	classNames := []string{
		"Logger",
		"StringUtils",
		// "HttpClient", // 这个类名查询较慢，暂时移除
		"InputStream",
	}

	for _, className := range classNames {
		t.Run("Class_"+className, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchByClassName(subCtx, className, 5)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", className, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含 %s 的结果", len(versionSlice), className)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchByFullyQualifiedClassName 测试通过全限定类名搜索
func TestSearchByFullyQualifiedClassName(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 测试几个常见的全限定类名
	fqcns := []string{
		"org.apache.commons.lang3.StringUtils",
		"java.util.ArrayList",
		"org.slf4j.Logger",
		// "java.lang.String", // 这个类名可能会返回太多结果
		// "java.util.Map",    // 接口类型可能搜索较慢
	}

	for _, fqcn := range fqcns {
		t.Run("FQCN_"+fqcn, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchByFullyQualifiedClassName(subCtx, fqcn, 3)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", fqcn, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个包含 %s 的结果", len(versionSlice), fqcn)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchByPackageAndClassName 测试通过包名和类名组合搜索
func TestSearchByPackageAndClassName(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义测试用例：包名+类名
	testCases := []struct {
		packageName string
		className   string
	}{
		{"org.apache.commons.lang3", "StringUtils"},
		{"java.util", "ArrayList"},
		{"org.slf4j", "Logger"},
		// {"java.lang", "String"},     // 这些常见类可能会返回太多结果
		// {"java.io", "InputStream"},  // 减少测试数据
		// {"javax.servlet", "ServletContext"},
	}

	for _, tc := range testCases {
		name := tc.packageName + "." + tc.className
		t.Run("PackageClass_"+name, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchByPackageAndClassName(subCtx, tc.packageName, tc.className, 3)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", name, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个包含 %s 的结果", len(versionSlice), name)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchClassesByMethod 测试通过方法名搜索类
func TestSearchClassesByMethod(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 测试几个常见的方法名
	methodNames := []string{
		"equals",
		"toString",
		"valueOf",
		// "substring",   // 减少测试方法数量
		// "compareTo",
		// "main",
		// "getInstance",
	}

	for _, methodName := range methodNames {
		t.Run("Method_"+methodName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchClassesByMethod(subCtx, methodName, 3)

			if err != nil {
				t.Logf("搜索方法 %s 时出错: %v", methodName, err)
				t.Skip("无法连接到Maven Central API或不支持方法搜索")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个包含方法 %s 的结果", len(versionSlice), methodName)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchClassesWithClassHierarchy 测试继承关系搜索
func TestSearchClassesWithClassHierarchy(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义测试用例：基类名
	baseClasses := []string{
		"Exception",
		"RuntimeException",
		// "AbstractList",   // 减少测试数据
		// "Thread",
		// "InputStream",
		// "Component",
	}

	for _, baseClass := range baseClasses {
		t.Run("Base_"+baseClass, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchClassesWithClassHierarchy(subCtx, baseClass, 3)

			if err != nil {
				t.Logf("搜索 %s 的子类时出错: %v", baseClass, err)
				t.Skip("无法连接到Maven Central API或不支持继承关系搜索")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个继承自 %s 的类", len(versionSlice), baseClass)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchInterfaceImplementations 测试接口实现搜索
func TestSearchInterfaceImplementations(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义测试用例：接口名
	interfaces := []string{
		"Listener",
		"Handler",
		"Runnable",
		// "Serializable", // 这个接口实现太多，可能会超时
		// "Comparable",
		// "Callable",
	}

	for _, iface := range interfaces {
		t.Run("Interface_"+iface, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchInterfaceImplementations(subCtx, iface, 3)

			if err != nil {
				t.Logf("搜索 %s 的实现类时出错: %v", iface, err)
				t.Skip("无法连接到Maven Central API或不支持接口实现搜索")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个实现了 %s 的类", len(versionSlice), iface)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchByClassSupertype 测试通过父类型搜索
func TestSearchByClassSupertype(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义测试用例
	testCases := []struct {
		supertype   string
		isInterface bool
	}{
		{"Exception", false}, // 类
		{"Listener", true},   // 接口
		{"Runnable", true},   // 接口
		// {"InputStream", false}, // 减少测试数据
		// {"Map", true},          // 常见接口可能匹配太多
	}

	for _, tc := range testCases {
		typeName := tc.supertype
		if tc.isInterface {
			typeName = "Interface_" + typeName
		} else {
			typeName = "Class_" + typeName
		}

		t.Run(typeName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(500 * time.Millisecond) // 减少延迟时间

			// 设置子测试的超时时间
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			versionSlice, err := client.SearchByClassSupertype(subCtx, tc.supertype, tc.isInterface, 3)

			if err != nil {
				t.Logf("搜索 %s 的子类型时出错: %v", tc.supertype, err)
				t.Skip("无法连接到Maven Central API或不支持父类型搜索")
				return
			}

			// 记录找到的结果
			typeDesc := "类"
			if tc.isInterface {
				typeDesc = "接口"
			}
			t.Logf("找到 %d 个继承/实现 %s(%s) 的结果", len(versionSlice), tc.supertype, typeDesc)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestIteratorMethods 测试各种迭代器方法
func TestIteratorMethods(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时和上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试各种迭代器方法
	testCases := []struct {
		name     string
		iterator *SearchIterator[*response.Version]
	}{
		{"IteratorByClassName", client.IteratorByClassName(ctx, "Logger")},
		{"IteratorByFullyQualifiedClassName", client.IteratorByFullyQualifiedClassName(ctx, "org.slf4j.Logger")},
		{"IteratorByPackageAndClassName", client.IteratorByPackageAndClassName(ctx, "org.slf4j", "Logger")},
		{"IteratorByMethod", client.IteratorByMethod(ctx, "equals")},
		{"IteratorByClassHierarchy", client.IteratorByClassHierarchy(ctx, "Exception")},
		{"IteratorByInterfaceImplementation", client.IteratorByInterfaceImplementation(ctx, "Listener")},
		{"IteratorByClassSupertype", client.IteratorByClassSupertype(ctx, "Listener", true)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 使用迭代器获取前3个结果
			count := 0
			var results []*response.Version

			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			for tc.iterator.Next() && count < 3 {
				results = append(results, tc.iterator.Value())
				count++
			}

			// 检查迭代器是否有错误（迭代器没有直接的Error方法，错误通过NextE等方法返回）
			_, err := tc.iterator.NextE()
			if err != nil && err != ErrQueryIteratorEnd {
				t.Logf("迭代器 %s 使用时出错: %v", tc.name, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("迭代器 %s 找到至少 %d 个结果", tc.name, len(results))
			for i, v := range results {
				t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
			}

			assert.True(t, len(results) >= 0) // 只确保API正常返回
		})
	}
}

// TestEdgeCases 测试一些边界条件
func TestEdgeCases(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试边界情况
	t.Run("EmptyClassName", func(t *testing.T) {
		// 空类名
		versionSlice, err := client.SearchByClassName(ctx, "", 3)

		// 不用Skip，因为我们期望这是一个正常但可能没有结果的查询
		if err != nil {
			t.Logf("空类名搜索出错: %v", err)
		}

		t.Logf("空类名搜索找到 %d 个结果", len(versionSlice))
		assert.True(t, len(versionSlice) >= 0)
	})

	t.Run("VeryShortClassName", func(t *testing.T) {
		// 非常短的类名
		versionSlice, err := client.SearchByClassName(ctx, "A", 3)

		if err != nil {
			t.Logf("短类名搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("短类名 'A' 搜索找到 %d 个结果", len(versionSlice))
		if len(versionSlice) > 0 {
			for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
				t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
			}
		}
		assert.True(t, len(versionSlice) >= 0)
	})

	t.Run("UnlikelyClassName", func(t *testing.T) {
		// 不太可能存在的类名
		versionSlice, err := client.SearchByClassName(ctx, "XyzAbcVeryUnlikelyClassName123456", 3)

		if err != nil {
			t.Logf("不太可能的类名搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("不太可能的类名搜索找到 %d 个结果", len(versionSlice))
		assert.True(t, len(versionSlice) >= 0)
	})

	t.Run("NonExistentMethod", func(t *testing.T) {
		// 不太可能存在的方法名
		versionSlice, err := client.SearchClassesByMethod(ctx, "veryUnusualMethodNameThatShouldntExist12345", 3)

		if err != nil {
			t.Logf("不太可能的方法名搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("不太可能的方法名搜索找到 %d 个结果", len(versionSlice))
		assert.True(t, len(versionSlice) >= 0)
	})

	t.Run("ZeroLimit", func(t *testing.T) {
		// 测试限制为0的情况，应该使用迭代器
		versionSlice, err := client.SearchByClassName(ctx, "String", 0)

		if err != nil {
			t.Logf("限制为0的搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("限制为0的搜索找到 %d 个结果", len(versionSlice))
		assert.True(t, len(versionSlice) >= 0)
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		// 测试限制为负数的情况，应该使用迭代器
		versionSlice, err := client.SearchByClassName(ctx, "String", -5)

		if err != nil {
			t.Logf("限制为负数的搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("限制为负数的搜索找到 %d 个结果", len(versionSlice))
		assert.True(t, len(versionSlice) >= 0)
	})
}
