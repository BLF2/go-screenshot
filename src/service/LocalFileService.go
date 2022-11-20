package service

import (
	"github.com/sirupsen/logrus"
	"os"
	"screenshot-go/src/config"
	"time"
)

func ExportLocalFile(fileNameBuff map[string][]byte, localPath string, goTraceId string) error {

	for fileName := range fileNameBuff {
		var pathFull = localPath + "/" + time.Now().Format("2006-01-02")
		_, pathNotExistsError := os.Stat(pathFull)
		if os.IsNotExist(pathNotExistsError) {
			//创建
			pathCreateError := os.Mkdir(pathFull, os.ModePerm)
			if pathCreateError != nil {
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("创建本地文件失败 pathFull=%s ", pathFull)
				continue
			}
		}
		var pathAndFileName = pathFull + "/" + fileName
		file, err := os.OpenFile(pathAndFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorln(err)
			continue
		}

		file.Write(fileNameBuff[fileName])
		file.Close()
	}

	return nil
}
