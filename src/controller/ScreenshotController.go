package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"screenshot-go/src/config"
	"screenshot-go/src/dto"
	"screenshot-go/src/service"
	"strconv"
)

func InitController() {

	//关闭日志颜色
	gin.DisableConsoleColor()
	//创建一个gin实例
	engine := gin.New()
	//gin实例使用gin.Recovery()中间件，防止系统报错导致服务挂掉
	engine.Use(gin.Recovery())

	//定义一个请求group router 类似java的controller上的RequestMapping注解
	//并设定这个group router 使用日志中间件
	screenshotRouter := engine.Group("/screenshot-go").Use(config.LoggerMiddleware())
	//health check
	screenshotRouter.GET("/health", func(ctx *gin.Context) {

		ctx.JSON(http.StatusOK, dto.Success())
	})

	//截图 上传OSS
	screenshotRouter.POST("/capture_pic", func(ctx *gin.Context) {

		var capturePicReq dto.CapturePicReq
		ctx.BindJSON(&capturePicReq)
		goTraceId := ctx.GetString(config.GoTraceId)
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("接收到截图请求 req=%v\n", capturePicReq)
		if len(capturePicReq.PicInfo) == 0 {
			ctx.JSON(http.StatusOK, dto.Fail(strconv.Itoa(http.StatusBadRequest), "爬取url不能为空"))
			return
		}
		service.CaptureConcurrent(capturePicReq.PicInfo, capturePicReq.BucketName, capturePicReq.PicDir, capturePicReq.OssUrl,
			capturePicReq.AccessKeyId, capturePicReq.AccessKeySecret, goTraceId)
		ctx.JSON(http.StatusOK, dto.Success())
	})

	//截图 写入本地路径
	screenshotRouter.POST("/capture_localPic", func(ctx *gin.Context) {

		var capturePicReq dto.CapturePicReq
		ctx.BindJSON(&capturePicReq)
		goTraceId := ctx.GetString(config.GoTraceId)
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("接收到截图请求 req=%v\n", capturePicReq)
		//FIXME 这个为了测试  我写死了本地路径
		service.CaptureLocalConcurrent(capturePicReq.PicInfo, "/Users/blf2/Documents/workspace/golang/logs", goTraceId)
		ctx.JSON(http.StatusOK, dto.Success())
	})

	//测试日志使用
	engine.POST("/testLog", config.LoggerMiddleware(), func(ctx *gin.Context) {
		goTraceId := ctx.GetString(config.GoTraceId)
		bodyBytes, _ := ctx.GetRawData()
		body := string(bodyBytes)
		config.LogEntry.WithFields(logrus.Fields{
			config.GoTraceId: goTraceId,
		}).Infof("接收到日志格式测试 body=%s", body)
		ctx.JSON(http.StatusOK, dto.Success())
	})

	//测试是否可以启动chrome 一般配合headless=false一起使用
	engine.GET("/chromeTest", config.LoggerMiddleware(), func(ctx *gin.Context) {

		goTraceId := ctx.GetString(config.GoTraceId)
		service.TestStartChrome(config.ChromeCtx, goTraceId)
		ctx.JSON(http.StatusOK, dto.Success())
	})

	engine.Run(":8080")
}
