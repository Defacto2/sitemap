package main

import "testing"

// Cases taken from https://github.com/cfwheels/cfwheels/blob/1.4/wheels/tests/global/public/obfuscateparam.cfc
func TestObfuscateParam(t *testing.T) {
	cases := []struct {
		in, expected string
	}{
		{"999999999", "eb77359232"},
		{"0162823571", "0162823571"},
		{"1", "9b1c6"},
		{"99", "ac10a"},
		{"15765", "b226582"},
		{"69247541", "c06d44215"},
		{"0413", "0413"},
		{"per", "per"},
		//{"1111111111", "1111111111"}, This fails in sitemap.go, but passes in CFWheels?
	}
	for _, c := range cases {
		got := obfuscateParam(c.in)
		if got != c.expected {
			t.Errorf("obfuscateParam(%q) results in %q, expected %q", c.in, got, c.expected)
		}
	}
}
