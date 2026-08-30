package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeK8sVersion(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"v1.28.0", "1.28.0"},
		{"1.28.0", "1.28.0"},
		{"v1.37.0", "1.37.0"},
		{" v1.28.0 ", "1.28.0"},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalizeK8sVersion(c.in); got != c.want {
			t.Errorf("normalizeK8sVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// pendingNodeImage reproduces the image-tag construction in provisionPendingNode
// to guard against the double-v ("vv1.28.0") bug.
func pendingNodeImage(version string) string {
	version = normalizeK8sVersion(version)
	if version == "" {
		version = "1.28.0"
	}
	return "kindest/node:v" + version
}

func TestPendingNodeImageNoDoubleV(t *testing.T) {
	for _, v := range []string{"v1.28.0", "1.28.0"} {
		got := pendingNodeImage(v)
		want := "kindest/node:v1.28.0"
		if got != want {
			t.Errorf("pendingNodeImage(%q) = %q, want %q", v, got, want)
		}
		if strings.Contains(got, "vv") {
			t.Errorf("image %q contains double-v", got)
		}
	}
}

func TestExtractJSONObject(t *testing.T) {
	const clean = `{"serverVersion":{"gitVersion":"v1.28.0"},"clientVersion":{"gitVersion":"v1.36.0"}}`
	cases := map[string]string{
		"clean":                 clean,
		"warning prefix":        "WARNING: client/server version mismatch\n" + clean,
		"warning suffix":        clean + "\nWARNING: something went wrong",
		"warnings both sides":   "WARNING: a\n" + clean + "\nWARNING: b",
		"warning embeds braces": "WARNING: brace } here\n" + clean + "\nWARNING: another { here",
	}
	for name, in := range cases {
		got := extractJSONObject(in)
		var v struct {
			ServerVersion struct {
				GitVersion string `json:"gitVersion"`
			} `json:"serverVersion"`
		}
		if err := json.Unmarshal([]byte(got), &v); err != nil {
			t.Errorf("%s: extractJSONObject produced invalid JSON %q: %v", name, got, err)
			continue
		}
		if v.ServerVersion.GitVersion != "v1.28.0" {
			t.Errorf("%s: got gitVersion %q, want v1.28.0", name, v.ServerVersion.GitVersion)
		}
	}
}
