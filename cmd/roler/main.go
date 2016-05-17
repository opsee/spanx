package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/opsee/spanx/policies"
	"github.com/opsee/spanx/roler"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(policies.GetPolicy())
		return
	}

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "annotated":
			fmt.Println(policies.GetPolicyWithComments())
		case "stack":
			tmpl := template.Must(template.New("role").Parse(roler.RoleTemplate))
			if err != nil {
				log.Fatal("couldn't parse role template: ", err)
			}

			var out bytes.Buffer
			err = tmpl.Execute(&out, struct {
				ExternalID string
			}{"YOUR_OPSEE_EXTERNAL_ID"})
			if err != nil {
				log.Fatal("couldn't execute role template: ", err)
			}

			fmt.Println(out.String())
		}
	}
}
