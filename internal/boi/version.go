package boi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

var apkVersionRegexp = regexp.MustCompile(`mementomori_(\d+\.\d+\.\d+)\.apk`)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseAPKVersion(varsJS string) (string, error) {
	varsJS = normalizeScriptText(varsJS)
	match := apkVersionRegexp.FindStringSubmatch(varsJS)
	if len(match) != 2 {
		return "", fmt.Errorf("apk version not found")
	}
	return match[1], nil
}

func normalizeScriptText(s string) string {
	if strings.IndexByte(s, 0) == -1 {
		return s
	}
	b := []byte(s)
	if len(b) >= 2 {
		if decoded, ok := decodeUTF16LEBytes(b); ok {
			return decoded
		}
	}
	return strings.ReplaceAll(s, "\x00", "")
}

func decodeUTF16LEBytes(b []byte) (string, bool) {
	if len(b)%2 != 0 {
		return "", false
	}
	start := 0
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		start = 2
	}
	if (len(b)-start)%2 != 0 {
		return "", false
	}
	words := make([]uint16, 0, (len(b)-start)/2)
	for i := start; i < len(b); i += 2 {
		words = append(words, uint16(b[i])|uint16(b[i+1])<<8)
	}
	return string(utf16.Decode(words)), true
}

func ParseVersion(s string) (Version, error) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, err
	}
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func VersionCandidates(base string) ([]string, error) {
	v, err := ParseVersion(base)
	if err != nil {
		return nil, err
	}
	candidates := []string{v.String()}
	for i := 1; i <= 5; i++ {
		candidates = append(candidates, Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + i}.String())
	}
	for i := 1; i <= 5; i++ {
		candidates = append(candidates, Version{Major: v.Major, Minor: v.Minor + i, Patch: 0}.String())
	}
	for i := 1; i <= 5; i++ {
		candidates = append(candidates, Version{Major: v.Major + i, Minor: 0, Patch: 0}.String())
	}
	return candidates, nil
}
