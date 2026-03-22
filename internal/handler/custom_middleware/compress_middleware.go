package custom_middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (cw *compressWriter) Write(bt []byte) (int, error) {
	return cw.writer.Write(bt)
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gW, err := gzip.NewWriterLevel(w, gzip.BestCompression)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			wc := &compressWriter{
				ResponseWriter: w,
				writer:         gW,
			}
			defer gW.Close()
			w = wc
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			rc, err := gzip.NewReader(r.Body)
			r.Body = rc
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer rc.Close()
		}
		next.ServeHTTP(w, r)
	})
}
