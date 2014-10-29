package divsd


import "testing"

// A global test for the encoding/decoding functions
func TestSkipFields(t *testing.T) {
	postFix := "memberlist: Something"
	testStr := "2014/10/29 11:11:04 [WARN] " + postFix

	if skipFields(testStr, " ", 3) != postFix {
		t.Fatalf("Did not get expected string: '%s'", testStr)
	}
}
