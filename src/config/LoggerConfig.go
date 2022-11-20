package config

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"screenshot-go/src/util"
	"time"
)

var (
	// 初始化日志对象
	Logger = logrus.New()
	//带输出字段的entry
	LogEntry *logrus.Entry
)

func init() {

	//获取配置文件中的信息
	logPath := os.Getenv("logDir") // 日志存放路径
	logFileName := "screenshot"
	logFileSuffix := "log"
	pathAndFileName := path.Join(logPath, logFileName+"."+logFileSuffix)
	logFile, err := os.OpenFile(pathAndFileName, os.O_RDWR|os.O_CREATE, 0755) // 初始化日志文件对象
	if err != nil {
		fmt.Println("%v ,%s", err, string(debug.Stack()))
		panic("create log file error")
	}

	//设置自定义格式的日志输出 方便和java一样分割和索引日志
	Logger.SetFormatter(new(JavaStyleLogFormatter))
	//显示行号 便于查问题
	Logger.SetReportCaller(true)
	//设置日志的输出文件
	Logger.Out = logFile
	// 日志分隔：
	//1. 每天产生的日志写在不同的文件；
	//2. 只保留一定时间的日志（例如：一星期）

	// 设置日志级别
	Logger.SetLevel(logrus.InfoLevel)
	logWriter, rotateError := rotatelogs.New(
		// 日志文件名格式
		path.Join(logPath, logFileName+"_%Y%m%d"+"."+logFileSuffix),
		// 最多保留7天之内的日志
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 一天保存一个日志文件
		rotatelogs.WithRotationTime(24*time.Hour),
		//软连接 screenshot.log永远指向最新的那一天的
		rotatelogs.WithLinkName(pathAndFileName),
	)

	if rotateError != nil {
		fmt.Println("%v %s", rotateError, string(debug.Stack()))
		panic("设置日志格式和分割配置失败")
	}

	// 所有级别的日志都使用logWriter写日志
	writeMap := lfshook.WriterMap{
		logrus.DebugLevel: logWriter,
		logrus.InfoLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.FatalLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}
	//初始化钩子 格式也设置成java形式的
	Hook := lfshook.NewHook(writeMap, new(JavaStyleLogFormatter))
	Logger.AddHook(Hook)
	//初始化logEntry
	LogEntry = logrus.NewEntry(Logger)
}

/*
gin使用的日志中间件
*/
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		url := c.Request.RequestURI
		//非health check才生成traceId
		if HealthUrl == url {
			c.Next()
			return
		} else {
			uuidStr := util.GenerateUuidSample()
			c.Set(GoTraceId, uuidStr)
		}

		startTime := time.Now()
		method := c.Request.Method
		var reqBody string
		if http.MethodPost == method {

			reqBytes, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				Logger.Errorln("读取请求失败 %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": "500",
					"msg":  "读取请求body体失败",
				})
				c.Abort()
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBytes))
			reqBody = string(reqBytes)
		}
		blw := &BodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		c.Next() // 调用该请求的剩余处理程序
		stopTime := time.Since(startTime)
		spendTime := fmt.Sprintf("%dms", int(math.Ceil(float64(stopTime.Nanoseconds()/1000000))))
		statusCode := c.Writer.Status()
		resBody := blw.body.String()
		msg := fmt.Sprintf("path=%s SpendTime=%s method=%s status=%d reqBody=%s resBody=%s", url, spendTime,
			method, statusCode, reqBody, resBody)

		if len(c.Errors) > 0 {
			LogEntry.WithFields(logrus.Fields{
				GoTraceId: c.GetString(GoTraceId),
			}).Error(msg, c.Errors.ByType(gin.ErrorTypePrivate))
		}
		if statusCode >= 500 {
			LogEntry.WithFields(logrus.Fields{
				GoTraceId: c.GetString(GoTraceId),
			}).Error(msg)
		} else if statusCode >= 400 {
			LogEntry.WithFields(logrus.Fields{
				GoTraceId: c.GetString(GoTraceId),
			}).Warn(msg)
		} else {
			LogEntry.WithFields(logrus.Fields{
				GoTraceId: c.GetString(GoTraceId),
			}).Info(msg)
		}
	}
}

/**
重写request流 不然日志中读取了request body  controller里面就读不到了
*/
type BodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

//重写BodyLogWriter的Write方法
func (w BodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// 整一个类似Java日志输出格式
type JavaStyleLogFormatter struct {
}

func (jslf *JavaStyleLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	timeStr := time.Now().Format("2006-01-02 15:04:05")
	//时间 level msg 文件+行号
	msg := fmt.Sprintf("[%s] [%s] [%s] [%s] \n", timeStr, entry.Level.String(), entry.Message,
		entry.Data[GoTraceId])

	return []byte(msg), nil
}
