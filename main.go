package main

import (
	"log"
	"net/http"
	"os"

	"mogost-tools/tools"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	log.Println("正在启动 Mogost 工具集服务器...")

	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 主页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Mogost 工具集",
		})
	})

	// 工具1：文件比较
	fileCompare := r.Group("/api/file-compare")
	{
		fileCompare.POST("/upload", tools.HandleFileCompareUpload)
		fileCompare.GET("/compare", tools.HandleFileCompare)
	}

	// 工具2：CSV查看器
	csvViewer := r.Group("/api/csv")
	{
		csvViewer.POST("/upload", tools.HandleCSVUpload)
		csvViewer.GET("/view", tools.HandleCSVView)
	}

	// 工具3：压缩文件解压和交易文件比较
	archiveCompare := r.Group("/api/archive-compare")
	{
		archiveCompare.POST("/upload", tools.HandleArchiveUpload)
		archiveCompare.GET("/compare", tools.HandleArchiveCompare)
	}

	// 创建必要的目录
	createDirectories()

	log.Println("Mogost 工具集服务器启动在 :8080")
	log.Println("请在浏览器中访问: http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}

func createDirectories() {
	dirs := []string{
		"uploads",
		"uploads/file-compare",
		"uploads/csv",
		"uploads/archive-compare",
		"static",
		"templates",
		"temp",
	}

	log.Println("正在创建必要目录...")
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("创建目录失败 %s: %v", dir, err)
		} else {
			log.Printf("创建目录成功: %s", dir)
		}
	}
}
