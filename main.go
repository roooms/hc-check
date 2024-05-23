package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	// URLs to fetch and their corresponding keys (without the base URL)
	urlsAndKeys := map[string][]string{
		"/v1/sys/health": {"version", "license.expiry_time", "replication_dr_mode"},
		"/v1/sys/storage/raft/autopilot/configuration": {"response.key"},
	}

	// Get the value of the environment variables
	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		fmt.Println("Error: VAULT_TOKEN environment variable not set")
		return
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		fmt.Println("Error: VAULT_ADDR environment variable not set")
		return
	}

	// Make HTTP requests and store responses
	for path, keys := range urlsAndKeys {
		url := vaultAddr + path
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Set the headers using the environment variables
		req.Header.Set("X-Vault-Token", vaultToken)
		req.Header.Set("X-Vault-Addr", vaultAddr)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			fmt.Printf("Error decoding response from %s: %v\n", url, err)
			return
		}

		// Extract the specified keys from the JSON response
		for _, key := range keys {
			value, ok := getValueByKey(data, key)
			if !ok {
				fmt.Printf("Response from %s (%s): key not found\n", path, key)
				continue
			}

			fmt.Printf("Response from %s (%s): %v\n", path, key, value)
		}
	}
}

// getValueByKey recursively retrieves the value for a given key from a nested map
func getValueByKey(data map[string]interface{}, key string) (interface{}, bool) {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		value, ok := data[keys[0]]
		return value, ok
	}

	firstKey := keys[0]
	remainingKeys := keys[1:]
	nestedData, ok := data[firstKey]
	if !ok {
		return nil, false
	}

	nestedMap, ok := nestedData.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return getValueByKey(nestedMap, strings.Join(remainingKeys, "."))
}
