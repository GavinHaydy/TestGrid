package handler

import (
	"fmt"
	"gateway/internal/svc"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type RouteConfig struct {
	Path   string // 匹配路径前缀
	Target string // 目标服务地址
}

// RegisterRoutes 注册所有的路由（包括文档聚合和业务接口转发）
func RegisterRoutes(engine *rest.Server, ctx *svc.ServiceContext) {

	RegisterDocRoutes(engine, ctx)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, service := range ctx.Config.Services {

		target, err := url.Parse(service.Target)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		for _, prefix := range service.Prefix {

			paths := []string{
				"/" + prefix,
				"/" + prefix + "/:path",
				"/" + prefix + "/:path/:path",
				"/" + prefix + "/:path/:path/:path",
			}

			for _, method := range methods {

				var routes []rest.Route

				for _, p := range paths {
					routes = append(routes, rest.Route{
						Method:  method,
						Path:    p,
						Handler: proxyHandler(proxy),
					})
				}

				engine.AddRoutes(routes)
				logx.Info(engine.Routes())
			}
		}
	}
}

// RegisterDocRoutes 专门用来注册 Scalar 文档聚合页面的路由
// 修改点：同样增加 ctx *svc.ServiceContext 参数
func RegisterDocRoutes(engine *rest.Server, ctx *svc.ServiceContext) {

	engine.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/docs",
		Handler: func(w http.ResponseWriter, r *http.Request) {

			var sources strings.Builder

			for _, service := range ctx.Config.Services {

				sources.WriteString(fmt.Sprintf(`
{
	title: "%s",
	url: "/docs/%s.json",
},
`, service.Name, strings.ToLower(service.Name)))
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Gateway Docs</title>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</head>

<body>

<div id="app"></div>

<script>
Scalar.createApiReference('#app',{
	sources:[
%s
	]
})
</script>

</body>
</html>`, sources.String())
		},
	})

	for _, service := range ctx.Config.Services {

		service := service

		engine.AddRoute(rest.Route{
			Method: http.MethodGet,
			Path:   "/docs/" + strings.ToLower(service.Name) + ".json",
			Handler: proxyTo(
				fmt.Sprintf("%s/docs/%s.json",
					service.Target,
					strings.ToLower(service.Name),
				),
			),
		})
	}
}

// proxyHandler 处理业务接口的反向代理
func proxyHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

// proxyTo 处理 swagger.json 文档文件的反向代理
func proxyTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, resp.Body)
	}
}
