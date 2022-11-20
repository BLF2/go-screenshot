package service

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"math"
	"runtime/debug"
	"screenshot-go/src/config"
	"strconv"
	"time"
)

/**
fileNameUrlMap: key为文件名 value为url
topDivId: 页面上最上面显示第一行字的div 用来定位页面的最高处 这里传入这个div的Id
waitVisibleExpr: 页面加载到这个选择器说明页面加载完成了
maxHighId: 获取页面最高的大小的div的id
chromeCtx: 谷歌浏览器实例
goTraceId: 日志中的traceId

返回值：
map的key为文件名，value是byte数组，返回生成的pdf的文件字节

注意：我这个是动态页面 是滚动加载的页面，需要模拟人手动去滚动滑块，然后等待加载数据，一直到底部为止
我当前的页面可以通过某个div获取到高度，如果你无法获取到高度，就往下滚动，监测本次滚动所能到达的高度和
上一次的是否一致，是的话就说明滚动到底了
*/
func ScreenPdf(fileNameUrlMap map[string]string, topDivId string, waitVisibleExpr string,
	maxHighId string, chromeCtx context.Context, goTraceId string) map[string][]byte {
	var (
		err       error
		resultMap = make(map[string][]byte)
	)
	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("开始进行截图... param=%s ", fileNameUrlMap)

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("初始化chrome完成...")

	for fileName, url := range fileNameUrlMap {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("开始截图，当前处理文件名=%s url=%s", fileName, url)
		chromeTabCtx, cancelFunc := chromedp.NewContext(chromeCtx, chromedp.WithLogf(config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof))
		//空任务触发初始化
		err = chromedp.Run(chromeTabCtx, make([]chromedp.Action, 0, 1)...)
		chromedp.Sleep(time.Second * 2)
		if err != nil {
			config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("初始化chrome并执行第一个Task失败跳过此截图 fileName=%s", fileName)
			continue
		}
		buf := make([]byte, 0)
		err = chromedp.Run(chromeTabCtx, chromedp.Tasks{
			chromedp.Navigate(url),
			chromedp.Sleep(time.Second * 10),
			chromedp.ActionFunc(func(ctx context.Context) error {
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("开始等待页面加载 检测点=%s fileName=%s", waitVisibleExpr, fileName)
				return nil
			}),
			chromedp.WaitVisible(waitVisibleExpr, chromedp.ByID),
			chromedp.ActionFunc(func(ctx context.Context) error {
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("页面加载完成 检测点=%s fileName=%s", waitVisibleExpr, fileName)
				var html string
				chromedp.InnerHTML(waitVisibleExpr, &html, chromedp.ByID)
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("获取到的页面html=%s", html)
				return nil
			}),
			chromedp.Sleep(time.Second * 15),
			chromedp.ActionFunc(func(ctx context.Context) error {
				//获取可视界面的高度
				var jsGetClientHigh = "document.body.clientHeight"
				clientHigh := getHighByJs(jsGetClientHigh, ctx)
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("可视高度为%d ", clientHigh)
				//获取最高的
				var jsGetMaxHigh = "document.getElementById('" + maxHighId + "').offsetHeight"
				maxHigh := getHighByJs(jsGetMaxHigh, ctx)
				config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("最大高度为%d ", maxHigh)
				var currentHigh = clientHigh
				//滚动
				for {
					if currentHigh < maxHigh {
						jsScroll := "document.getElementById('" + topDivId + "').scrollTop=" + strconv.Itoa(currentHigh)
						chromedp.EvalAsValue(&runtime.EvaluateParams{
							Expression:    jsScroll,
							ReturnByValue: false,
						}).Do(ctx)
						time.Sleep(time.Second * 15)
						currentHigh += clientHigh
					} else {
						config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof("跳出高度%d fileName=%s", currentHigh, fileName)
						break
					}
				}
				//滚动完成后滚回第一屏
				jsScroll0 := "document.getElementById('" + topDivId + "').scrollTop=0"
				chromedp.EvalAsValue(&runtime.EvaluateParams{
					Expression:    jsScroll0,
					ReturnByValue: false,
				}).Do(ctx)
				time.Sleep(time.Second * 1)
				//纸张设置为A0
				buf, _, err = page.PrintToPDF().WithPaperWidth(33.1).WithPaperHeight(46.8).WithPrintBackground(true).Do(ctx)
				return err
			}),
		})

		if err != nil {
			config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("截图出现报错 跳过当前PDF fileName=%s err=%v ", fileName, err)
			continue
		}
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).
			Infof("截图生成bytes完成 当前fileName=%s byteLength=%d", fileName, len(buf))
		resultMap[fileName] = buf
		cancelFunc()
	}

	return resultMap
}

func getHighByJs(jsGetHigh string, ctx context.Context) int {
	result, _, _ := chromedp.EvalAsValue(&runtime.EvaluateParams{
		Expression:    jsGetHigh,
		ReturnByValue: true,
	}).Do(ctx)
	json, _ := result.Value.MarshalJSON()
	clientHigh := bytesToInt(json)
	return clientHigh
}

func bytesToInt(bys []byte) int {
	length := float64(len(bys)) - 1
	var x float64
	for _, value := range bys {
		tmp := math.Pow(10, length)
		x = x + (float64(value)-48)*tmp
		length--
	}
	return int(x)

}

func TestStartChrome(chromeCtx context.Context, goTraceId string) {

	config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infoln("开始进行chrome测试...")

	chromeTabCtx, cancelFunc := chromedp.NewContext(chromeCtx, chromedp.WithLogf(config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infof))
	defer cancelFunc()

	err := chromedp.Run(chromeTabCtx, chromedp.Tasks{
		chromedp.Navigate("https://www.baidu.com"),
		chromedp.Sleep(time.Second * 3),
		chromedp.WaitNotPresent("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {

			config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Infoln("成功进入Chrome")
			return nil
		}),
	})

	if err != nil {
		config.LogEntry.WithFields(logrus.Fields{config.GoTraceId: goTraceId}).Errorf("启动Chrome报错 %v %s", err, string(debug.Stack()))
	}
}
