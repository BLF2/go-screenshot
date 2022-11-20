package service

import (
	"github.com/sirupsen/logrus"
	"screenshot-go/src/config"
	"screenshot-go/src/util"
)

func CaptureConcurrent(fileNameUrlMap map[string]string, bucketName string, path string, endPoint string, accessKeyId string,
	accessKeySecret string, goTraceId string) error {

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infoln("进入并行截图流程")

	//每6个文件分组
	splitMapList := util.SplitMap(fileNameUrlMap, 6)
	next := splitMapList.Front()
	for {
		if next == nil {
			break
		}
		subTraceId := util.GenerateUuidSample()
		parentAndSubTraceId := goTraceId + "->" + subTraceId
		//使用goruntime开启多协程进行截图，提高效率
		go Capture(next.Value.(map[string]string), bucketName, path, endPoint, accessKeyId, accessKeySecret, parentAndSubTraceId)
		next = next.Next()
	}
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).WithFields(logrus.Fields{
		config.GoTraceId: goTraceId,
	}).Infoln("并行截图提交完成")
	return nil
}

func CaptureLocalConcurrent(fileNameUrlMap map[string]string, localPath string, goTraceId string) error {

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infoln("进入并行截图流程Local")
	splitMapList := util.SplitMap(fileNameUrlMap, 1)
	next := splitMapList.Front()
	for {
		if next == nil {
			break
		}
		subTraceId := util.GenerateUuidSample()
		parentAndSubTraceId := goTraceId + "->" + subTraceId
		go CaptureLocal(next.Value.(map[string]string), localPath, parentAndSubTraceId)
		next = next.Next()
	}
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infoln("并行截图提交完成Local")
	return nil
}

func Capture(fileNameUrlMap map[string]string, bucketName string, path string, endPoint string, accessKeyId string,
	accessKeySecret string, goTraceId string) {

	//我要爬取的网站可以显示的第一行就是下面这个divId
	var topDivId = "headId"
	//下面这个是个ByID的选择器，等到出现这个选择器就认为加载完成了
	//可以使用chrome的开发模式 右键点击这个div->复制->选择器来获取
	var waitVisibleExpr = "#bodyDivId"
	//我的这个div上存放着当前页面的最大高度
	var maxHighId = "maxHighId"
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("开始调用pdf生成逻辑")
	fileNameBytesMap := ScreenPdf(fileNameUrlMap, topDivId, waitVisibleExpr,
		maxHighId, config.ChromeCtx, goTraceId)
	if len(fileNameBytesMap) == 0 {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Warnf("截图返回的pdf信息为空 本次截图完成")
		return
	}
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("截图并生成PDF完成，开始调用上传OSS文件")
	for key, bytes := range fileNameBytesMap {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("文件信息 fileName=%s byteCnt=%d", key, len(bytes))
	}

	PutFiles(bucketName, path, endPoint, accessKeyId, accessKeySecret, fileNameBytesMap, goTraceId)
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("截图生成PDF并上传OSS完成")
}

func CaptureLocal(fileNameUrlMap map[string]string, localPath string, goTraceId string) {

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("开始调用pdf生成逻辑")
	//我要爬取的网站可以显示的第一行就是下面这个divId
	var topDivId = "headId"
	//下面这个是个ByID的选择器，等到出现这个选择器就认为加载完成了
	//可以使用chrome的开发模式 右键点击这个div->复制->选择器来获取
	var waitVisibleExpr = "#bodyDivId"
	//我的这个div上存放着当前页面的最大高度
	var maxHighId = "maxHighId"
	fileNameBytesMap := ScreenPdf(fileNameUrlMap, topDivId, waitVisibleExpr, maxHighId, config.ChromeCtx, goTraceId)
	if len(fileNameBytesMap) == 0 {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Warnf("截图返回的pdf信息为空 本次截图完成")
		return
	}
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("截图并生成PDF完成，开始调用上传OSS文件")
	for key, bytes := range fileNameBytesMap {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("文件信息 fileName=%s byteCnt=%d", key, len(bytes))
	}
	ExportLocalFile(fileNameBytesMap, localPath, goTraceId)
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("截图生成PDF并上传OSS完成")
}
