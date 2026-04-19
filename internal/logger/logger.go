package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var SugarLogger *zap.SugaredLogger

func RequestLogger(h http.Handler) http.Handler {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	SugarLogger = logger.Sugar()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			SugarLogger.Infow("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"request_id", middleware.GetReqID(r.Context()),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start),
			)
		}()

		h.ServeHTTP(ww, r)
	})
}
