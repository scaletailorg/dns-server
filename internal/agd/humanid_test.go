package agd_test

import (
	"strings"
	"testing"

	"github.com/AdguardTeam/AdGuardDNS/internal/agd"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHumanID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		in         string
		wantErrMsg string
	}{{
		name:       "success",
		in:         testHumanIDStr,
		wantErrMsg: "",
	}, {
		name:       "too_long",
		in:         testLongStr,
		wantErrMsg: `bad human id "` + testLongStr + `": too long: got 200 bytes, max 63`,
	}, {
		name:       "too_small",
		in:         "",
		wantErrMsg: `bad human id "": too short: got 0 bytes, min 1`,
	}, {
		name: "bad_start",
		in:   "-My-Device-X--10",
		wantErrMsg: `bad human id "-My-Device-X--10": bad hostname label "-My-Device-X--10": ` +
			`bad hostname label rune '-'`,
	}, {
		name: "bad_middle",
		in:   "My-Device-X---10",
		wantErrMsg: `bad human id "My-Device-X---10": at index 11: ` +
			`max 2 consecutive hyphens are allowed`,
	}, {
		name: "bad_rune",
		in:   "My-Device-X--10!",
		wantErrMsg: `bad human id "My-Device-X--10!": bad hostname label "My-Device-X--10!": ` +
			`bad hostname label rune '!'`,
	}, {
		name: "bad_end",
		in:   "My-Device-X--10-",
		wantErrMsg: `bad human id "My-Device-X--10-": bad hostname label "My-Device-X--10-": ` +
			`bad hostname label rune '-'`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id, err := agd.NewHumanID(tc.in)
			testutil.AssertErrorMsg(t, tc.wantErrMsg, err)

			if tc.wantErrMsg == "" {
				assert.Equal(t, tc.in, string(id))
			}
		})
	}
}

func TestNewHumanIDLower(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		in         string
		wantErrMsg string
	}{{
		name:       "success",
		in:         testHumanIDLowerStr,
		wantErrMsg: "",
	}, {
		name:       "bad_case",
		in:         "my-device-X--10",
		wantErrMsg: `bad lowercase human id "my-device-X--10": at index 10: 'X' is not lowercase`,
	}, {
		name: "bad_rune",
		in:   "My-Device-X--10!",
		wantErrMsg: `bad lowercase human id "My-Device-X--10!": ` +
			`bad hostname label "My-Device-X--10!": bad hostname label rune '!'`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id, err := agd.NewHumanIDLower(tc.in)
			testutil.AssertErrorMsg(t, tc.wantErrMsg, err)

			if tc.wantErrMsg == "" {
				assert.Equal(t, tc.in, string(id))
			}
		})
	}
}

func TestHumanIDParser_ParseNormalized(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		in         string
		wantErrMsg string
		wantID     agd.HumanID
	}{{
		name:       "success",
		in:         testHumanIDStr,
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "too_long",
		in:         testLongStr,
		wantErrMsg: "",
		wantID:     agd.HumanID(strings.Repeat("a", agd.MaxHumanIDLen)),
	}, {
		name:       "too_small",
		in:         "",
		wantErrMsg: `bad non-normalized human id "": too short: got 0 bytes, min 1`,
	}, {
		name:       "bad_start",
		in:         "-My-Device-X--10",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "bad_middle",
		in:         "My-Device-X---10",
		wantErrMsg: "",
		wantID:     "My-Device-X-10",
	}, {
		name:       "bad_rune",
		in:         "My-Device-X--10!",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "bad_end",
		in:         "My-Device-X--10-",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "bad_chars_start",
		in:         "абв-My-Device-X--10",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "bad_chars_end",
		in:         "My-Device-X--10-абв",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "bad_chars_middle",
		in:         "My-Device-Xабв10",
		wantErrMsg: "",
		wantID:     "My-Device-X-10",
	}, {
		name:       "bad_chars_middle_hyphens",
		in:         "My-Device-X-абв-10",
		wantErrMsg: "",
		wantID:     "My-Device-X-10",
	}, {
		name:       "bad_chars_middle_two_hyphens",
		in:         "My-Device-X--абв--10",
		wantErrMsg: "",
		wantID:     "My-Device-X-10",
	}, {
		name:       "bad_chars_middle_two_hyphens_other",
		in:         "My-DeviceабвX--10",
		wantErrMsg: "",
		wantID:     testHumanID,
	}, {
		name:       "one_bad_char",
		in:         "!",
		wantErrMsg: `bad non-normalized human id "!": cannot normalize`,
		wantID:     "",
	}, {
		name:       "only_bad_chars",
		in:         "!!!",
		wantErrMsg: `bad non-normalized human id "!!!": cannot normalize`,
		wantID:     "",
	}}

	p := agd.NewHumanIDParser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id, err := p.ParseNormalized(tc.in)
			assert.Equalf(t, tc.wantID, id, "original: %q", tc.in)
			testutil.AssertErrorMsg(t, tc.wantErrMsg, err)
		})
	}
}

// Sinks for benchmarks.
var (
	humanIDSink agd.HumanID
	errSink     error
)

func BenchmarkHumanIDParser_ParseNormalized(b *testing.B) {
	benchCases := []struct {
		name           string
		in             string
		wantErrPresent bool
	}{{
		name:           "valid",
		in:             testHumanIDStr,
		wantErrPresent: false,
	}, {
		name:           "normalized",
		in:             testHumanIDStr + "-!!!",
		wantErrPresent: false,
	}, {
		name:           "normalized_long",
		in:             testLongStr,
		wantErrPresent: false,
	}, {
		name:           "bad",
		in:             "!!!",
		wantErrPresent: true,
	}}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			p := agd.NewHumanIDParser()

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				humanIDSink, errSink = p.ParseNormalized(bc.in)
			}

			if bc.wantErrPresent {
				require.Empty(b, humanIDSink)
				require.Error(b, errSink)
			} else {
				require.NotEmpty(b, humanIDSink)
				require.NoError(b, errSink)
			}
		})
	}

	// Most recent result, on a ThinkPad X13 with a Ryzen Pro 7 CPU:
	//	goos: linux
	//	goarch: amd64
	//	pkg: github.com/AdguardTeam/AdGuardDNS/internal/agd
	//	cpu: AMD Ryzen 7 PRO 4750U with Radeon Graphics
	//	BenchmarkHumanIDParser_ParseNormalized/valid-16         	23716176	        51.48 ns/op	       0 B/op	       0 allocs/op
	//	BenchmarkHumanIDParser_ParseNormalized/normalized-16    	 1000000	      1071 ns/op	      88 B/op	       3 allocs/op
	//	BenchmarkHumanIDParser_ParseNormalized/normalized_long-16         	  504042	      4086 ns/op	     128 B/op	       4 allocs/op
	//	BenchmarkHumanIDParser_ParseNormalized/bad-16                     	  752247	      1361 ns/op	     184 B/op	       5 allocs/op
}

func FuzzHumanIDParser_ParseNormalized(f *testing.F) {
	p := agd.NewHumanIDParser()

	f.Fuzz(func(t *testing.T, input string) {
		s, err := p.ParseNormalized(input)
		if err != nil {
			require.Empty(t, s)

			return
		}

		assert.NotEmpty(t, s)
		assert.LessOrEqual(t, len(s), agd.MaxHumanIDLen)
	})
}
