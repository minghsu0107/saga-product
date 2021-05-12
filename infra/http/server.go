package http

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/http/middleware"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	ginmiddleware "github.com/slok/go-http-metrics/middleware/gin"
	"go.opencensus.io/plugin/ochttp"
)

// Server is the http wrapper
type Server struct {
	Port   string
	Engine *gin.Engine
	Router *Router
	svr    *http.Server
}

// NewEngine is a factory for gin engine instance
// Global Middlewares and api log configurations are registered here
func NewEngine(config *conf.Config) *gin.Engine {
	gin.SetMode(config.GinMode)
	if config.GinMode == "release" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	gin.DefaultWriter = io.Writer(config.Logger.Writer)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LogMiddleware(config.Logger.ContextLogger))
	engine.Use(middleware.CORSMiddleware())

	mdlw := prommiddleware.New(prommiddleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Prefix: config.AppName,
		}),
	})
	engine.Use(ginmiddleware.Handler("", mdlw))
	return engine
}

// NewServer is the factory for server instance
func NewServer(config *conf.Config, engine *gin.Engine, router *Router) *Server {
	return &Server{
		Port:   config.HTTPPort,
		Engine: engine,
		Router: router,
	}
}

// RegisterRoutes method register all endpoints
func (s *Server) RegisterRoutes() {
}

// Run is a method for starting server
func (s *Server) Run() error {
	s.RegisterRoutes()
	addr := ":" + s.Port
	s.svr = &http.Server{
		Addr: addr,
		// default propagation format: B3
		Handler: &ochttp.Handler{
			Handler: s.Engine,
			// IsHealthEndpoint holds the function to use for determining if the
			// incoming HTTP request should be considered a health check. This is in
			// addition to the private isHealthEndpoint func which may also indicate
			// tracing should be skipped.
			// IsHealthEndpoint: nil,
		},
	}
	log.Infoln("http server listening on ", addr)
	err := s.svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GracefulStop the server
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
