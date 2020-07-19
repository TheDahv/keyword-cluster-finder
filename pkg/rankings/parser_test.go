package rankings

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("./test-data/6290/apartment-building-parking.json")
	if err != nil {
		t.Errorf("could not open test file: %v", err)
		t.FailNow()
	}
	defer f.Close()

	serp, err := Parse(f)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		t.FailNow()
	}

	if kw := serp.Keyword; kw != "apartment building parking" {
		t.Errorf("expected 'apartment building parking', got %s", kw)
	}
	if l := len(serp.Members); l != 20 {
		t.Errorf("expected 20 entries, got %d", l)
	}
}
