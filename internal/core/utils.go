package core

import (
	"encoding/base64"
	"strings"
)

// toURLSafeBase64 converts a standard base64 string to URL-safe base64
// This is used to create dedup_keys from SHA1 hashes
func toURLSafeBase64(base64Hash string) string {
	// Replace + with -, / with _, and remove trailing =
	result := strings.ReplaceAll(base64Hash, "+", "-")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.TrimRight(result, "=")
	return result
}

// fromURLSafeBase64 converts a URL-safe base64 string to a standard base64 string
// This is used to convert dedup_keys back to SHA1 hashes
func fromURLSafeBase64(urlSafeBase64 string) string {
	result := strings.ReplaceAll(urlSafeBase64, "-", "+")
	result = strings.ReplaceAll(result, "_", "/")
	if pad := len(result) % 4; pad > 0 {
		result += strings.Repeat("=", 4-pad)
	}
	return result
}

// SHA1ToDedupeKey converts a SHA1 hash (raw bytes) to a dedup_key
func SHA1ToDedupeKey(sha1Hash []byte) string {
	return toURLSafeBase64(base64.StdEncoding.EncodeToString(sha1Hash))
}

// DedupeKeyToSHA1 converts a dedup_key back to SHA1 hash (raw bytes)
func DedupeKeyToSHA1(dedupKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(fromURLSafeBase64(dedupKey))
}
