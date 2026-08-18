package lottery

import "testing"

func TestParseCivilDate(t *testing.T) {
	got := ParseCivilDate("2026-08-17 21:25:00")
	if got.Format("2006-01-02") != "2026-08-17" {
		t.Fatalf("got %v", got)
	}
	got = ParseCivilDate("2026-08-17(一)")
	if got.Format("2006-01-02") != "2026-08-17" {
		t.Fatalf("kl8 date got %v", got)
	}
}

func TestDateParamDoesNotShiftUTCMidnight(t *testing.T) {
	d := ParseCivilDate("2026-08-17")
	if DateParam(d) != "2026-08-17" {
		t.Fatalf("param %s", DateParam(d))
	}
}
