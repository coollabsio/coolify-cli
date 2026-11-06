package appflags

import (
	"encoding/base64"
	"unicode/utf8"
)

// EncodeContent returns value as base64, which the Coolify API requires for
// inline file content such as a Dockerfile. A value that is already valid
// base64 of UTF-8 text is returned unchanged, so callers who encode it
// themselves keep working. The check mirrors the server's own
// isBase64Encoded test: standard alphabet, padded, and it must round-trip.
func EncodeContent(value string) string {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && utf8.Valid(decoded) {
		return value
	}
	return base64.StdEncoding.EncodeToString([]byte(value))
}
