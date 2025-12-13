package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ResponseCompress struct {
	http.ResponseWriter
	header http.Header
	status int
	buf    *bytes.Buffer
}

func (l *ResponseCompress) WriteHeader(code int) {
	if l.status != 0 {
		return
	}
	l.status = code
}

func (l *ResponseCompress) Header() http.Header {
	return l.header
}

func (l *ResponseCompress) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	return l.buf.Write(b)
}

func NewResponseCompress(w http.ResponseWriter) *ResponseCompress {
	return &ResponseCompress{
		ResponseWriter: w,
		header:         make(http.Header),
		buf:            new(bytes.Buffer),
	}
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		newRequest, err := decompressRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writer := NewResponseCompress(w)

		next.ServeHTTP(writer, newRequest)

		acceptEncoding := r.Header.Get("Accept-Encoding")
		contentType := writer.header.Get("Content-Type")
		var needCompress bool
		if strings.Contains(acceptEncoding, "gzip") &&
			(contentType == "application/json" || contentType == "text/plain") {
			needCompress = true
		}

		for k, vv := range writer.header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		var body []byte
		if needCompress {
			w.Header().Set("Content-Encoding", "gzip")
			var err error
			body, err = compressResponse(writer)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			body = writer.buf.Bytes()
		}

		if writer.status == 0 {
			writer.status = http.StatusOK
		}

		w.WriteHeader(writer.status)

		if _, err := w.Write(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	})
}

func decompressRequest(r *http.Request) (*http.Request, error) {
	if r.Header.Get("Content-Encoding") == "gzip" &&
		(r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == "text/plain") {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("%s", "Ошибка декодирования gzip")
		}
		defer gz.Close()

		decompressedBody, err := io.ReadAll(gz)
		if err != nil {
			return nil, fmt.Errorf("%s", "Ошибка чтения разжатого тела")
		}
		newReq := r.Clone(r.Context())
		newReq.Body = io.NopCloser(bytes.NewBuffer(decompressedBody))
		newReq.ContentLength = int64(len(decompressedBody))
		newReq.Header.Del("Content-Encoding")
		return newReq, nil
	} else {
		return r, nil
	}
}

func compressResponse(resp *ResponseCompress) ([]byte, error) {
	var gzipBuf bytes.Buffer
	gz := gzip.NewWriter(&gzipBuf)
	_, err := gz.Write(resp.buf.Bytes())
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return gzipBuf.Bytes(), nil
}
