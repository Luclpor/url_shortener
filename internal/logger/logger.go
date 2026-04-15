package logger

import (
	"bytes"
	"fmt"
	"io"
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
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}

			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		Logger.Info("incoming request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("body", string(bodyBytes)),
		)

		h.ServeHTTP(ww, r)
		Logger.Info(fmt.Sprintf("%s %s %s", r.Method, r.URL, time.Since(start)))
		defer func() {
			Logger.Info(fmt.Sprintf("request completed response status %d, bytes size %d", ww.Status(), ww.BytesWritten()))
		}()
	})
}
