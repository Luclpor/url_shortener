package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	writer       *gzip.Writer
	supportsGzip bool
	useGzip      bool
	wroteHeader  bool
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if cw.wroteHeader {
		return
	}
	cw.wroteHeader = true

	headers := cw.Header()

	if cw.supportsGzip {
		contentType := headers.Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			cw.useGzip = true
			headers.Set("Content-Encoding", "gzip")
			headers.Del("Content-Length")

			cw.writer, _ = gzip.NewWriterLevel(cw.ResponseWriter, gzip.BestSpeed)
		}
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

func (cw *compressWriter) Close() error {
	if cw.writer != nil {
		return cw.writer.Close()
	}
	return nil
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

		cw := &compressWriter{
			ResponseWriter: w,
			supportsGzip:   strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"),
		}
		defer cw.Close()

		next.ServeHTTP(cw, r)
	})
}
