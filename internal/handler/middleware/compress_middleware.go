package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	useGzip     bool
	wroteHeader bool
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if cw.wroteHeader {
		return
	}
	cw.wroteHeader = true

	contentType := cw.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		cw.useGzip = true
		cw.Header().Set("Content-Encoding", "gzip")
		cw.Header().Del("Content-Length")
	}

	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *compressWriter) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}

	if cw.useGzip {
		return cw.writer.Write(data)
	}

	return cw.ResponseWriter.Write(data)
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			rc, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip request body", http.StatusBadRequest)
				return
			}
			defer rc.Close()
			r.Body = rc
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

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

		next.ServeHTTP(cw, r)
	})
}
