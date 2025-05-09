package web

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// registerSwagger 注册Swagger路由
func (s *Server) registerSwagger() {
	s.log.Info("Registering Swagger routes")

	// 注册Swagger UI路由
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)

	// 默认路径: /swagger/index.html
	s.router.GET("/swagger/*any", swaggerHandler)
}

// registerProfiling 注册性能分析路由
func (s *Server) registerProfiling() {
	s.log.Info("Registering profiling routes")

	// 创建一个单独的路由组
	profGroup := s.router.Group("/debug/pprof")
	{
		profGroup.GET("/", gin.WrapF(pprof.Index))
		profGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		profGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		profGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		profGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		profGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		profGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		profGroup.GET("/profile", gin.WrapF(pprof.Profile))
		profGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		profGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		profGroup.GET("/trace", gin.WrapF(pprof.Trace))
		profGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	}
}
