package helpers

import (
    "encoding/json"
    "net"
    "net/http"
    "strings"

    "github.com/rs/xid"
    "booking_cinema_golang/internal/utils/constants"
)

// APIResponse chuẩn cho tất cả API
type APIResponse struct {
    Status  int         `json:"status"`
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"trace_id,omitempty"`
}

// WriteJSON ghi response JSON với trace ID
func WriteJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
    traceID := xid.New().String()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Trace-ID", traceID)
    w.WriteHeader(status)
    
    json.NewEncoder(w).Encode(APIResponse{
        Status:  status,
        Success: success,
        Message: message,
        Data:    data,
        TraceID: traceID,
    })
}

// WriteError ghi error response
func WriteError(w http.ResponseWriter, err error) {
    errMsg := err.Error()
    status := constants.ErrorStatusMap[errMsg]
    if status == 0 {
        status = http.StatusInternalServerError
        errMsg = constants.ErrInternalServer
    }
    
    WriteJSON(w, status, false, errMsg, nil)
}

// GetClientIP lấy IP thật của client
func GetClientIP(r *http.Request) string {
    // Check X-Forwarded-For
    if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
        ips := strings.Split(fwd, ",")
        for _, ip := range ips {
            ip = strings.TrimSpace(ip)
            if ip != "" && !isPrivateIP(ip) {
                return ip
            }
        }
    }

    // Check X-Real-IP
    if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
        if !isPrivateIP(realIP) {
            return realIP
        }
    }

    // Fallback to RemoteAddr
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        ip = r.RemoteAddr
    }

    // Handle IPv6 localhost
    if ip == "::1" || ip == "0:0:0:0:0:0:0:1" {
        return "127.0.0.1"
    }

    return ip
}

// isPrivateIP kiểm tra IP có phải private không
func isPrivateIP(ipStr string) bool {
    ip := net.ParseIP(ipStr)
    if ip == nil {
        return false
    }
    
    privateIPBlocks := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "169.254.0.0/16",
        "127.0.0.0/8",
    }
    
    for _, block := range privateIPBlocks {
        _, ipnet, _ := net.ParseCIDR(block)
        if ipnet.Contains(ip) {
            return true
        }
    }
    
    return false
}

// ParseJSON parse request body với giới hạn kích thước
func ParseJSON(r *http.Request, v interface{}) error {
    defer r.Body.Close()
    
    // Limit body size to 1MB
    limitedReader := http.MaxBytesReader(nil, r.Body, 1048576)
    
    return json.NewDecoder(limitedReader).Decode(v)
}