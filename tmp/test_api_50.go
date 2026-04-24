package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8082/api/v1"

type Response struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

func main() {
	fmt.Println("=== STARTING API TEST (50 TEST CASES) ===")

	uniq := time.Now().Unix()
	userEmail := fmt.Sprintf("testuser%d@example.com", uniq)
	username := fmt.Sprintf("testuser%d", uniq)
	password := "Pass@123"
	var token string
	var movieID string
	var cinemaID string
	var roomID string
	var showtimeID string
	var bookingID string

	// TC-01: Register
	fmt.Print("TC-01: Register user... ")
	regBody := map[string]string{
		"username":  username,
		"email":     userEmail,
		"password":  password,
		"full_name": "Test User",
		"phone":     "0123456789",
	}
	res, code := post("/register", regBody, "")
	if code == 201 && res.Success {
		fmt.Println("PASS")
	} else {
		fmt.Printf("FAIL (Code: %d, Msg: %s)\n", code, res.Message)
	}

	// TC-02: Register Duplicate
	fmt.Print("TC-02: Register duplicate email... ")
	res, code = post("/register", regBody, "")
	if code == 409 {
		fmt.Println("PASS")
	} else {
		fmt.Printf("FAIL (Code: %d)\n", code)
	}

	// TC-06: Login
	fmt.Print("TC-06: Login... ")
	loginBody := map[string]string{
		"email":    userEmail,
		"password": password,
	}
	res, code = post("/login", loginBody, "")
	if code == 200 && res.Success {
		var loginData struct {
			Token string `json:"token"`
		}
		json.Unmarshal(res.Data, &loginData)
		token = loginData.Token
		fmt.Println("PASS")
	} else {
		fmt.Printf("FAIL (Code: %d)\n", code)
	}

	// TC-13: List Movies
	fmt.Print("TC-13: List movies... ")
	res, code = get("/movies", "")
	if code == 200 && res.Success {
		var movies []struct{ ID string `json:"id"` }
		json.Unmarshal(res.Data, &movies)
		if len(movies) > 0 {
			movieID = movies[0].ID
			fmt.Println("PASS")
		} else {
			fmt.Println("WARNING (No movies found, skipping dependent tests)")
		}
	} else {
		fmt.Printf("FAIL (Code: %d)\n", code)
	}

	// TC-24: List Cinemas
	fmt.Print("TC-24: List cinemas... ")
	res, code = get("/cinemas", "")
	if code == 200 && res.Success {
		var cinemas []struct{ ID string `json:"id"` }
		json.Unmarshal(res.Data, &cinemas)
		if len(cinemas) > 0 {
			cinemaID = cinemas[0].ID
			fmt.Println("PASS")
		} else {
			fmt.Println("WARNING (No cinemas found)")
		}
	} else {
		fmt.Printf("FAIL (Code: %d)\n", code)
	}

	// TC-25: List Rooms
	if cinemaID != "" {
		fmt.Print("TC-25: List rooms for cinema... ")
		res, code = get("/rooms?cinema_id="+cinemaID, "")
		if code == 200 && res.Success {
			var rooms []struct{ ID string `json:"id"` }
			json.Unmarshal(res.Data, &rooms)
			if len(rooms) > 0 {
				roomID = rooms[0].ID
				fmt.Println("PASS")
			} else {
				fmt.Println("WARNING (No rooms found)")
			}
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// TC-26: List seats by room
	if roomID != "" {
		fmt.Print("TC-26: List seats by room... ")
		res, code = get("/seats/room/"+roomID, "")
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// TC-35: List seats by showtime
	if showtimeID != "" {
		fmt.Print("TC-35: List seats by showtime... ")
		res, code = get("/seats/showtime/"+showtimeID, "")
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}


	// TC-15: Movie Detail
	if movieID != "" {
		fmt.Print("TC-15: Get movie detail... ")
		res, code = get("/movies/"+movieID, "")
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// TC-29: List Showtimes
	if movieID != "" {
		fmt.Print("TC-29: List showtimes... ")
		res, code = get("/showtimes?movie_id="+movieID, "")
		if code == 200 && res.Success {
			var showtimes []struct{ ID string `json:"id"` }
			json.Unmarshal(res.Data, &showtimes)
			if len(showtimes) > 0 {
				showtimeID = showtimes[0].ID
				fmt.Println("PASS")
			} else {
				fmt.Println("WARNING (No showtimes found for this movie)")
			}
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// TC-36: Create Booking
	if token != "" && showtimeID != "" {
		fmt.Print("TC-36: Create booking... ")
		bookBody := map[string]any{
			"showtime_id": showtimeID,
			"seats":       []string{"B1", "B2"},
		}
		res, code = post("/bookings", bookBody, token)
		if code == 201 && res.Success {
			var b struct{ ID string `json:"id"` }
			json.Unmarshal(res.Data, &b)
			bookingID = b.ID
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d, Msg: %s)\n", code, res.Message)
		}
	}

	// TC-45: My Bookings
	if token != "" {
		fmt.Print("TC-45: List my bookings... ")
		res, code = get("/bookings/user/me", token) // We used /user/{id} in main.go, but "me" logic might be missing or we use /bookings/my
		// Wait, I mapped /bookings/user/{id} to ListMyBookings which ignores {id} and uses claims.
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// TC-46: Cancel Booking
	if token != "" && bookingID != "" {
		fmt.Print("TC-46: Cancel booking... ")
		res, code = delete("/bookings/"+bookingID, token)
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	// ADMIN TESTS
	fmt.Println("\n--- ADMIN TESTS ---")
	adminEmail := "admin@gmail.com"
	adminPassword := "password123"
	var adminToken string

	fmt.Print("Admin Login... ")
	res, code = post("/login", map[string]string{"email": adminEmail, "password": adminPassword}, "")
	if code == 200 {
		var loginData struct {
			Token string `json:"token"`
		}
		json.Unmarshal(res.Data, &loginData)
		adminToken = loginData.Token
		fmt.Println("PASS")
	} else {
		fmt.Printf("FAIL (Code: %d, Msg: %s)\n", code, res.Error)
	}


	if adminToken != "" {
		// TC-48: List Users
		fmt.Print("TC-48: List users... ")
		res, code = get("/admin/users", adminToken)
		if code == 200 && res.Success {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}

		// TC-19: Admin Add Movie
		fmt.Print("TC-19: Admin add movie... ")
		newMovie := map[string]any{
			"title_vi":      "Phim Test Case",
			"duration_mins": 120,
			"status":        "now_showing",
			"rating_label":  "P",
		}
		res, code = post("/admin/movies", newMovie, adminToken)
		if code == 201 {
			fmt.Println("PASS")
		} else {
			fmt.Printf("FAIL (Code: %d)\n", code)
		}
	}

	fmt.Println("\n=== TEST COMPLETED ===")
}

func get(path, token string) (Response, int) {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, 0
	}
	defer resp.Body.Close()
	var r Response
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &r)
	return r, resp.StatusCode
}

func post(path string, data any, token string) (Response, int) {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, 0
	}
	defer resp.Body.Close()
	var r Response
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &r)
	return r, resp.StatusCode
}

func delete(path string, token string) (Response, int) {
	req, _ := http.NewRequest("DELETE", baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, 0
	}
	defer resp.Body.Close()
	var r Response
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &r)
	return r, resp.StatusCode
}
