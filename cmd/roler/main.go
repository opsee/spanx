package main

import (
	"fmt"
	"github.com/opsee/spanx/policies"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "annotated" {
		fmt.Println(policies.GetPolicyWithComments())
	} else {
		fmt.Println(policies.GetPolicy())
	}
}
