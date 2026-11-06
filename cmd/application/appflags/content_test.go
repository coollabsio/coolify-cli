package appflags

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeContent(t *testing.T) {
	raw := "FROM busybox:1.37\nRUN exit 42\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	assert.Equal(t, encoded, EncodeContent(raw), "raw text is encoded")
	assert.Equal(t, encoded, EncodeContent(encoded), "already-encoded text is not encoded twice")
	assert.Equal(t, "RlJPTSBidXN5Ym94OjEuMzcKUlVOIGV4aXQgNDIK", EncodeContent(raw))
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("FROM scratch")), EncodeContent("FROM scratch"), "single line, no newline")
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("FROM x\r\nCMD [\"a\"]")), EncodeContent("FROM x\r\nCMD [\"a\"]"), "CRLF and quotes")
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("FROM é\n")), EncodeContent("FROM é\n"), "non-ASCII text")

	// Base64 of bytes that are not UTF-8 is not a Dockerfile: encode it as text.
	notUTF8 := base64.StdEncoding.EncodeToString([]byte{0xff, 0xfe, 0xfd})
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(notUTF8)), EncodeContent(notUTF8))
}
