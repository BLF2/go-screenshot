可以使用golang操作chrome完成一打印当前页面到PDF操作，支持滚动加载的页面
### 环境
+ go环境：go 1.18
+ chrome dev tool 驱动：chromedp v0.8.6
+ 浏览器：chrome 107.0.5304.110 (正式版本) (x86_64) 
+ 日志： logrus v1.9.0

### 说明
1. 本系统有两个环境变量
   1. headless：这个环境变量不加默认是`true`，代表了使用chrome的headless模式，如果想在本地调试，看下代码执行过程中浏览器行为是什么样子的，可以将这个参数设置为false，形如：`headless=true`
   2. logDir：这个参数是当前项目输出日志的文件存放地，linux,mac形如:`/home/xxx/screenshot/logs`,windows下形如：`D:/opt/screenshot/logs`，这个路径就是`D:\opt\screenshot\logs`
2. 系统使用了一个chrome实例多个Tab页的形式进行多个任务执行，每个任务一个Tab页