package order

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	conf "github.com/minghsu0107/saga-product/config"
	infra_http "github.com/minghsu0107/saga-product/infra/http"
	"github.com/minghsu0107/saga-product/infra/http/middleware"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	ginmiddleware "github.com/slok/go-http-metrics/middleware/gin"
	"go.opencensus.io/plugin/ochttp"
)

// OrderServer implementation
type OrderServer struct {
	Port           string
	Engine         *gin.Engine
	Router         *Router
	svr            *http.Server
	jwtAuthChecker *middleware.JWTAuthChecker
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
			Prefix: config.App,
		}),
	})
	engine.Use(ginmiddleware.Handler("", mdlw))
	return engine
}

// NewOrderServer is the factory of order server
func NewOrderServer(config *conf.Config, engine *gin.Engine, router *Router, jwtAuthChecker *middleware.JWTAuthChecker) infra_http.Server {
	return &OrderServer{
		Port:           config.HTTPPort,
		Engine:         engine,
		Router:         router,
		jwtAuthChecker: jwtAuthChecker,
	}
}

// RegisterRoutes method register all endpoints
func (s *OrderServer) RegisterRoutes() {
	orderGroup := s.Engine.Group("/api/order")
	orderGroup.Use(s.jwtAuthChecker.JWTAuth())
	{
		orderGroup.GET("/:id", s.Router.GetDetailedOrder)
	}
}

// Run is a method for starting server
func (s *OrderServer) Run() error {
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
func (s *OrderServer) GracefulStop(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
