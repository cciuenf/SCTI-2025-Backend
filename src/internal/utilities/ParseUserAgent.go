package utilities

import "strings"

func ParseUserAgent(userAgent string) string {
	deviceInfo := "Unknown"

	if strings.Contains(userAgent, "Windows") {
		deviceInfo = "Windows"
	} else if strings.Contains(userAgent, "Mac") {
		deviceInfo = "Mac"
	} else if strings.Contains(userAgent, "iPhone") {
		deviceInfo = "iPhone"
	} else if strings.Contains(userAgent, "Android") {
		deviceInfo = "Android"
	} else if strings.Contains(userAgent, "Linux") {
		deviceInfo = "Linux"
	} else {
		deviceInfo = "Other"
	}

	return deviceInfo
}
