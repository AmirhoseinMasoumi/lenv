package vm

import "testing"

func TestInstanceNameDeterministic(t *testing.T) {
p := `C:\Projects\demo`
a := InstanceName(p)
b := InstanceName(p)
if a != b {
t.Fatalf("expected deterministic name, got %q and %q", a, b)
}
}

