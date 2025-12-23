package layout

import (
	"fmt"
	"strings"
)

// MediaGroupFor maps a file extension to a canonical media group.
func MediaGroupFor(ext string) (string, error) {
	clean := strings.ToLower(strings.TrimSpace(ext))
	clean = strings.TrimPrefix(clean, ".")

	switch clean {
	case "d64", "d71", "d81", "g64":
		return "disks", nil
	case "tap", "t64":
		return "tape", nil
	case "crt":
		return "cart", nil
	default:
		return "", fmt.Errorf("unknown media extension: %s", ext)
	}
}
