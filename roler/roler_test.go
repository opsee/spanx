package roler

import (
	"testing"
)

func TestParseAccountARN(t *testing.T) {
	acc, err := parseARNAccount("arn:aws:iam::975383256012:root")
	if err != nil {
		t.Fatal(err)
	}

	if acc != 975383256012 {
		t.Fatal("account didn't match")
	}

	acc, err = parseARNAccount("arn:aws:iam::933693344490:user/mark")
	if err != nil {
		t.Fatal(err)
	}

	if acc != 933693344490 {
		t.Fatal("account didn't match")
	}
}
