package testfixture

import (
	"strings"
	"testing"
)

func TestFixtureRejectsNormalAndAmbiguousTargets(t *testing.T) {
	run := strings.Repeat("a", 32)
	base := "postgres://purchasing_fixture:" + strings.Repeat("b", 64) + "@127.0.0.1:15432/purchasing_test_" + run + "?sslmode=disable"
	if _, err := targetConfig(base, run); err != nil {
		t.Fatal("explicit synthetic target rejected")
	}
	for _, raw := range []string{"", strings.Replace(base, ":15432/", ":5432/", 1), strings.Replace(base, ":15432/", ":5433/", 1), strings.Replace(base, "127.0.0.1", "postgres", 1), strings.Replace(base, "purchasing_test_"+run, "purchasing_db", 1), strings.Replace(base, ":"+strings.Repeat("b", 64), "", 1), base + "&host=other", base + "&service=normal", base + "&sslmode=prefer"} {
		if _, err := targetConfig(raw, run); err == nil {
			t.Fatal("unowned or ambiguous target accepted")
		}
	}
}
