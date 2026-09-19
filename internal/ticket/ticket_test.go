package ticket

import "testing"

func TestValidation(t *testing.T) {
	if ValidateCreate("", StatusOpen, PriorityNormal) == nil {
		t.Fatal("empty subject accepted")
	}
	if ValidateCreate("subject", Status("bad"), PriorityNormal) == nil {
		t.Fatal("bad status accepted")
	}
	if ValidateComment(" ") == nil {
		t.Fatal("blank comment accepted")
	}
}
