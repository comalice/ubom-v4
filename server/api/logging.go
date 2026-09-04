package api

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

const requestIDHeader = "X-Request-ID"

type responseRecorder struct {
	http.ResponseWriter
	status          int
	bytes           int
	responseStarted bool
	log             *slog.Logger
	requestID       string
	method          string
	path            string
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.responseStarted {
		return
	}
	w.status = status
	w.responseStarted = true
	w.logResponseStarted()
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(value []byte) (int, error) {
	if !w.responseStarted {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(value)
	w.bytes += n
	return n, err
}

func (w *responseRecorder) Flush() {
	if !w.responseStarted {
		w.WriteHeader(http.StatusOK)
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseRecorder) logResponseStarted() {
	w.log.Info("http response started",
		slog.String("request_id", w.requestID),
		slog.String("method", w.method),
		slog.String("path", w.path),
		slog.Int("status", w.status),
	)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := requestID(r)
		w.Header().Set(requestIDHeader, requestID)
		started := time.Now()
		logger.Info("http request started",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		recorder := &responseRecorder{
			ResponseWriter: w,
			log:            logger,
			requestID:      requestID,
			method:         r.Method,
			path:           r.URL.Path,
		}
		defer func() {
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.Info("http request completed",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("bytes", recorder.bytes),
				slog.Float64("duration_ms", float64(time.Since(started).Microseconds())/1000),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}

func requestID(r *http.Request) string {
	if value := r.Header.Get(requestIDHeader); value != "" && len(value) <= 128 {
		return value
	}

	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return hex.EncodeToString([]byte(time.Now().String()))
}
