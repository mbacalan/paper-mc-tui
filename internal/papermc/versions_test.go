package papermc

import "testing"

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.21.10", "1.21.9", 1}, // numeric, not lexical
		{"1.21.9", "1.21.10", -1},
		{"1.20.6", "1.21.1", -1},
		{"26.1.2", "1.21.11", 1}, // CalVer outranks legacy
		{"26.2", "26.1.2", 1},
		{"26.1.2", "26.1.1", 1},
		{"1.21.11", "1.21.11-rc3", 1}, // release outranks pre-release
		{"1.21.11-rc3", "1.21.11", -1},
		{"1.21.9-pre4", "1.21.9-pre3", 1},
		{"1.21", "1.21.0", 0},
		{"1.21.11", "1.21.11", 0},
	}
	for _, tc := range cases {
		if got := compareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSortedVersionsNewestFirst(t *testing.T) {
	grouped := map[string][]string{
		"1.20": {"1.20.6"},
		"26.1": {"26.1.2", "26.1.1"},
		"1.21": {"1.21.11", "1.21.10"},
		"26.2": {"26.2-rc-2"},
	}
	got := sortedVersions(grouped)
	want := []string{"26.2-rc-2", "26.1.2", "26.1.1", "1.21.11", "1.21.10", "1.20.6"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortedVersions = %v, want %v", got, want)
		}
	}
}

func TestIsPrerelease(t *testing.T) {
	for _, v := range []string{"26.2-rc-2", "1.21.11-pre5", "1.21.9-rc1"} {
		if !isPrerelease(v) {
			t.Errorf("isPrerelease(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"26.1.2", "1.21.11", "1.20.6"} {
		if isPrerelease(v) {
			t.Errorf("isPrerelease(%q) = true, want false", v)
		}
	}
}
