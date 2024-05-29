package main

import (
	"Prometheus-Go-Template/collector"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

/**
 * @Author: 南宫乘风
 * @Description:
 * @File:  main.go
 * @Email: 1794748404@qq.com
 * @Date: 2024-05-29 17:31
 */
var (
	// Set during go build
	// version   string
	// gitCommit string

	// 命令行参数
	listenAddr       = flag.String("web.listen-port", "8080", "An port to listen on for web interface and telemetry.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "app", "Prometheus metrics namespace, as the prefix of metrics name")
)

// query
func Query(w http.ResponseWriter, r *http.Request) {
	//模拟业务查询耗时0~1s
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	_, _ = io.WriteString(w, "some results")
}

// 主函数：启动一个HTTP服务器，提供Prometheus指标的导出和一个简单的HTML页面用于访问指标。
func main() {
	// 解析命令行参数
	flag.Parse()
	apiRequestCounter := collector.NewAPIRequestCounter(*metricsNamespace)
	memoryUsageGauge := collector.NewMemoryUsageGauge(*metricsNamespace)
	apiResponseMethod := collector.WebRequestTotal
	apiResponseTimeHistogram := collector.WebRequestDuration
	apiResponseCpuSummary := collector.NewCPULoadSummary(*metricsNamespace)
	// 创建一个新的Prometheus指标注册表
	registry := prometheus.NewRegistry()
	// 注册APIRequestCounter实例到Prometheus注册表
	registry.MustRegister(apiRequestCounter)
	// 注册MemoryUsageGauge实例到Prometheus注册表
	registry.MustRegister(memoryUsageGauge)
	// 注册APIResponseTimeHistogram实例到Prometheus注册表
	registry.MustRegister(apiResponseMethod)
	registry.MustRegister(apiResponseTimeHistogram)
	registry.MustRegister(apiResponseCpuSummary)

	// 设置HTTP服务器以处理Prometheus指标的HTTP请求
	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// 设置根路径的处理函数，用于返回一个简单的HTML页面，包含指向指标页面的链接
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
	            <head><title>A Prometheus Exporter</title></head>
	            <body>
	            <h1>A Prometheus Exporter</h1>
	            <p><a href='/metrics'>Metrics</a></p>
	            </body>
	            </html>`))
	})
	// 模拟API请求的处理函数
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		apiRequestCounter.IncrementRequestCount()
		// 模拟API处理时间
		w.Write([]byte("API请求处理成功"))
	})

	http.HandleFunc("/query", collector.Monitor(Query))
	// 模拟每秒记录一次CPU负载
	go func() {
		for {
			load := rand.Float64() * 100 // 模拟CPU负载
			apiResponseCpuSummary.RecordCPULoad(load)
			time.Sleep(time.Second)
		}
	}()

	// 记录启动日志并启动HTTP服务器监听
	log.Printf("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}
