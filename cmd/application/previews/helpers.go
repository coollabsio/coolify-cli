package previews

import (
	"fmt"
	"strconv"
)

// minimumVersion is the first Coolify version with the preview list, get, and create endpoints.
const minimumVersion = "4.4.0"

func parsePullRequestID(value string) (int, error) {
	prID, err := strconv.Atoi(value)
	if err != nil || prID <= 0 {
		return 0, fmt.Errorf("invalid pr_id: must be a positive integer")
	}
	return prID, nil
}
