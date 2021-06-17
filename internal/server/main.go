package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// Server is the API server.
type Server struct {
	kub backend.Backend
}

// New will instantiate a Server object.
func New(kub backend.Backend) *Server {
	return &Server{kub: kub}
}

// Run will initialize the http api server and configure all available
// routers.
func (s *Server) Run() error {
	if !klog.V(2) {
		gin.SetMode(gin.ReleaseMode)
	}

	router := s.getGinEngine()

	socket := viper.GetString("server.socket")
	if socket == "" {
		port := viper.GetString("server.listen-addr")
		klog.Infof("api server started listening on %s", port)
		if viper.GetBool("server.tls-enable") {
			cert := viper.GetString("server.tls-cert-file")
			key := viper.GetString("server.tls-key-file")
			router.RunTLS(port, cert, key)
		} else {
			router.Run(port)
		}
	} else {
		klog.Infof("api server started listening on %s", socket)
		router.RunUnix(socket)
	}

	return nil
}

// getGinEngine will return a gin.Engine router and configure the
// appropriate middleware.
func (s *Server) getGinEngine() *gin.Engine {
	router := gin.New()
	router.Use(httputil.VersionAliasMiddleware(router))
	router.Use(gin.Logger())
	router.Use(httputil.RequestLoggerMiddleware())
	router.Use(httputil.ResponseLoggerMiddleware())
	router.Use(gin.Recovery())

	insp := viper.GetBool("registry.inspector")
	if insp {
		klog.Infof("image inspector enabled")
	}

	pfwrd := viper.GetBool("port-forward")
	if pfwrd {
		klog.Infof("port-forwarding services to 127.0.0.1")
	}

	revprox := viper.GetBool("reverse-proxy")
	if revprox && !pfwrd {
		klog.Infof("enabled reverse-proxy services to localhost")
	}
	if revprox && pfwrd {
		klog.Infof("ignored reverse-proxy as port-forward is enabled")
		revprox = false
	}

	reqcpu := viper.GetString("kubernetes.request-cpu")
	if reqcpu != "" {
		klog.Infof("default cpu request: %s", reqcpu)
	}
	reqmem := viper.GetString("kubernetes.request-memory")
	if reqmem != "" {
		klog.Infof("default memory request: %s", reqmem)
	}

	routes.New(router, s.kub, routes.Config{
		Inspector:     insp,
		RequestCPU:    reqcpu,
		RequestMemory: reqmem,
		PortForward:   pfwrd,
		ReverseProxy:  revprox,
	})

	return router
}
