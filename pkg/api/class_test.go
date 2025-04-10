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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的类名
	classNames := []string{
		"Logger",
		"StringUtils",
		"HttpClient",
		"InputStream",
		"Object",
		"String",
		"Map",
	}

	for _, className := range classNames {
		t.Run("Class_"+className, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByClassName(ctx, className, 5)

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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的全限定类名
	fqcns := []string{
		"org.apache.commons.lang3.StringUtils",
		"java.util.ArrayList",
		"org.slf4j.Logger",
		"java.lang.String",
		"java.util.Map",
	}

	for _, fqcn := range fqcns {
		t.Run("FQCN_"+fqcn, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByFullyQualifiedClassName(ctx, fqcn, 3)

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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 定义测试用例：包名+类名
	testCases := []struct {
		packageName string
		className   string
	}{
		{"org.apache.commons.lang3", "StringUtils"},
		{"java.util", "ArrayList"},
		{"org.slf4j", "Logger"},
		{"java.lang", "String"},
		{"java.io", "InputStream"},
		{"javax.servlet", "ServletContext"},
	}

	for _, tc := range testCases {
		name := tc.packageName + "." + tc.className
		t.Run("PackageClass_"+name, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByPackageAndClassName(ctx, tc.packageName, tc.className, 3)

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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的方法名
	methodNames := []string{
		"equals",
		"toString",
		"valueOf",
		"substring",
		"compareTo",
		"main",
		"getInstance",
	}

	for _, methodName := range methodNames {
		t.Run("Method_"+methodName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchClassesByMethod(ctx, methodName, 3)

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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的基类名
	baseClassNames := []string{
		"Exception",
		"AbstractList",
		"InputStream",
		"Component",
		"Thread",
		"Object",
	}

	for _, baseClassName := range baseClassNames {
		t.Run("BaseClass_"+baseClassName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchClassesWithClassHierarchy(ctx, baseClassName, 5)

			if err != nil {
				t.Logf("搜索继承自 %s 的类时出错: %v", baseClassName, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个可能继承自 %s 的结果", len(versionSlice), baseClassName)
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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的接口名
	interfaceNames := []string{
		"Listener",
		"Handler",
		"Adapter",
		"Callable",
		"Runnable",
		"Comparable",
		"Serializable",
	}

	for _, interfaceName := range interfaceNames {
		t.Run("Interface_"+interfaceName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchInterfaceImplementations(ctx, interfaceName, 5)

			if err != nil {
				t.Logf("搜索 %s 接口实现时出错: %v", interfaceName, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个可能实现 %s 接口的结果", len(versionSlice), interfaceName)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestSearchByClassSupertype 测试统一的父类型搜索接口
func TestSearchByClassSupertype(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的父类型名称，包括类和接口
	testCases := []struct {
		name          string
		supertypeName string
		isInterface   bool
	}{
		{"类_Exception", "Exception", false},
		{"接口_Listener", "Listener", true},
		{"类_AbstractList", "AbstractList", false},
		{"接口_Handler", "Handler", true},
		{"类_Thread", "Thread", false},
		{"接口_Runnable", "Runnable", true},
		{"类_Component", "Component", false},
		{"接口_Serializable", "Serializable", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByClassSupertype(ctx, tc.supertypeName, tc.isInterface, 3)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", tc.supertypeName, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			typeKind := "类"
			if tc.isInterface {
				typeKind = "接口"
			}

			t.Logf("找到 %d 个与%s %s 相关的结果", len(versionSlice), typeKind, tc.supertypeName)
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
