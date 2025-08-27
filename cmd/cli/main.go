package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	baseURL      string
	accessToken  string
	refreshToken string
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "anybase",
		Short: "AnyBase CLI - Interact with AnyBase API",
		Long:  `A CLI tool to interact with AnyBase API for testing and management.`,
	}

	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "http://localhost:8080", "Base URL of the API")
	rootCmd.PersistentFlags().StringVar(&accessToken, "token", "", "Access token for authentication")

	// Auth commands
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication operations",
	}

	// Register command
	registerCmd := &cobra.Command{
		Use:   "register [email] [password]",
		Short: "Register a new user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")

			req := RegisterRequest{
				Email:     args[0],
				Password:  args[1],
				FirstName: firstName,
				LastName:  lastName,
			}

			resp, err := makeRequest("POST", "/api/v1/auth/register", req, "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			printJSON(resp)
		},
	}
	registerCmd.Flags().String("first-name", "", "First name")
	registerCmd.Flags().String("last-name", "", "Last name")

	// Login command
	loginCmd := &cobra.Command{
		Use:   "login [email] [password]",
		Short: "Login to get access token",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			req := LoginRequest{
				Email:    args[0],
				Password: args[1],
			}

			resp, err := makeRequest("POST", "/api/v1/auth/login", req, "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Extract tokens
			var authResp map[string]interface{}
			if err := json.Unmarshal(resp, &authResp); err == nil {
				if token, ok := authResp["access_token"].(string); ok {
					fmt.Printf("\n🔑 Access Token saved! Use it with --token flag or set ANYBASE_TOKEN env variable\n")
					fmt.Printf("export ANYBASE_TOKEN=%s\n\n", token)
					accessToken = token
				}
				if refresh, ok := authResp["refresh_token"].(string); ok {
					refreshToken = refresh
				}
			}

			printJSON(resp)
		},
	}

	authCmd.AddCommand(registerCmd, loginCmd)

	// User commands
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
	}

	// Get profile command
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Get user profile",
		Run: func(cmd *cobra.Command, args []string) {
			token := getToken()
			if token == "" {
				fmt.Println("Error: Authentication required. Please login first or provide --token")
				return
			}

			resp, err := makeRequest("GET", "/api/v1/users/profile", nil, token)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			printJSON(resp)
		},
	}

	userCmd.AddCommand(profileCmd)

	// Health check command
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Check API health",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := makeRequest("GET", "/health", nil, "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			printJSON(resp)
		},
	}

	// Test command - quick test sequence
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run a quick test sequence",
		Run: func(cmd *cobra.Command, args []string) {
			email := fmt.Sprintf("test%d@example.com", time.Now().Unix())
			password := "TestPass123!"

			fmt.Println("🧪 Running test sequence...")
			fmt.Printf("📧 Email: %s\n\n", email)

			// Register
			fmt.Println("1️⃣ Registering user...")
			regReq := RegisterRequest{
				Email:     email,
				Password:  password,
				FirstName: "Test",
				LastName:  "User",
			}
			resp, err := makeRequest("POST", "/api/v1/auth/register", regReq, "")
			if err != nil {
				fmt.Printf("❌ Registration failed: %v\n", err)
				return
			}
			fmt.Println("✅ Registration successful!")

			// Login
			fmt.Println("\n2️⃣ Logging in...")
			loginReq := LoginRequest{
				Email:    email,
				Password: password,
			}
			resp, err = makeRequest("POST", "/api/v1/auth/login", loginReq, "")
			if err != nil {
				fmt.Printf("❌ Login failed: %v\n", err)
				return
			}

			var authResp map[string]interface{}
			if err := json.Unmarshal(resp, &authResp); err != nil {
				fmt.Printf("❌ Failed to parse response: %v\n", err)
				return
			}

			token, _ := authResp["access_token"].(string)
			fmt.Println("✅ Login successful!")

			// Get profile
			fmt.Println("\n3️⃣ Getting profile...")
			resp, err = makeRequest("GET", "/api/v1/users/profile", nil, token)
			if err != nil {
				fmt.Printf("❌ Get profile failed: %v\n", err)
				return
			}
			fmt.Println("✅ Profile retrieved!")

			fmt.Println("\n🎉 All tests passed!")
		},
	}

	rootCmd.AddCommand(authCmd, userCmd, healthCmd, testCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func makeRequest(method, endpoint string, body interface{}, token string) ([]byte, error) {
	url := baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func getToken() string {
	if accessToken != "" {
		return accessToken
	}
	return os.Getenv("ANYBASE_TOKEN")
}

func printJSON(data []byte) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Println(string(data))
		return
	}

	formatted, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return
	}

	// Add colors for better readability (optional)
	output := string(formatted)
	output = strings.ReplaceAll(output, `"access_token"`, "\033[32m\"access_token\"\033[0m")
	output = strings.ReplaceAll(output, `"refresh_token"`, "\033[32m\"refresh_token\"\033[0m")
	output = strings.ReplaceAll(output, `"email"`, "\033[36m\"email\"\033[0m")
	
	fmt.Println(output)
}