// +build ignore

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// XOR key - must match the key in internal/config/embedded.go
var xorKey = []byte{0x4B, 0x65, 0x79, 0x5F, 0x32, 0x30, 0x32, 0x35, 0x5F, 0x58, 0x4F, 0x52, 0x5F, 0x53, 0x65, 0x63}

func encryptConfig(data []byte) string {
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ xorKey[i%len(xorKey)]
	}
	return base64.StdEncoding.EncodeToString(encrypted)
}

func main() {
	// Read config.json
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config.json: %v\n", err)
		os.Exit(1)
	}

	// Validate JSON
	var testConfig map[string]interface{}
	if err := json.Unmarshal(data, &testConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid JSON in config.json: %v\n", err)
		os.Exit(1)
	}

	// Encrypt
	encrypted := encryptConfig(data)

	// Generate embedded.go file
	embeddedPath := filepath.Join("internal", "config", "embedded.go")

	content := fmt.Sprintf(`package config

import (
	"encoding/base64"
)

// Embedded encrypted configuration
// This file is auto-generated - do not edit manually
// Generated from config.json
const encryptedConfig = "%s"

// XOR key for obfuscation
var xorKey = []byte{0x4B, 0x65, 0x79, 0x5F, 0x32, 0x30, 0x32, 0x35, 0x5F, 0x58, 0x4F, 0x52, 0x5F, 0x53, 0x65, 0x63}

func decryptConfig() ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedConfig)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i++ {
		decrypted[i] = encrypted[i] ^ xorKey[i%%len(xorKey)]
	}

	return decrypted, nil
}
`, encrypted)

	if err := os.WriteFile(embeddedPath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing embedded.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Configuration encrypted and embedded successfully!")
	fmt.Printf("✓ Updated: %s\n", embeddedPath)
	fmt.Println("\nNow you can build your application with:")
	fmt.Println("  go build -ldflags=\"-s -w\" -o custos.exe ./cmd/custos")
}
