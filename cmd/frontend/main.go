package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"net/url"
	"net/http/httputil"
)

func main() {
	// Serve static files from the "frontend" directory at project root.
	// frontendDir := filepath.Join(".", "frontend")

	// mux := http.NewServeMux()
	// mux.Handle("/", http.FileServer(http.Dir(frontendDir)))
	// log.Printf("Serving frontend from %s\n", frontendDir)
	// addr := ":5173"
	// log.Printf("Serving frontend from %s on http://localhost%s\n", frontendDir, addr)
	// if err := http.ListenAndServe(addr, mux); err != nil {
	// 	log.Fatal(err)
	// }
	// 1. Khai báo thư mục chứa file tĩnh
	frontendDir := filepath.Join(".", "frontend")
	mux := http.NewServeMux()

	// 2. Cấu hình Reverse Proxy trỏ về Backend API của bạn
	// THAY ĐỔI CỔNG NÀY THÀNH CỔNG BACKEND ĐANG CHẠY (ví dụ: 8080 hoặc 8081)
	backendURL, err := url.Parse("http://localhost:8081") 
	if err != nil {
		log.Fatal("Lỗi parse URL Backend:", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// 3. Phân luồng Request
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Nếu request bắt đầu bằng /api/, chuyển tiếp cho Backend xử lý
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("[PROXY] Forwarding API request: %s", r.URL.Path)
			proxy.ServeHTTP(w, r)
			return
		}

		// Nếu không, trả về các file giao diện HTML/CSS/JS bình thường
		http.FileServer(http.Dir(frontendDir)).ServeHTTP(w, r)
	})

	// 4. Chạy Server
	addr := ":5173"
	log.Printf("Frontend server is running on http://localhost%s", addr)
	log.Printf("Proxying API requests to %s", backendURL.String())
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

