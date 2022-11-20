package config

//traceId,日志链路打印使用，不想java一样可以获取线程ID来查看逻辑是什么执行的
//go 里面是goruntime 无法获取协程ID，因此需要使用这个id来打印日志
const GoTraceId = "goTraceId"

//healthcheck 路径 日志中不想打出health日志，判断使用
const HealthUrl = "/screenshot-go/health"
