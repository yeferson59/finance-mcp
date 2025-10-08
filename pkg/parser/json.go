package parser

import (
	"io"
	"sync"

	"github.com/bytedance/sonic"
)

// JSON represents a high-performance JSON parser optimized for financial data.
// It uses sonic's optimized configuration to provide better performance
// for API responses containing stock/market data.
type JSON struct {
	// config holds the sonic API configuration
	config sonic.API
	// mu protects concurrent access for thread safety
	mu sync.RWMutex
}

// NewJSON creates a new optimized JSON parser instance.
// The parser is configured specifically for financial data parsing with
// optimized settings for number handling and performance.
//
// Returns a thread-safe parser ready for concurrent use.
func NewJSON() *JSON {
	config := sonic.Config{
		UseNumber:        true,
		EscapeHTML:       false,
		CompactMarshaler: true,
		CopyString:       true,
		ValidateString:   true,
	}.Froze()

	return &JSON{
		config: config,
	}
}

// Parse parses JSON data from an io.Reader into the provided destination.
// This method maintains compatibility with the existing interface.
//
// Parameters:
//   - dst: Destination any to unmarshal JSON into
//   - src: io.Reader containing JSON data
//
// Returns error if parsing fails or if input is invalid.
func (j *JSON) Parse(dst any, src io.Reader) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	decoder := j.config.NewDecoder(src)
	return decoder.Decode(dst)
}

// ParseBytes parses JSON data directly from byte slice into the provided destination.
// This method is more efficient than Parse() as it avoids io.Reader overhead
// and is optimized for API responses that are typically received as []byte.
//
// Parameters:
//   - dst: Destination any to unmarshal JSON into
//   - data: JSON data as byte slice
//
// Returns error if parsing fails or if input is invalid.
//
// This method should be preferred over Parse() when working with []byte data
// as it provides better performance and lower memory allocation.
func (j *JSON) ParseBytes(dst any, data []byte) error {
	return j.config.Unmarshal(data, dst)
}

// ParseString parses JSON data from string into the provided destination.
// Convenient method for parsing JSON strings without byte conversion overhead.
//
// Parameters:
//   - dst: Destination any to unmarshal JSON into
//   - data: JSON data as string
//
// Returns error if parsing fails or if input is invalid.
func (j *JSON) ParseString(dst any, data string) error {
	return j.config.UnmarshalFromString(data, dst)
}

// MarshalBytes marshals the provided data into JSON byte slice.
// Uses the optimized sonic configuration for consistent, high-performance serialization.
//
// Parameters:
//   - src: Data to marshal into JSON
//
// Returns JSON byte slice and error if marshaling fails.
func (j *JSON) MarshalBytes(src any) ([]byte, error) {
	return j.config.Marshal(src)
}

// MarshalString marshals the provided data into JSON string.
// Convenient method that avoids byte-to-string conversion overhead.
//
// Parameters:
//   - src: Data to marshal into JSON
//
// Returns JSON string and error if marshaling fails.
func (j *JSON) MarshalString(src any) (string, error) {
	return j.config.MarshalToString(src)
}

// Config returns the underlying sonic configuration.
// This can be useful for debugging or advanced customization.
func (j *JSON) Config() sonic.API {
	return j.config
}

// Default provides a singleton instance of the optimized JSON parser.
// This is convenient for simple use cases where you don't need multiple
// parser instances with different configurations.
var Default = NewJSON()

// ParseBytes is a convenience function using the default parser instance.
func ParseBytes(dst any, data []byte) error {
	return Default.ParseBytes(dst, data)
}

// ParseString is a convenience function using the default parser instance.
func ParseString(dst any, data string) error {
	return Default.ParseString(dst, data)
}

// MarshalBytes is a convenience function using the default parser instance.
func MarshalBytes(src any) ([]byte, error) {
	return Default.MarshalBytes(src)
}

// MarshalString is a convenience function using the default parser instance.
func MarshalString(src any) (string, error) {
	return Default.MarshalString(src)
}
