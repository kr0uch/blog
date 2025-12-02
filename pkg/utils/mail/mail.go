package mail

import (
	"net"
	"strings"
	"unicode/utf8"
)

func IsValidEmail(email string) bool {
	totalLength := utf8.RuneCountInString(email)
	if totalLength == 0 || totalLength > 254 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	localPart := parts[0]
	domainPart := parts[1]

	localLength := utf8.RuneCountInString(localPart)

	if localLength == 0 || localLength > 64 {
		return false
	}

	domainLength := utf8.RuneCountInString(domainPart)

	if domainLength == 0 || domainLength > 253 {
		return false
	}

	return hasMXRecord(domainPart)
}

func hasMXRecord(domain string) bool {
	records, err := net.LookupMX(domain)
	if err != nil || len(records) == 0 {
		return false
	}
	return true
}
