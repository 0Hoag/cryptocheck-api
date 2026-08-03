package mongo

import "testing"

func TestPostSort(t *testing.T) {
	if got := postSort(""); got[0].Key != "created_at" || got[0].Value != -1 {
		t.Fatalf("default post sort = %#v, want newest-first", got)
	}
	if got := postSort("newest"); got[0].Value != -1 {
		t.Fatalf("newest sort = %#v, want descending", got)
	}
	if got := postSort("oldest"); got[0].Value != 1 {
		t.Fatalf("oldest sort = %#v, want ascending", got)
	}
}
