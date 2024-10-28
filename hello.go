package main

import (
    "flag"
    "net/http"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/hello-api.yaml", "配置文件路径")

func main() {
    flag.Parse()

    var c rest.RestConf
    conf.MustLoad(*configFile, &c)

    server := rest.MustNewServer(c)
    defer server.Stop()

    server.AddRoute(rest.Route{
        Method:  http.MethodGet,
        Path:    "/hello",
        Handler: func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("Hello World!"))
        },
    })

    server.Start()
}