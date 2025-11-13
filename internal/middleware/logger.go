package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: 200}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		log.Printf(
			"[%s] %s %s %d %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.status,
			time.Since(start),
			r.UserAgent(),
		)
	})
}

func SecurityLogger(event string, r *http.Request, details string) {
	log.Printf(
		"[SECURITY] %s | IP: %s | URI: %s | User-Agent: %s | Details: %s",
		event,
		r.RemoteAddr,
		r.RequestURI,
		r.UserAgent(),
		details,
	)
}

func LogFailedLogin(r *http.Request, email string) {
	SecurityLogger("FAILED_LOGIN", r, "Email: "+email)
}

func LogSuccessfulLogin(r *http.Request, email string) {
	log.Printf(
		"[LOGIN] Successful login | Email: %s | IP: %s",
		email,
		r.RemoteAddr,
	)
}

func LogRegistration(r *http.Request, email string) {
	log.Printf(
		"[REGISTRATION] New user registered | Email: %s | IP: %s",
		email,
		r.RemoteAddr,
	)
}
