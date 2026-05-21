package auth

import "testing"

func TestTokenManager(t *testing.T) {
	tm := NewTokenManager("my-secret")

	tok, err := tm.Issue(10)
	if err != nil {
		t.Fatal(err)
	}

	uid, err := tm.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if uid != 10 {
		t.Errorf("uid = %d, want 10", uid)
	}
}

func TestTokenManager_BadSecret(t *testing.T) {
	tm1 := NewTokenManager("aaa")
	tm2 := NewTokenManager("bbb")

	tok, _ := tm1.Issue(1)
	_, err := tm2.Verify(tok)
	if err == nil {
		t.Error("should fail with different secret")
	}
}
