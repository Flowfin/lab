package contexts

import (
	"reflect"
	"testing"
)

// TestTheReaderExpandsAMatrixName pins what a matrix job declares. Six entries
// under one name are six contexts a gate has to hold, and a reader that returned
// the unexpanded string would compare against a name no ruleset can ever carry.
func TestTheReaderExpandsAMatrixName(t *testing.T) {
	const file = `
jobs:
  build:
    name: build (${{ matrix.goos }}/${{ matrix.goarch }})
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: windows
            goarch: arm64
    steps:
      - name: Build
        run: echo build
`
	declared, err := readWorkflow("build.yml", file)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var got []string
	for _, entry := range declared {
		got = append(got, entry.Name)
	}
	want := []string{"build (linux/amd64)", "build (windows/arm64)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("read %v, want %v", got, want)
	}
}

// TestTheReaderDoesNotReadAStepNameAsAJobName is the near-miss worth spending
// the effort on. Both keys are spelled `name:` and both sit under the same job,
// so a reader keyed on the word rather than on where it sits would declare a
// step's name as a check name, and that string would then be compared against a
// gate and found missing for a reason nobody could follow.
func TestTheReaderDoesNotReadAStepNameAsAJobName(t *testing.T) {
	const file = `
jobs:
  records:
    name: records
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@0000000000000000000000000000000000000000
      - name: Walk the records
        run: go run ./cmd/lab check .
`
	declared, err := readWorkflow("records.yml", file)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(declared) != 1 || declared[0].Name != "records" {
		t.Errorf("read %+v, and the only check name this file declares is \"records\"", declared)
	}
}

// TestAJobWithNoNameDeclaresItsJobID pins the platform's own fallback, which is
// the rule a required context has to match against dependency-review.yml.
func TestAJobWithNoNameDeclaresItsJobID(t *testing.T) {
	const file = `
jobs:
  dependency-review:
    runs-on: ubuntu-latest
    steps:
      - name: Dependency review
        uses: actions/dependency-review-action@0000000000000000000000000000000000000000
`
	declared, err := readWorkflow("dependency-review.yml", file)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(declared) != 1 || declared[0].Name != "dependency-review" || !declared[0].FromJobID {
		t.Errorf("read %+v, want the job id declared and marked as coming from the job id", declared)
	}
}

// TestAWorkflowNameIsNotAJobName pins that the string at the top of the file is
// the workflow's rather than a job's. A gate holds check names, and a reader
// that declared the workflow name would put a string in the comparison that no
// check run is ever called.
func TestAWorkflowNameIsNotAJobName(t *testing.T) {
	const file = `
name: Build and test

on:
  pull_request:

jobs:
  vet:
    name: vet
    runs-on: ubuntu-latest
    steps:
      - name: Vet
        run: go vet ./...
`
	declared, err := readWorkflow("build.yml", file)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(declared) != 1 || declared[0].Name != "vet" {
		t.Errorf("read %+v, want only the job name", declared)
	}
}

// TestThisRepositoryDeclaresTheNamesItsWorkflowsCarry reads the real directory
// rather than a string, so a workflow written in a shape this reader was not
// built for is caught here rather than as a wall of refusals on a pull request.
func TestThisRepositoryDeclaresTheNamesItsWorkflowsCarry(t *testing.T) {
	declared, err := ReadWorkflows(workflowsInThisRepository)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, entry := range declared {
		if entry.Name == "" {
			t.Errorf("%s declares an empty check name", entry.Workflow)
		}
	}
	t.Logf("this tree declares %d check name(s)", len(declared))
}
