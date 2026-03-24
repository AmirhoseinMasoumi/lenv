package vm

import "testing"

func TestManagedRuntimeURLsDefault(t *testing.T) {
	u := managedQEMUURL()
	if u == "" {
		t.Fatal("managedQEMUURL should not be empty")
	}
	sum := managedQEMUChecksumURL()
	if sum == "" {
		t.Fatal("managedQEMUChecksumURL should not be empty")
	}
}
