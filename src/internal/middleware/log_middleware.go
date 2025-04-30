package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"scti/config"
	"scti/internal/models"
	"scti/internal/utilities"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LogEntry struct {
	Timestamp   time.Time   `json:"timestamp"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Status      int         `json:"status"`
	IPAddress   string      `json:"ip_address"`
	UserAgent   string      `json:"user_agent"`
	Duration    int64       `json:"duration_ms"`
	UserID      string      `json:"user_id,omitempty"`
	UserEmail   string      `json:"user_email,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	IsVerified  bool        `json:"is_verified,omitempty"`
	IsMaster    bool        `json:"is_master,omitempty"`
	IsSuper     bool        `json:"is_super,omitempty"`
	AdminStatus interface{} `json:"admin_status,omitempty"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body
	return rw.ResponseWriter.Write(body)
}

func LoggingMiddleware(next http.Handler, logsDir string) http.Handler {
	if err := os.MkdirAll(filepath.Join(logsDir, "events"), 0755); err != nil {
		fmt.Printf("Error creating logs directory: %v\n", err)
	}

	var authMutex sync.Mutex
	var eventMutexes = make(map[string]*sync.Mutex)
	var eventMutexLock sync.Mutex

	eventRegex := regexp.MustCompile(`^/events/([^/]+)`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Get IP address, preferring X-Forwarded-For if available
		ipAddress := r.Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = r.RemoteAddr
		}

		var claims *models.UserClaims
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			secretKey := config.GetJWTSecret()
			token, err := jwt.ParseWithClaims(tokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if err == nil && token.Valid {
				if parsedClaims, ok := token.Claims.(*models.UserClaims); ok {
					claims = parsedClaims
				}
			}
		}

		// If token parsing failed, try the context as fallback
		if claims == nil {
			claims = utilities.GetUserFromContext(r.Context())
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(startTime).Milliseconds()

		logEntry := LogEntry{
			Timestamp: startTime,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    rw.statusCode,
			IPAddress: ipAddress,
			UserAgent: r.UserAgent(),
			Duration:  duration,
		}

		if claims != nil {
			logEntry.UserID = claims.ID
			logEntry.UserEmail = claims.Email
			logEntry.UserName = fmt.Sprintf("%s %s", claims.Name, claims.LastName)
			logEntry.IsVerified = claims.IsVerified
			logEntry.IsMaster = claims.IsMaster
			logEntry.IsSuper = claims.IsSuper

			if claims.AdminStatus != "" {
				var adminStatusMap map[string]string
				if err := json.Unmarshal([]byte(claims.AdminStatus), &adminStatusMap); err == nil {
					logEntry.AdminStatus = adminStatusMap
				} else {
					logEntry.AdminStatus = claims.AdminStatus
				}
			}
		}

		if strings.HasPrefix(r.URL.Path, "/register") ||
			strings.HasPrefix(r.URL.Path, "/login") ||
			strings.HasPrefix(r.URL.Path, "/logout") ||
			strings.HasPrefix(r.URL.Path, "/verify") ||
			strings.HasPrefix(r.URL.Path, "/forgot-password") ||
			strings.HasPrefix(r.URL.Path, "/change-password") ||
			strings.HasPrefix(r.URL.Path, "/refresh-tokens") ||
			strings.HasPrefix(r.URL.Path, "/revoke-refresh-token") {
			// Auth routes go to system_logs.json
			writeToSystemLog(logsDir, logEntry, &authMutex)
		} else if matches := eventRegex.FindStringSubmatch(r.URL.Path); len(matches) > 1 {
			// Event routes go to events/{slug}_logs.json
			slug := matches[1]
			writeToEventLog(logsDir, slug, logEntry, &eventMutexes, &eventMutexLock)
		} else {
			// Any other route goes to system_logs.json
			writeToSystemLog(logsDir, logEntry, &authMutex)
		}
	})
}

func writeToSystemLog(logsDir string, entry LogEntry, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	logFilePath := filepath.Join(logsDir, "system_logs.json")
	writeLogEntry(logFilePath, entry)
}

func writeToEventLog(logsDir string, slug string, entry LogEntry, mutexes *map[string]*sync.Mutex, mutexLock *sync.Mutex) {
	mutexLock.Lock()
	mutex, exists := (*mutexes)[slug]
	if !exists {
		mutex = &sync.Mutex{}
		(*mutexes)[slug] = mutex
	}
	mutexLock.Unlock()

	mutex.Lock()
	defer mutex.Unlock()

	logFilePath := filepath.Join(logsDir, "events", fmt.Sprintf("%s_logs.json", slug))
	writeLogEntry(logFilePath, entry)
}

func writeLogEntry(filePath string, entry LogEntry) {
	var entries []LogEntry

	if file, err := os.OpenFile(filePath, os.O_RDONLY, 0644); err == nil {
		defer file.Close()

		data, err := io.ReadAll(file)
		if err == nil && len(data) > 0 {
			err = json.Unmarshal(data, &entries)
			if err != nil {
				entries = []LogEntry{}
			}
		}
	}

	entries = append(entries, entry)

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling log entries: %v\n", err)
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Printf("Error writing to log file %s: %v\n", filePath, err)
	}
}

func WithLogging(handler http.Handler, logsDir string) http.Handler {
	return LoggingMiddleware(handler, logsDir)
}
