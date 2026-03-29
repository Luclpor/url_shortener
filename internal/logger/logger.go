package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func RequestLogger(h http.Handler) http.Handler {
	Logger, _ = zap.NewProduction()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		h.ServeHTTP(ww, r)
		Logger.Info(fmt.Sprintf("%s %s %s", r.Method, r.URL, time.Since(start)))
		defer func() {
			Logger.Info(fmt.Sprintf("request completed response status %d, bytes size %d", ww.Status(), ww.BytesWritten()))
		}()
	})
}
