package rest

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	handler    *Handler
}

func NewServer(addr string, h *Handler) *Server {
	router := gin.Default()
	h.RegisterRoutes(router)

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		handler: h,
	}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

