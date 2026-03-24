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

func TestGetRuntimeStatus(t *testing.T) {
	st, err := GetRuntimeStatus()
	if err != nil {
		t.Fatalf("GetRuntimeStatus failed: %v", err)
	}
	if st.RootDir == "" {
		t.Fatal("expected non-empty runtime root")
	}
	if st.ManagedDir == "" {
		t.Fatal("expected non-empty managed dir")
	}
}

func TestRuntimeManifestURLsDefault(t *testing.T) {
	if managedQEMUManifestURL() == "" {
		t.Fatal("managedQEMUManifestURL should not be empty")
	}
	if managedQEMUManifestSigURL() == "" {
		t.Fatal("managedQEMUManifestSigURL should not be empty")
	}
}
