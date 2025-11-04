package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type GzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func GzipMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(res, req)
			return
		}
		if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
			dgz, err := gzip.NewReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			req.Body = struct {
				io.Reader
				io.Closer
			}{dgz, req.Body}
			defer dgz.Close()
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(res, gzip.BestSpeed)
		if err != nil {
			io.WriteString(res, err.Error())
			return
		}
		defer gz.Close()

		res.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(GzipResponseWriter{ResponseWriter: res, Writer: gz}, req)
	})
}
