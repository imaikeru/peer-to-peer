package validator_test

import (
	"fmt"
	"testing"

	"github.com/imaikeru/peer-to-peer/client/validator"
)

func TestValidateCommandsTableDriven(t *testing.T) {

	testValidator := validator.CreateValidator()

	var tests = []struct {
		command string
		isValid bool
	}{
		{"   disconnect            ", true},
		{"disconnect", true},
		{"   disconnect       be     ", false},
		{"be disconnect", false},
		{"  list-files    ", true},
		{"list-files", true},
		{" dsf list-files    ", false},
		{"list-files sfdf", false},
		{`download ivancho "E:\ivancho\file1.txt" "E:\petio\file1copy.txt"`, true},
		{`download ivancho "E:\ivancho\file1.txt" "E:\petio\file1copy.txt" download`, false},
		{`register ivancho "file1" "file2" "file3"`, true},
		{`register ivancho "file1" "file2" "file3" file 5`, false},
		{`unregister ivancho "file1" "file2" "file3" file4`, false},
		{`unregister ivancho "file1" "file2" "file3"`, true},
		{` asdkalsdkl `, false},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("validate(%s) should return %t", tt.command, tt.isValid)
		t.Run(testname, func(t *testing.T) {
			ans := testValidator.Validate(tt.command)
			if ans != tt.isValid {
				t.Errorf("got %t, want %t", ans, tt.isValid)
			}
		})
	}
}
