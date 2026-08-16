package glob

import "testing"

// TestOpenIssues verifies the reproduction cases from the GitHub issues
// reported against the pre-v1 versions. Each case is labeled with its issue
// number: https://github.com/gobwas/glob/issues/<n>.
func TestOpenIssues(t *testing.T) {
	for _, c := range []struct {
		issue    int
		pat, str string
		sep      []rune
		exp      bool
	}{
		{66, `{**/daxing}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing,daxing}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing,daxing,x}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing,daxing,x,y}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{54, "1?5ö", "155ö", nil, true},
		{54, "1ö?5", "1ö55", nil, true},
		{54, "1?ö5hello", "15ö5hello", nil, true},
		{54, "1?5helloö", "155helloö", nil, true},
		{51, "a*ant", "an ant", nil, true},
		{51, "a*ant", "ant", nil, false},
		{51, "br*r", "brother", nil, true},
		{51, "br*r", "br", nil, false},
		{51, "so*so", "so so", nil, true},
		{51, "so*so", "so", nil, false},
		{50, "yandex:*.exe:page.*", "yandex:service.exe:page.12345", nil, true},
		{50, "*yandex:*.exe:page.*", "yandex:service.exe:page.12345", nil, true},
		{50, "{*yandex:*.exe:page.*}", "yandex:service.exe:page.12345", nil, true},
		{50, "{*yandex:*.exe:page.*,x}", "yandex:service.exe:page.12345", nil, true},
		{43, "{,*.}google*", "google.com", nil, true},
		{43, "{,*.}google*", "a.google.com", nil, true},
		{43, "{,*.}google*", "agoogle.com", nil, false},
		{43, "{*.,}google*", "google.com", nil, true},
		{43, "{*.,}google*", "a.google.com", nil, true},
		{43, "{*.,}google*", "agoogle.com", nil, false},
		{41, "hell[o]", "hello", nil, true},
		{41, "ångstr[ö]m", "ångström", nil, true},
		{41, "hell?", "hello", nil, true},
		{41, "ångstr?m", "ångström", nil, true},
		{33, "test*201901*.12*", "test.20190101.120000", nil, true},
		{33, "test.*201901*.14*", "test.20190101.140000", nil, true},
		{33, "test.*201901*.14*", "test.20190101.130000", nil, false},
		{36, "test,pattern", "test,pattern", nil, true},
		{36, `test\,pattern`, "test,pattern", nil, true},
	} {
		got := MustCompile(c.pat, c.sep...).Match(c.str)
		if got != c.exp {
			t.Errorf("issue #%d: Match(%q, %q) = %t; want %t", c.issue, c.pat, c.str, got, c.exp)
		}
	}
	// Crashers: old versions panicked, new must return a syntax error.
	if _, err := Compile("/{a{*.json{", '/'); err == nil {
		t.Errorf("issue #59: Compile succeeded; want error")
	}
	if _, err := Compile("0{", '.', '/'); err == nil {
		t.Errorf("issue #56: Compile succeeded; want error")
	}
}
