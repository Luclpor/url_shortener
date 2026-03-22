package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if !cw.wroteHeader {
		cw.wroteHeader = true
		headers := cw.Header()
		headers.Set("Content-Encoding", "gzip")
		headers.Del("Content-Length")
	}
	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *compressWriter) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}
	return cw.writer.Write(data)
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			rc, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer rc.Close()
			r.Body = rc
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
			if err != nil {
				http.Error(w, "failed to init gzip writer", http.StatusInternalServerError)
				return
			}
			defer gw.Close()

			cw := &compressWriter{
				ResponseWriter: w,
				writer:         gw,
			}
			w = cw
		}

		next.ServeHTTP(w, r)
	})
}
