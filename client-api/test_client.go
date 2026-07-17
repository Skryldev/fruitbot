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
)

// ============================================================
// Configuration
// ============================================================

const (
	BaseURL    = "http://localhost:8080"
	RestoreKey = "Enter-Your-Restore-Key"
)

// ============================================================
// API Response Types
// ============================================================

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Retry   bool   `json:"retry"`
}

// ============================================================
// HTTP Client
// ============================================================

var client = &http.Client{
	Timeout: 30 * time.Second,
}

// ============================================================
// Helper Functions
// ============================================================

func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func doRequest(method, path string, body interface{}) (*APIResponse, error) {
	url := BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}

func get(path string) (*APIResponse, error) {
	return doRequest("GET", path, nil)
}

func post(path string, body interface{}) (*APIResponse, error) {
	return doRequest("POST", path, body)
}

func check(name string, resp *APIResponse, err error) bool {
	fmt.Printf("\n🧪 %s\n", name)
	fmt.Println(strings.Repeat("-", 60))

	if err != nil {
		fmt.Printf("  ❌ FAIL: %v\n", err)
		return false
	}

	if resp.Success {
		fmt.Printf("  ✅ PASS (status: success)\n")
		if resp.Data != nil {
			fmt.Printf("  📦 Data: ")
			prettyPrint(resp.Data)
		}
		return true
	}

	if resp.Error != nil {
		fmt.Printf("  ⚠️  ERROR:\n")
		fmt.Printf("     Code:    %s\n", resp.Error.Code)
		fmt.Printf("     Message: %s\n", resp.Error.Message)
		if resp.Error.Details != "" {
			fmt.Printf("     Details: %s\n", resp.Error.Details)
		}
		fmt.Printf("     Retry:   %v\n", resp.Error.Retry)
		return false
	}

	fmt.Printf("  ❓ UNKNOWN RESPONSE\n")
	prettyPrint(resp)
	return false
}

// ============================================================
// Test Cases
// ============================================================

func testHealth() bool {
	resp, err := get("/health")
	return check("Health Check", resp, err)
}

func testStatus() bool {
	resp, err := get("/api/status")
	return check("Server Status", resp, err)
}

func testLoadPlayer() bool {
	resp, err := post("/api/player/load", map[string]interface{}{
		"restore_key":  RestoreKey,
		"save_session": true,
	})
	return check("Load Player", resp, err)
}

func testPlayerInfo() bool {
	resp, err := get("/api/player/info?restore_key=" + RestoreKey)
	return check("Player Info", resp, err)
}

func testCollectGold() bool {
	resp, err := post("/api/cards/collect-gold", map[string]interface{}{
		"restore_key": RestoreKey,
	})
	return check("Collect Gold", resp, err)
}

func testBuyCardPack() bool {
	resp, err := post("/api/store/buy-pack", map[string]interface{}{
		"restore_key": RestoreKey,
		"pack_type":   1,
	})
	return check("Buy Card Pack (Brown)", resp, err)
}

func testSendTribeMessage() bool {
	resp, err := post("/api/tribe/message", map[string]interface{}{
		"restore_key": RestoreKey,
		"text":        "Hello from Go test client! 👋",
	})
	return check("Send Tribe Message", resp, err)
}

func testListAccounts() bool {
	resp, err := get("/api/accounts")
	return check("List Accounts", resp, err)
}

func testAddAccount() bool {
	resp, err := post("/api/accounts", map[string]interface{}{
		"name":         "test_account",
		"restore_key":  RestoreKey,
		"mobile_model": "iPhone 15 Pro",
		"device_name":  "GoTestClient",
		"store_type":   "appstore",
	})
	return check("Add Account", resp, err)
}

func testLoadAllPlayers() bool {
	resp, err := post("/api/accounts/load-all", nil)
	return check("Load All Players", resp, err)
}

func testCollectGoldAll() bool {
	resp, err := post("/api/accounts/collect-gold-all", nil)
	return check("Collect Gold (All)", resp, err)
}

func testMessageAll() bool {
	resp, err := post("/api/accounts/message-all", map[string]interface{}{
		"text": "Broadcast from Go test client! 📢",
	})
	return check("Message All Tribes", resp, err)
}

func testMultiStats() bool {
	resp, err := get("/api/accounts/stats")
	return check("Multi Stats", resp, err)
}

func testRemoveAccount() bool {
	req, _ := http.NewRequest("DELETE", BaseURL+"/api/accounts/test_account", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return check("Remove Account", nil, err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	return check("Remove Account", &apiResp, nil)
}

func testErrorHandling() {
	fmt.Printf("\n🧪 %s\n", "Error Handling Tests")
	fmt.Println(strings.Repeat("-", 60))

	// Missing restore_key
	resp, err := post("/api/player/load", map[string]interface{}{
		"save_session": true,
	})
	if err == nil && !resp.Success {
		fmt.Printf("  ✅ Missing restore_key → correctly rejected (%s)\n", resp.Error.Code)
	} else {
		fmt.Printf("  ❌ Expected error for missing restore_key\n")
	}

	// Missing text for message
	resp, err = post("/api/tribe/message", map[string]interface{}{
		"restore_key": RestoreKey,
	})
	if err == nil && !resp.Success {
		fmt.Printf("  ✅ Missing text → correctly rejected (%s)\n", resp.Error.Code)
	} else {
		fmt.Printf("  ❌ Expected error for missing text\n")
	}

	// Invalid pack_type
	resp, err = post("/api/store/buy-pack", map[string]interface{}{
		"restore_key": RestoreKey,
		"pack_type":   999,
	})
	if err == nil && !resp.Success {
		fmt.Printf("  ✅ Invalid pack_type → correctly rejected (%s)\n", resp.Error.Code)
	} else {
		fmt.Printf("  ❌ Expected error for invalid pack_type\n")
	}
}

// ============================================================
// Main
// ============================================================

func main() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║        🍎 FruitBot API Client Test Suite 🍎        ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Server:      %s\n", BaseURL)
	fmt.Printf("  Restore Key: %s\n", RestoreKey)
	fmt.Println()

	startTime := time.Now()

	results := make(map[string]bool)

	// ============================================================
	// 1. Health & Info
	// ============================================================
	fmt.Println("━━━ 1. Health & Info ━━━")
	results["Health"] = testHealth()
	results["Status"] = testStatus()

	// ============================================================
	// 2. Player Operations
	// ============================================================
	fmt.Println("\n━━━ 2. Player Operations ━━━")
	results["LoadPlayer"] = testLoadPlayer()
	results["PlayerInfo"] = testPlayerInfo()

	// ============================================================
	// 3. Game Actions
	// ============================================================
	fmt.Println("\n━━━ 3. Game Actions ━━━")
	results["CollectGold"] = testCollectGold()
	results["BuyPack"] = testBuyCardPack()
	// results["SendMessage"] = testSendTribeMessage() // نیاز به tribe داره

	// ============================================================
	// 4. Multi-Account
	// ============================================================
	fmt.Println("\n━━━ 4. Multi-Account ━━━")
	results["ListAccounts"] = testListAccounts()
	results["AddAccount"] = testAddAccount()
	results["LoadAll"] = testLoadAllPlayers()
	results["CollectGoldAll"] = testCollectGoldAll()
	results["MultiStats"] = testMultiStats()
	results["RemoveAccount"] = testRemoveAccount()

	// ============================================================
	// 5. Error Handling
	// ============================================================
	fmt.Println("\n━━━ 5. Error Handling ━━━")
	testErrorHandling()

	// ============================================================
	// Summary
	// ============================================================
	elapsed := time.Since(startTime)
	
	passCount := 0
	failCount := 0
	for _, passed := range results {
		if passed {
			passCount++
		} else {
			failCount++
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Printf("║  ✅ Passed: %-2d  ❌ Failed: %-2d  ⏱️  Time: %-6s ║\n", 
		passCount, failCount, elapsed.Round(time.Millisecond))
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	if failCount > 0 {
		os.Exit(1)
	}
}