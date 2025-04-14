package api

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// 定义常用的文件扩展名
const (
	POM     = "pom"
	JAR     = "jar"
	WAR     = "war"
	AAR     = "aar"
	SOURCES = "sources.jar"
	JAVADOC = "javadoc.jar"
	TESTS   = "tests.jar"
)

// Download 从Maven中央仓库下载指定路径的文件
//
// 这是SDK中所有下载功能的核心方法。它通过访问Maven仓库的基础URL（默认为https://repo1.maven.org/maven2）
// 加上指定的文件路径来获取文件内容。该方法支持通过配置的Client参数进行重试、超时控制和缓存。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePath: 文件在Maven仓库中的相对路径，例如"com/google/guava/guava/31.1-jre/guava-31.1-jre.jar"
//
// 返回:
//   - []byte: 下载文件的二进制内容
//   - error: 如果下载过程中出现错误，如网络问题、文件不存在等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 下载特定版本的POM文件
//	pomPath := "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.pom"
//	pomData, err := client.Download(ctx, pomPath)
//	if err != nil {
//	    log.Fatalf("下载POM文件失败: %v", err)
//	}
//
//	// 使用下载的内容（例如保存到文件）
//	err = os.WriteFile("commons-lang3.pom", pomData, 0644)
//	if err != nil {
//	    log.Fatalf("保存POM文件失败: %v", err)
//	}
//
//	// 也可以使用BuildArtifactPath辅助函数构建文件路径
//	jarPath := api.BuildArtifactPath("org.apache.commons", "commons-lang3", "3.12.0", "jar")
//	jarData, err := client.Download(ctx, jarPath)
func (c *Client) Download(ctx context.Context, filePath string) ([]byte, error) {
	return c.downloadWithCache(ctx, filePath)
}

// DownloadFile 下载文件并直接保存到本地文件系统
//
// 该方法是Download的扩展，它不仅下载文件内容，还会自动创建必要的目录结构，
// 并将下载的内容写入指定的本地路径。这使得获取Maven制品并存储到本地变得非常简单，
// 无需手动处理文件的创建和写入操作。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePath: 文件在Maven仓库中的相对路径
//   - localPath: 要保存到的本地文件路径（绝对或相对路径）
//
// 返回:
//   - error: 如果下载或保存过程中出现任何错误，如网络问题、文件系统权限问题等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 下载特定版本的JAR文件并保存到本地
//	remoteJarPath := "org/apache/logging/log4j/log4j-core/2.17.1/log4j-core-2.17.1.jar"
//	localJarPath := "dependencies/log4j-core-2.17.1.jar"
//
//	err := client.DownloadFile(ctx, remoteJarPath, localJarPath)
//	if err != nil {
//	    log.Fatalf("下载并保存文件失败: %v", err)
//	}
//	fmt.Printf("成功下载文件到: %s\n", localJarPath)
//
//	// 使用BuildArtifactPath辅助函数构建远程路径
//	remotePomPath := api.BuildArtifactPath("org.apache.logging.log4j", "log4j-api", "2.17.1", "pom")
//	localPomPath := "dependencies/log4j-api-2.17.1.pom"
//
//	err = client.DownloadFile(ctx, remotePomPath, localPomPath)
//	if err != nil {
//	    log.Fatalf("下载并保存POM文件失败: %v", err)
//	}
func (c *Client) DownloadFile(ctx context.Context, filePath, localPath string) error {
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(localPath, data, 0644)
}

// DownloadToWriter 下载文件并直接写入到io.Writer接口
//
// 该方法允许将下载的内容直接写入任何实现了io.Writer接口的对象，而不是保存到文件系统。
// 这使得API更加灵活，支持将下载内容直接写入HTTP响应、内存缓冲区、压缩流或其他自定义writer。
// 当需要对下载内容进行进一步处理而不是直接保存文件时，这个方法特别有用。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePath: 文件在Maven仓库中的相对路径
//   - writer: 实现了io.Writer接口的对象，下载的内容将写入此对象
//
// 返回:
//   - error: 如果下载或写入过程中出现任何错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 示例1: 下载到内存缓冲区
//	var buffer bytes.Buffer
//	pomPath := api.BuildArtifactPath("com.google.guava", "guava", "31.1-jre", "pom")
//
//	err := client.DownloadToWriter(ctx, pomPath, &buffer)
//	if err != nil {
//	    log.Fatalf("下载到缓冲区失败: %v", err)
//	}
//
//	// 现在可以处理buffer中的内容
//	fmt.Printf("下载的POM文件大小: %d 字节\n", buffer.Len())
//
//	// 示例2: 下载并直接写入HTTP响应
//	// 在HTTP处理函数中:
//	// func handleDownload(w http.ResponseWriter, r *http.Request) {
//	//     client := api.NewClient()
//	//     jarPath := api.BuildArtifactPath("org.springframework", "spring-core", "5.3.20", "jar")
//	//
//	//     w.Header().Set("Content-Type", "application/java-archive")
//	//     w.Header().Set("Content-Disposition", "attachment; filename=spring-core-5.3.20.jar")
//	//
//	//     err := client.DownloadToWriter(r.Context(), jarPath, w)
//	//     if err != nil {
//	//         http.Error(w, "下载文件失败: "+err.Error(), http.StatusInternalServerError)
//	//         return
//	//     }
//	// }
func (c *Client) DownloadToWriter(ctx context.Context, filePath string, writer io.Writer) error {
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// BuildArtifactPath 构建Maven制品在仓库中的标准路径
//
// 该方法根据Maven坐标（groupId、artifactId、version）和文件类型（extension、classifier）
// 构建出制品在Maven仓库中的标准路径。它遵循Maven规范的目录结构和文件命名约定，
// 自动处理groupId中的点号转换为目录分隔符，并正确组合文件名中的分类器部分。
//
// 参数:
//   - groupId: 制品的组ID，如"org.springframework"
//   - artifactId: 制品的ID，如"spring-core"
//   - version: 制品的版本号，如"5.3.20"
//   - extension: 文件扩展名，如"jar"、"pom"、"war"等
//   - classifier: 可选的分类器，如"sources"、"javadoc"、"tests"等；可以省略
//
// 返回:
//   - string: 构建好的相对路径，可直接用于Download方法
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 构建POM文件路径
//	pomPath := api.BuildArtifactPath("org.apache.commons", "commons-lang3", "3.12.0", "pom")
//	// 结果: "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.pom"
//
//	// 构建JAR文件路径
//	jarPath := api.BuildArtifactPath("org.apache.commons", "commons-lang3", "3.12.0", "jar")
//	// 结果: "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar"
//
//	// 构建带分类器的文件路径（源码JAR）
//	sourcesPath := api.BuildArtifactPath("org.apache.commons", "commons-lang3", "3.12.0", "jar", "sources")
//	// 结果: "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0-sources.jar"
//
//	// 使用构建的路径下载文件
//	pom, err := client.Download(ctx, pomPath)
//	if err != nil {
//	    log.Fatalf("下载POM失败: %v", err)
//	}
func BuildArtifactPath(groupId, artifactId, version, extension string, classifier ...string) string {
	basePath := fmt.Sprintf("%s/%s/%s", strings.ReplaceAll(groupId, ".", "/"), artifactId, version)

	fileName := fmt.Sprintf("%s-%s", artifactId, version)

	// 添加分类器
	if len(classifier) > 0 && classifier[0] != "" {
		fileName += "-" + classifier[0]
	}

	// 添加扩展名
	fileName += "." + extension

	return fmt.Sprintf("%s/%s", basePath, fileName)
}

// DownloadPom 下载Maven项目的POM文件
//
// 该方法是对Download方法的便捷封装，专门用于下载Maven项目的POM（Project Object Model）文件。
// POM文件是Maven项目的核心配置文件，包含项目信息、依赖关系、构建设置等元数据。
// 该方法自动构建正确的POM文件路径，简化了下载过程。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"org.apache.commons"
//   - artifactId: 制品的ID，如"commons-lang3"
//   - version: 制品的版本号，如"3.12.0"
//
// 返回:
//   - []byte: 下载的POM文件内容，为XML格式的字节数组
//   - error: 如果下载过程中出现错误，如网络问题、文件不存在等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 下载Apache Commons Lang的POM文件
//	pomData, err := client.DownloadPom(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
//	if err != nil {
//	    log.Fatalf("下载POM文件失败: %v", err)
//	}
//
//	// 使用下载的POM内容（例如解析XML或保存到文件）
//	fmt.Printf("POM文件大小: %d 字节\n", len(pomData))
//
//	// 保存到本地文件
//	err = os.WriteFile("commons-lang3.pom", pomData, 0644)
//	if err != nil {
//	    log.Fatalf("保存POM文件失败: %v", err)
//	}
//
//	// 解析POM XML内容（示例）
//	// var pomModel maven.Model
//	// err = xml.Unmarshal(pomData, &pomModel)
//	// if err != nil {
//	//     log.Fatalf("解析POM内容失败: %v", err)
//	// }
//	// fmt.Printf("项目名称: %s\n", pomModel.Name)
func (c *Client) DownloadPom(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, POM)
	return c.Download(ctx, path)
}

// DownloadJar 下载Maven项目的JAR文件
//
// 该方法是对Download方法的便捷封装，专门用于下载Maven项目的JAR（Java Archive）文件。
// JAR文件是Java应用程序的标准打包格式，包含编译后的类文件、资源文件和元数据。
// 该方法自动构建正确的JAR文件路径，简化了下载过程。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"com.google.guava"
//   - artifactId: 制品的ID，如"guava"
//   - version: 制品的版本号，如"31.1-jre"
//
// 返回:
//   - []byte: 下载的JAR文件二进制内容
//   - error: 如果下载过程中出现错误，如网络问题、文件不存在等
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 下载Google Guava库的JAR文件
//	jarData, err := client.DownloadJar(ctx, "com.google.guava", "guava", "31.1-jre")
//	if err != nil {
//	    log.Fatalf("下载JAR文件失败: %v", err)
//	}
//
//	// 查看JAR文件大小
//	fmt.Printf("JAR文件大小: %.2f MB\n", float64(len(jarData))/1024/1024)
//
//	// 保存到本地文件
//	err = os.WriteFile("guava-31.1-jre.jar", jarData, 0644)
//	if err != nil {
//	    log.Fatalf("保存JAR文件失败: %v", err)
//	}
//
//	// 将JAR添加到类路径（伪代码）
//	// classLoader.addJarToClasspath(jarData)
//
//	// 或者保存到特定目录以用于Java项目
//	libDir := "lib"
//	if err := os.MkdirAll(libDir, 0755); err == nil {
//	    jarPath := filepath.Join(libDir, "guava-31.1-jre.jar")
//	    _ = os.WriteFile(jarPath, jarData, 0644)
//	}
func (c *Client) DownloadJar(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR)
	return c.Download(ctx, path)
}

// DownloadSources 下载指定Maven坐标的源代码包
//
// 该方法是下载源代码包的便捷方法，它会根据提供的Maven坐标构建标准路径，
// 然后从仓库中下载源代码包。源代码包通常包含库的完整源代码，方便开发人员
// 在开发过程中查看和调试。
//
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - groupId: Maven坐标的组ID，例如"com.google.guava"
//   - artifactId: Maven坐标的制品ID，例如"guava"
//   - version: Maven坐标的版本号，例如"31.1-jre"
//
// 返回:
//   - []byte: 下载的源代码包内容
//   - error: 如果下载失败，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 下载Guava库的源代码
//	sourceData, err := client.DownloadSources(ctx, "com.google.guava", "guava", "31.1-jre")
//	if err != nil {
//	    log.Fatalf("下载源代码失败: %v", err)
//	}
//
//	// 查看源代码包大小
//	fmt.Printf("源代码包大小: %.2f MB\n", float64(len(sourceData))/1024/1024)
//
//	// 保存到本地文件
//	err = os.WriteFile("guava-31.1-jre-sources.jar", sourceData, 0644)
//	if err != nil {
//	    log.Fatalf("保存源代码包失败: %v", err)
//	}
//
//	// 解压源代码包以便查看源码（伪代码）
//	// srcDir := "src"
//	// if err := os.MkdirAll(srcDir, 0755); err == nil {
//	//     jar := bytes.NewReader(sourceData)
//	//     unzip(jar, srcDir) // 将JAR包解压到src目录
//	//     fmt.Println("源代码已解压到：", srcDir)
//	// }
//
//	// 或者将源代码附加到IDE项目中（伪代码）
//	// ide.attachSources("com.google.guava:guava:31.1-jre", sourceData)
func (c *Client) DownloadSources(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR, "sources")
	return c.Download(ctx, path)
}

// DownloadJavadoc 下载指定Maven坐标的JavaDoc文档包
//
// 该方法是下载JavaDoc文档包的便捷方法，它会根据提供的Maven坐标构建标准路径，
// 然后从仓库中下载JavaDoc文档包。JavaDoc包通常包含格式化的HTML文档，描述了
// 库中的类、方法和字段的用法。
//
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - groupId: Maven坐标的组ID，例如"org.springframework"
//   - artifactId: Maven坐标的制品ID，例如"spring-core"
//   - version: Maven坐标的版本号，例如"5.3.23"
//
// 返回:
//   - []byte: 下载的JavaDoc文档包内容
//   - error: 如果下载失败，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 下载Spring Framework核心模块的JavaDoc文档
//	javadocData, err := client.DownloadJavadoc(ctx, "org.springframework", "spring-core", "5.3.23")
//	if err != nil {
//	    log.Fatalf("下载JavaDoc文档失败: %v", err)
//	}
//
//	// 查看JavaDoc文档包大小
//	fmt.Printf("JavaDoc文档包大小: %.2f MB\n", float64(len(javadocData))/1024/1024)
//
//	// 保存到本地文件
//	err = os.WriteFile("spring-core-5.3.23-javadoc.jar", javadocData, 0644)
//	if err != nil {
//	    log.Fatalf("保存JavaDoc文档包失败: %v", err)
//	}
//
//	// 解压JavaDoc文档包以便在浏览器中查看（伪代码）
//	// docDir := "docs"
//	// if err := os.MkdirAll(docDir, 0755); err == nil {
//	//     jar := bytes.NewReader(javadocData)
//	//     unzip(jar, docDir) // 将JAR包解压到docs目录
//	//     fmt.Println("可在浏览器中打开：", docDir + "/index.html")
//	// }
//
//	// 或者集成到IDE中查看JavaDoc（伪代码）
//	// ide.attachJavadoc("org.springframework:spring-core:5.3.23", javadocData)
func (c *Client) DownloadJavadoc(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR, "javadoc")
	return c.Download(ctx, path)
}

// DownloadArtifact 根据指定参数下载制品
//
// 该方法提供了一种根据Artifact对象下载制品的便捷方式，支持指定文件扩展名和可选的分类器。
// 它使用Artifact对象中的GroupId、ArtifactId和LatestVersion信息构建下载路径，
// 使得从已获取的搜索结果中直接下载制品变得简单。
//
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - artifact: 包含制品坐标信息的Artifact对象，通常是搜索结果的一部分
//   - extension: 文件扩展名，如"jar"、"pom"等
//   - classifier: 可选的分类器，如"sources"、"javadoc"等
//
// 返回:
//   - []byte: 下载的制品文件内容
//   - error: 如果下载失败，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 假设我们已经通过搜索获得了Artifact对象
//	artifact := &response.Artifact{
//	    GroupId:       "org.apache.commons",
//	    ArtifactId:    "commons-lang3",
//	    LatestVersion: "3.12.0",
//	}
//
//	// 下载JAR文件
//	jarData, err := client.DownloadArtifact(ctx, artifact, "jar")
//	if err != nil {
//	    log.Fatalf("下载JAR文件失败: %v", err)
//	}
//
//	// 下载源代码文件
//	sourcesData, err := client.DownloadArtifact(ctx, artifact, "jar", "sources")
//	if err != nil {
//	    log.Fatalf("下载源代码文件失败: %v", err)
//	}
//
//	// 处理下载的文件，例如保存到本地
//	err = os.WriteFile("commons-lang3-3.12.0.jar", jarData, 0644)
//	if err != nil {
//	    log.Fatalf("保存文件失败: %v", err)
//	}
func (c *Client) DownloadArtifact(ctx context.Context, artifact *response.Artifact, extension string, classifier ...string) ([]byte, error) {
	path := BuildArtifactPath(artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion, extension, classifier...)
	return c.Download(ctx, path)
}

// DownloadArtifactWithVersion 根据指定版本下载制品
//
// 该方法提供了一种根据Version对象下载制品的便捷方式，支持指定文件扩展名和可选的分类器。
// 它使用Version对象中的GroupId、ArtifactId和Version信息构建下载路径，
// 尤其适用于从版本搜索结果中直接下载指定版本的制品。
//
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - artifact: 包含制品版本信息的Version对象，通常是版本搜索结果的一部分
//   - extension: 文件扩展名，如"jar"、"pom"等
//   - classifier: 可选的分类器，如"sources"、"javadoc"等
//
// 返回:
//   - []byte: 下载的制品文件内容
//   - error: 如果下载失败，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 假设我们已经通过版本搜索获得了Version对象
//	version := &response.Version{
//	    GroupId:    "org.apache.commons",
//	    ArtifactId: "commons-lang3",
//	    Version:    "3.11.0", // 特定版本
//	}
//
//	// 下载特定版本的JAR文件
//	jarData, err := client.DownloadArtifactWithVersion(ctx, version, "jar")
//	if err != nil {
//	    log.Fatalf("下载JAR文件失败: %v", err)
//	}
//
//	// 下载特定版本的源代码
//	sourcesData, err := client.DownloadArtifactWithVersion(ctx, version, "jar", "sources")
//	if err != nil {
//	    log.Fatalf("下载源代码文件失败: %v", err)
//	}
//
//	// 处理下载的文件，例如保存到本地
//	err = os.WriteFile("commons-lang3-3.11.0.jar", jarData, 0644)
//	if err != nil {
//	    log.Fatalf("保存文件失败: %v", err)
//	}
func (c *Client) DownloadArtifactWithVersion(ctx context.Context, artifact *response.Version, extension string, classifier ...string) ([]byte, error) {
	path := BuildArtifactPath(artifact.GroupId, artifact.ArtifactId, artifact.Version, extension, classifier...)
	return c.Download(ctx, path)
}

// ArtifactFile 表示一个制品文件类型
type ArtifactFile struct {
	Type       string // 文件类型的标识，如"pom", "jar"等
	Extension  string // 文件扩展名
	Classifier string // 可选的分类器
}

// 预定义的常用制品文件类型
var (
	PomFile     = ArtifactFile{Type: "POM", Extension: POM}
	JarFile     = ArtifactFile{Type: "JAR", Extension: JAR}
	SourcesFile = ArtifactFile{Type: "SOURCES", Extension: JAR, Classifier: "sources"}
	JavadocFile = ArtifactFile{Type: "JAVADOC", Extension: JAR, Classifier: "javadoc"}
	TestsFile   = ArtifactFile{Type: "TESTS", Extension: JAR, Classifier: "tests"}
	WarFile     = ArtifactFile{Type: "WAR", Extension: WAR}
	AarFile     = ArtifactFile{Type: "AAR", Extension: AAR}
)

// CommonArtifactFiles 返回常用的制品文件类型列表
func CommonArtifactFiles() []ArtifactFile {
	return []ArtifactFile{
		PomFile, JarFile, SourcesFile, JavadocFile,
	}
}

// DownloadResult 表示一个下载结果
type DownloadResult struct {
	FileType ArtifactFile // 文件类型
	Data     []byte       // 文件数据
	Error    error        // 下载过程中的错误，如果有的话
	Path     string       // 文件在仓库中的路径
	SHA1     string       // SHA1摘要，仅当验证时可用
	MD5      string       // MD5摘要，仅当验证时可用
	SHA256   string       // SHA256摘要，仅当验证时可用
}

// DownloadMultipleFiles 并行下载多个不同类型的文件
//
// 此方法允许同时下载多个与指定制品相关的文件，如POM、JAR、源代码和JavaDoc文档。
// 所有下载操作都在独立的goroutine中并行执行，以提高效率。
//
// 参数:
//   - ctx: 上下文对象，可用于取消或设置超时
//   - groupId: Maven坐标中的组ID
//   - artifactId: Maven坐标中的制品ID
//   - version: 制品的版本号
//   - fileTypes: 要下载的文件类型列表，使用ArtifactFile类型指定
//
// 返回:
//   - map[string]*DownloadResult: 以文件类型标识为键的下载结果映射，每个结果包含文件数据和可能的错误
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 定义要下载的文件类型
//	fileTypes := []ArtifactFile{
//	    PomFile,
//	    JarFile,
//	    SourcesFile,
//	    JavadocFile,
//	}
//
//	// 并行下载多个文件
//	results := client.DownloadMultipleFiles(ctx, "org.apache.commons", "commons-lang3", "3.12.0", fileTypes)
//
//	// 处理下载结果
//	for fileType, result := range results {
//	    if result.Error != nil {
//	        fmt.Printf("下载%s失败: %v\n", fileType, result.Error)
//	        continue
//	    }
//
//	    fmt.Printf("成功下载%s，大小: %d字节\n", fileType, len(result.Data))
//	    // 处理下载的数据...
//	}
func (c *Client) DownloadMultipleFiles(ctx context.Context, groupId, artifactId, version string, fileTypes []ArtifactFile) map[string]*DownloadResult {
	results := make(map[string]*DownloadResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, fileType := range fileTypes {
		wg.Add(1)
		go func(ft ArtifactFile) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			result := &DownloadResult{
				FileType: ft,
				Data:     data,
				Error:    err,
				Path:     path,
			}

			mu.Lock()
			results[ft.Type] = result
			mu.Unlock()
		}(fileType)
	}

	wg.Wait()
	return results
}

// DownloadWithChecksum 下载文件并验证其校验和
//
// 此方法不仅下载指定的文件，还会计算其校验和并尝试与远程仓库中的校验和文件进行比对验证。
// 支持SHA1、MD5和SHA256三种校验和类型。如果远程仓库中存在相应的校验和文件，将进行比对；
// 如果不存在，则只返回计算出的校验和。
//
// 参数:
//   - ctx: 上下文对象，可用于取消或设置超时
//   - filePath: 要下载的文件在仓库中的相对路径
//   - checksumType: 校验和类型，支持"sha1"、"md5"、"sha256"
//
// 返回:
//   - []byte: 下载的文件内容
//   - string: 计算出的校验和（十六进制字符串）
//   - error: 如果下载失败或校验和不匹配，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 构建文件路径
//	filePath := "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar"
//
//	// 下载文件并验证SHA1校验和
//	data, checksum, err := client.DownloadWithChecksum(ctx, filePath, "sha1")
//	if err != nil {
//	    log.Fatalf("下载或验证文件失败: %v", err)
//	}
//
//	fmt.Printf("成功下载文件，大小: %d字节, SHA1: %s\n", len(data), checksum)
//
//	// 保存文件
//	err = os.WriteFile("commons-lang3-3.12.0.jar", data, 0644)
//	if err != nil {
//	    log.Fatalf("保存文件失败: %v", err)
//	}
func (c *Client) DownloadWithChecksum(ctx context.Context, filePath string, checksumType string) ([]byte, string, error) {
	// 下载文件
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return nil, "", err
	}

	// 计算校验和
	var checksum string
	switch strings.ToLower(checksumType) {
	case "sha1":
		hash := sha1.Sum(data)
		checksum = hex.EncodeToString(hash[:])
	case "md5":
		hash := md5.Sum(data)
		checksum = hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256(data)
		checksum = hex.EncodeToString(hash[:])
	default:
		return data, "", fmt.Errorf("不支持的校验和类型: %s", checksumType)
	}

	// 下载对应的校验和文件进行验证
	checksumFilePath := filePath + "." + checksumType
	checksumFileData, err := c.Download(ctx, checksumFilePath)

	// 如果校验和文件不存在，直接返回计算出的校验和
	if err != nil {
		return data, checksum, nil
	}

	// 校验和文件通常只包含十六进制字符串
	remoteChecksum := strings.TrimSpace(string(checksumFileData))

	// 有时校验和文件可能包含文件名，需要提取第一段
	if parts := strings.Fields(remoteChecksum); len(parts) > 0 {
		remoteChecksum = parts[0]
	}

	// 比较校验和
	if checksum != remoteChecksum {
		return data, checksum, fmt.Errorf("校验和不匹配: 计算得到 %s，远程值 %s", checksum, remoteChecksum)
	}

	return data, checksum, nil
}

// DownloadProgress 下载进度回调函数
type DownloadProgress func(downloaded, total int64, fileName string)

// AsyncDownloadResult 异步下载结果
type AsyncDownloadResult struct {
	FileType ArtifactFile
	Data     []byte
	Error    error
	Path     string
}

// DownloadAsync 异步下载文件
//
// 此方法提供了异步下载能力，立即返回一个通道，下载完成后会通过该通道发送结果。
// 这对于需要在后台下载大文件而不阻塞主程序执行的场景特别有用。
//
// 参数:
//   - ctx: 上下文对象，可用于取消或设置超时
//   - filePath: 要下载的文件在仓库中的相对路径
//
// 返回:
//   - <-chan AsyncDownloadResult: 接收下载结果的通道，当下载完成时会发送结果
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 构建文件路径
//	filePath := "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar"
//
//	// 启动异步下载
//	resultChan := client.DownloadAsync(ctx, filePath)
//
//	// 继续执行其他任务...
//	// ...
//
//	// 在适当的时候获取下载结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("下载失败: %v", result.Error)
//	}
//
//	fmt.Printf("成功下载文件，大小: %d字节\n", len(result.Data))
//	// 处理下载的数据...
func (c *Client) DownloadAsync(ctx context.Context, filePath string) <-chan AsyncDownloadResult {
	resultChan := make(chan AsyncDownloadResult, 1)

	go func() {
		defer close(resultChan)

		data, err := c.Download(ctx, filePath)
		result := AsyncDownloadResult{
			Data:  data,
			Error: err,
			Path:  filePath,
		}

		resultChan <- result
	}()

	return resultChan
}

// ArtifactBundle 表示一个制品包，包含所有相关文件
type ArtifactBundle struct {
	GroupId    string
	ArtifactId string
	Version    string
	Pom        []byte
	Jar        []byte
	Sources    []byte
	Javadoc    []byte
	Tests      []byte
	OtherFiles map[string][]byte
	Errors     map[string]error
}

// DownloadCompleteBundle 下载制品的完整包，包括所有可用的相关文件
//
// 此方法尝试下载与指定制品相关的所有标准文件（POM、JAR、源代码、JavaDoc和测试）以及
// 任何指定的额外文件类型。所有下载操作并行执行以提高效率。即使某些文件下载失败，
// 方法仍会返回包含所有成功下载文件的捆绑包，同时记录错误信息。
//
// 参数:
//   - ctx: 上下文对象，可用于取消或设置超时
//   - groupId: Maven坐标中的组ID
//   - artifactId: Maven坐标中的制品ID
//   - version: 制品的版本号
//   - extraFiles: 可选的额外文件类型列表，用于下载标准文件之外的其他文件
//
// 返回:
//   - *ArtifactBundle: 包含所有成功下载文件的捆绑包
//   - error: 仅当所有必要文件（POM和JAR）都下载失败时返回错误
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 下载完整的制品包，包括一些额外文件
//	bundle, err := client.DownloadCompleteBundle(ctx,
//	    "org.apache.commons",
//	    "commons-lang3",
//	    "3.12.0",
//	    ArtifactFile{Type: "EXAMPLES", Extension: "jar", Classifier: "examples"})
//
//	if err != nil {
//	    log.Fatalf("下载制品包失败: %v", err)
//	}
//
//	// 查看下载结果
//	fmt.Printf("POM文件大小: %d字节\n", len(bundle.Pom))
//	fmt.Printf("JAR文件大小: %d字节\n", len(bundle.Jar))
//	fmt.Printf("源代码文件大小: %d字节\n", len(bundle.Sources))
//	fmt.Printf("JavaDoc文件大小: %d字节\n", len(bundle.Javadoc))
//
//	// 检查是否有下载错误
//	for fileType, err := range bundle.Errors {
//	    fmt.Printf("下载%s失败: %v\n", fileType, err)
//	}
func (c *Client) DownloadCompleteBundle(ctx context.Context, groupId, artifactId, version string, extraFiles ...ArtifactFile) (*ArtifactBundle, error) {
	// 创建基本的bundle结构
	bundle := &ArtifactBundle{
		GroupId:    groupId,
		ArtifactId: artifactId,
		Version:    version,
		OtherFiles: make(map[string][]byte),
		Errors:     make(map[string]error),
	}

	// 必要文件列表
	essentialFiles := []struct {
		fileType ArtifactFile
		target   *[]byte
		name     string
	}{
		{PomFile, &bundle.Pom, "POM"},
		{JarFile, &bundle.Jar, "JAR"},
		{SourcesFile, &bundle.Sources, "SOURCES"},
		{JavadocFile, &bundle.Javadoc, "JAVADOC"},
		{TestsFile, &bundle.Tests, "TESTS"},
	}

	// 下载必要文件
	var wg sync.WaitGroup
	var mu sync.Mutex
	var essentialErrors int = 0

	for _, file := range essentialFiles {
		wg.Add(1)
		go func(ft ArtifactFile, target *[]byte, name string) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				bundle.Errors[name] = err
				if name == "POM" || name == "JAR" {
					essentialErrors++
				}
			} else {
				*target = data
			}
		}(file.fileType, file.target, file.name)
	}

	// 下载额外文件
	for _, extraFile := range extraFiles {
		wg.Add(1)
		go func(ft ArtifactFile) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				bundle.Errors[ft.Type] = err
			} else {
				bundle.OtherFiles[ft.Type] = data
			}
		}(extraFile)
	}

	wg.Wait()

	// 检查是否所有必要文件都下载失败
	if essentialErrors >= 2 {
		return bundle, errors.New("必要文件（POM和JAR）下载失败")
	}

	return bundle, nil
}

// SaveBundle 将制品包保存到本地目录
//
// 此方法将下载的制品包保存到本地文件系统，遵循Maven仓库的标准目录结构
// （groupId/artifactId/version）。它会自动创建必要的目录结构，并仅保存成功
// 下载的文件。
//
// 参数:
//   - bundle: 要保存的制品包，通常是DownloadCompleteBundle方法的返回结果
//   - baseDir: 保存文件的基础目录路径
//
// 返回:
//   - error: 如果创建目录或写入文件过程中发生错误，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 下载完整的制品包
//	bundle, err := client.DownloadCompleteBundle(ctx,
//	    "org.apache.commons",
//	    "commons-lang3",
//	    "3.12.0")
//
//	if err != nil {
//	    log.Fatalf("下载制品包失败: %v", err)
//	}
//
//	// 将制品包保存到本地Maven仓库结构
//	err = client.SaveBundle(bundle, "/path/to/local/repository")
//	if err != nil {
//	    log.Fatalf("保存制品包失败: %v", err)
//	}
//
//	fmt.Println("制品包已成功保存到本地仓库")
func (c *Client) SaveBundle(bundle *ArtifactBundle, baseDir string) error {
	// 构建目标目录
	targetDir := filepath.Join(baseDir,
		strings.ReplaceAll(bundle.GroupId, ".", string(filepath.Separator)),
		bundle.ArtifactId,
		bundle.Version)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// 保存文件
	filesToSave := []struct {
		data     []byte
		fileName string
	}{
		{bundle.Pom, fmt.Sprintf("%s-%s.pom", bundle.ArtifactId, bundle.Version)},
		{bundle.Jar, fmt.Sprintf("%s-%s.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Sources, fmt.Sprintf("%s-%s-sources.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Javadoc, fmt.Sprintf("%s-%s-javadoc.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Tests, fmt.Sprintf("%s-%s-tests.jar", bundle.ArtifactId, bundle.Version)},
	}

	for _, file := range filesToSave {
		if file.data == nil || len(file.data) == 0 {
			continue
		}

		filePath := filepath.Join(targetDir, file.fileName)
		if err := os.WriteFile(filePath, file.data, 0644); err != nil {
			return err
		}
	}

	// 保存其他文件
	for fileType, data := range bundle.OtherFiles {
		fileName := fmt.Sprintf("%s-%s-%s.%s",
			bundle.ArtifactId, bundle.Version, strings.ToLower(fileType), "jar")
		filePath := filepath.Join(targetDir, fileName)

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
