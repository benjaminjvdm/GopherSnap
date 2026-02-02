package converter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))
	re := regexp.MustCompile(`^([\d.]+)\s*(kb|mb|b)?$`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	var multiplier float64 = 1
	switch unit {
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	}

	return int64(value * multiplier), nil
}
