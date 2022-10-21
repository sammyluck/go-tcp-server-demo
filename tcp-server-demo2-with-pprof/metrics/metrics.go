package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const metricsHTTPPort = 8889 //for prometheus to connect

var (
	ClientConnected prometheus.Gauge   // 当前已连接的客户端数量，对一个数值的即时测量值，反映一个值的瞬时快照
	ReqRecvTotal    prometheus.Counter // 每秒接收消息请求的数量
	RspSendTotal    prometheus.Counter // 每秒发送消息响应的数量
)

func init() {
	ReqRecvTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tcp_server_demo2_req_recv_total",
	})

	RspSendTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tcp_server_demo2_rsp_send_total",
	})

	ClientConnected = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tcp_server_demo2_client_connected",
	})

	prometheus.MustRegister(ReqRecvTotal, RspSendTotal, ClientConnected)

	// start the metrics server
	metricsServer := &http.Server{
		Addr: fmt.Sprintf(":%d", metricsHTTPPort),
	}

	mu := http.NewServeMux()
	mu.Handle("/metrics", promhttp.Handler())
	metricsServer.Handler = mu
	go func() {
		err := metricsServer.ListenAndServe()
		if err != nil {
			fmt.Println("prometheus-exporter http server start failed", err)
		}
	}()
	fmt.Printf("metrics server start ok(*:%d) \n", metricsHTTPPort)
}
