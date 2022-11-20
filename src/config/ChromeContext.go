package config

import (
	"context"
	"github.com/chromedp/chromedp"
	"os"
)

var ChromeCtx context.Context

/**
chrome初始化 全局使用这一个实例即可
*/
func init() {

	var headlessFlag chromedp.ExecAllocatorOption
	//headless这个默认是true，如果想要在本地调试的时候看下浏览器的行为，可以在
	//环境变量里添加headless=false 就可以在本地调试并观察浏览器被控制的行为了
	isHeadless := os.Getenv("headless")
	if isHeadless == "false" {
		headlessFlag = chromedp.Flag("headless", false)
	} else {
		headlessFlag = chromedp.Flag("headless", true)
	}
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		//不检查默认浏览器
		chromedp.NoDefaultBrowserCheck,
		//无头
		headlessFlag,
		//忽略错误
		chromedp.IgnoreCertErrors,
		//不加载gif图像 因为有可能会卡住
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		//关闭GPU渲染
		chromedp.DisableGPU,
		//不适用谷歌的sanbox模式运行
		chromedp.NoSandbox,
		//设置网站不是首次运行
		chromedp.NoFirstRun,
		//禁用网络安全标志
		chromedp.Flag("disable-web-security", true),
		//关闭插件支持
		chromedp.Flag("disable-extensions", true),
		//关闭默认浏览器检查
		chromedp.Flag("disable-default-apps", true),
		//初始大小
		chromedp.WindowSize(1920, 1080),
		//在呈现所有数据之前防止创建Pdf
		chromedp.Flag("run-all-compositor-stages-before-draw", true),
		//设置userAgent 不然chrome会标识自己是个chrome爬虫 会被反爬虫网页拒绝
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`), //设置UserAgent
	)

	ChromeCtx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
}
