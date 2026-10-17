package redact

import "strings"

// credentialKeyFragments are substrings that mark a JSON key as holding a
// credential. They exist for untyped data (map[string]any responses, --json
// request bodies, opaque `any` fields) where there is no struct tag to read.
var credentialKeyFragments = []string{
	"token",
	"secret",
	"password",
	"passwd",
	"api_key",
	"license_key",
	"private_key",
	"webhook_url",
}

// nonCredentialKeySuffixes mark keys that name or point at a credential
// without containing one (e.g. integration_token_uuid, actor_token_name).
var nonCredentialKeySuffixes = []string{"_uuid", "_id", "_ids", "_name", "_at", "_enabled"}

// IsCredentialKey reports whether an untyped JSON key looks like it holds a
// credential, so its string value should be masked.
//
// Typed output is covered by `sensitive:"true"` tags instead; this is the
// fallback for shapes with no tags, and it errs towards masking. Callers
// should only mask non-empty string values so booleans and unset secrets keep
// their meaning.
func IsCredentialKey(key string) bool {
	key = strings.ToLower(key)
	for _, suffix := range nonCredentialKeySuffixes {
		if strings.HasSuffix(key, suffix) {
			return false
		}
	}
	if strings.HasSuffix(key, "_key") {
		return true
	}
	for _, fragment := range credentialKeyFragments {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}
