package service

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"screenshot-go/src/config"
	"time"
)

func GetOssClient(endPoint string, accessKeyId string, accessKeySecret string) (*oss.Client, error) {

	client, err := oss.New(endPoint, accessKeyId, accessKeySecret, oss.Timeout(10, 120))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func PutFile(bucketName string, path string, fileName string, pdfBytes []byte, client *oss.Client, goTraceId string) error {

	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	var pathAndFileName = path + "/" + time.Now().Format("2006-01-02") + "/" + fileName
	reader := bytes.NewReader(pdfBytes)
	// 上传文件。
	err = bucket.PutObject(pathAndFileName, reader)
	if err != nil {
		return err
	}
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).
		Infof("上传oss文件成功 bucketName=%s objectKey=%s pdfLength=%d",
			bucketName, pathAndFileName, len(pdfBytes))
	return nil
}

func PutFiles(bucketName string, path string, endPoint string, accessKeyId string,
	accessKeySecret string, fileNameBufMap map[string][]byte, goTraceId string) {

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("进入OSS上传逻辑 开始获取OSS client")
	client, err := GetOssClient(endPoint, accessKeyId, accessKeySecret)
	if err != nil {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("获取OSS client报错 直接返回 err=%v statck=%s", err, string(debug.Stack()))
		return
	}

	for fileName, bytes := range fileNameBufMap {
		err := PutFile(bucketName, path, fileName, bytes, client, goTraceId)
		if err != nil {
			config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("上传OSS文件失败 跳过当前文件上传 err=%v fileName=%s", err, fileName)
			continue
		}
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("上传OSS文件完成 fileName=%s byteCnt=%d", fileName, len(bytes))
	}
}
