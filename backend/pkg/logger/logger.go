package logger

import (
    "context"
    "log"
    "net/http"
    "time"
)

type Logger struct{}

func New() *Logger { return &Logger{} }

func (l *Logger) Fatal(v ...interface{}) { log.Fatal(v...) }
func (l *Logger) Fatalf(format string, v ...interface{}) { log.Fatalf(format, v...) }
func (l *Logger) Infof(format string, v ...interface{}) { log.Printf(format, v...) }
func (l *Logger) Info(v ...interface{}) { log.Print(v...) }
func (l *Logger) Errorf(format string, v ...interface{}) { log.Printf(format, v...) }

func ChiMiddleware(l *Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            l.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
        })
    }
}

func (l *Logger) WithContext(ctx context.Context) *Logger { return l }
