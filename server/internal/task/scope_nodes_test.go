package task

import (
	"reflect"
	"testing"
)

func TestParseScopeNodeIDs(t *testing.T) {
	ids, err := parseScopeNodeIDs("3,1,3, 2 ")
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{1, 2, 3}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("got %v want %v", ids, want)
	}
	if joinScopeNodeIDs(ids) != "1,2,3" {
		t.Fatalf("join: %q", joinScopeNodeIDs(ids))
	}
}

func TestParseScopeNodeIDsSingle(t *testing.T) {
	ids, err := parseScopeNodeIDs("42")
	if err != nil {
		t.Fatal(err)
	}
	if joinScopeNodeIDs(ids) != "42" {
		t.Fatalf("join single got %q", joinScopeNodeIDs(ids))
	}
}
