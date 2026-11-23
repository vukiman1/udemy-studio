package main

import "testing"

func TestSanitizeRedactsSecretsAndPortal(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "cloudfront signed url",
			input: `"https://vtt-c.udemycdn.com/a.vtt?Expires=1&Signature=abc~d&Key-Pair-Id=K1"`,
			want:  `"https://vtt-c.udemycdn.com/a.vtt?Expires=REDACTED&Signature=REDACTED&Key-Pair-Id=REDACTED"`,
		},
		{
			name:  "hls token keeps unrelated params",
			input: `"https://real-portal.udemy.com/x.m3u8?provider=cf&token=abc&v=1"`,
			want:  `"https://example-portal.udemy.com/x.m3u8?provider=cf&token=REDACTED&v=1"`,
		},
		{
			name:  "html escaped ampersand",
			input: `href=\"https://x.test/?a=1&amp;t=secret\"`,
			want:  `href=\"https://x.test/?a=1&amp;t=REDACTED\"`,
		},
		{
			name:  "bare json web token",
			input: `"eyJhbGciOi.eyJzdWIiOi.c2lnbmF0dXJl"`,
			want:  `"REDACTED"`,
		},
		{
			name:  "media license token",
			input: `"media_license_token": "opaque-value"`,
			want:  `"media_license_token": "REDACTED"`,
		},
		{
			name:  "plain text untouched",
			input: `"title": "HTTP/2 vs HTTP/3: what's next?"`,
			want:  `"title": "HTTP/2 vs HTTP/3: what's next?"`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := sanitize(testCase.input, "real-portal"); got != testCase.want {
				t.Errorf("sanitize(%s)\n got  %s\n want %s", testCase.input, got, testCase.want)
			}
		})
	}
}
