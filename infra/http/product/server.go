package product

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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ProductServer implementation
type ProductServer struct {
	App    string
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
			Prefix: config.App,
		}),
	})
	engine.Use(ginmiddleware.Handler("", mdlw))
	return engine
}

// NewProductServer is the factory of product server
func NewProductServer(config *conf.Config, engine *gin.Engine, router *Router) infra_http.Server {
	return &ProductServer{
		App:    config.App,
		Port:   config.HTTPPort,
		Engine: engine,
		Router: router,
	}
}

// RegisterRoutes method register all endpoints
func (s *ProductServer) RegisterRoutes() {
	apiGroup := s.Engine.Group("/api")
	{
		apiGroup.GET("/product/:id", s.Router.GetProduct)
		apiGroup.GET("/products", s.Router.ListProducts)
		apiGroup.POST("/product", s.Router.CreateProduct)
	}
}

// Run is a method for starting server
func (s *ProductServer) Run() error {
	s.RegisterRoutes()
	addr := ":" + s.Port
	s.svr = &http.Server{
		Addr:    addr,
		Handler: otelhttp.NewHandler(s.Engine, s.App+"_http"),
	}
	log.Infoln("http server listening on ", addr)
	err := s.svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GracefulStop the server
func (s *ProductServer) GracefulStop(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
