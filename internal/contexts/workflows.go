package contexts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WorkflowsDir is where the platform reads workflows from, and therefore the
// only directory a declared check name can come from.
const WorkflowsDir = ".github/workflows"

// ReadWorkflows reads every workflow file under the directory it is given and
// returns the check names they declare.
//
// WHY THIS READS THE FILES RATHER THAN A COMPLETED RUN. A run reports the names
// of the jobs that ran on it, which is a different set: a job filtered out by a
// condition reports nothing, and a workflow added on a branch reports nothing
// until it runs there. What the gate has to agree with is what the tree
// declares, because the tree is the thing a pull request changes and the rename
// this check exists to catch is a change to a file.
//
// WHAT IT UNDERSTANDS, AND WHERE IT STOPS. This module carries no dependencies
// and adding a YAML parser to read a dozen files of one known shape is a runtime
// cost paid for something narrower than the library. So it reads the shape the
// workflows in this tree are written in, by line and by indentation: a job is a
// key two spaces in under `jobs:`, its name is `name:` four spaces in, and a
// matrix is an `include:` list of flat key-and-value entries. It stops reading a
// job's own keys at `steps:`, because everything below that belongs to a step
// rather than to the job.
//
// Anything it does not understand is a refusal rather than a guess. A name it
// cannot finish expanding keeps its expression and Judge refuses it by name, so
// a workflow written in a shape this reader was not built for reddens the check
// instead of quietly declaring the wrong string.
func ReadWorkflows(dir string) ([]Declared, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, which is where a declared check name can only come from: %w", dir, err)
	}

	var declared []Declared
	for _, entry := range entries {
		if entry.IsDir() || !isWorkflowFile(entry.Name()) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", path, err)
		}
		found, err := readWorkflow(entry.Name(), string(data))
		if err != nil {
			return nil, err
		}
		declared = append(declared, found...)
	}

	if len(declared) == 0 {
		// A reader that returned nothing would make every deliberate absence
		// dangle and every required context unreported, which is a wall of
		// refusals describing a reader that found no files rather than a tree
		// that is wrong. Say the real thing instead.
		return nil, fmt.Errorf("%s declared no check name at all, so either there are no workflows there or this reader did not understand any of them", dir)
	}

	sort.Slice(declared, func(i, j int) bool {
		if declared[i].Workflow != declared[j].Workflow {
			return declared[i].Workflow < declared[j].Workflow
		}
		return declared[i].Name < declared[j].Name
	})
	return declared, nil
}

func isWorkflowFile(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// job is one job as this reader understood it.
type job struct {
	id      string
	name    string
	named   bool
	matrix  []map[string]string
	inSteps bool
}

// readWorkflow returns the check names one workflow file declares.
func readWorkflow(file, text string) ([]Declared, error) {
	var jobs []*job
	var current *job

	inJobs := false
	includeIndent := -1

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))

		if indent == 0 {
			inJobs = trimmed == "jobs:"
			current = nil
			includeIndent = -1
			continue
		}
		if !inJobs {
			continue
		}

		if indent == 2 && strings.HasSuffix(trimmed, ":") {
			current = &job{id: strings.TrimSuffix(trimmed, ":")}
			jobs = append(jobs, current)
			includeIndent = -1
			continue
		}
		if current == nil {
			continue
		}

		// Inside an include list: entries are flat key-and-value pairs, a new
		// entry starts at a dash, and the list ends at the first line that is
		// not deeper than the `include:` key itself.
		if includeIndent >= 0 {
			if indent > includeIndent {
				if err := readMatrixEntry(current, trimmed, file); err != nil {
					return nil, err
				}
				continue
			}
			includeIndent = -1
		}

		if indent == 4 {
			// Everything under steps: belongs to a step. A step carries a
			// name of its own and reading one as the job's would declare a
			// check name that no gate will ever see.
			if trimmed == "steps:" {
				current.inSteps = true
				continue
			}
			if !current.inSteps && strings.HasPrefix(trimmed, "name:") {
				current.name = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")))
				current.named = true
			}
			continue
		}
		if !current.inSteps && trimmed == "include:" {
			includeIndent = indent
		}
	}

	var declared []Declared
	for _, j := range jobs {
		name := j.name
		if !j.named {
			// The platform reports the job id where a job carries no name of
			// its own, so that is the string a required context would have to
			// match.
			name = j.id
		}
		expanded, err := expand(name, j.matrix, file)
		if err != nil {
			return nil, err
		}
		for _, one := range expanded {
			declared = append(declared, Declared{Name: one, Workflow: file, FromJobID: !j.named})
		}
	}
	return declared, nil
}

// readMatrixEntry reads one `key: value` line of an include list into the entry
// it belongs to.
func readMatrixEntry(j *job, trimmed, file string) error {
	if strings.HasPrefix(trimmed, "- ") {
		j.matrix = append(j.matrix, map[string]string{})
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
	}
	if len(j.matrix) == 0 {
		return fmt.Errorf("%s: the job %q has a line in its include list before any entry started, which is a shape this reader was not built for", file, j.id)
	}
	key, value, found := strings.Cut(trimmed, ":")
	if !found {
		return fmt.Errorf("%s: the job %q has the include line %q, which carries no key and value, and this reader will not guess at it", file, j.id, trimmed)
	}
	j.matrix[len(j.matrix)-1][strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
	return nil
}

// expand turns a job name carrying matrix expressions into one name per matrix
// entry. A name with no expression is itself, whether or not the job has a
// matrix, because a matrix job whose name mentions no matrix key reports one
// check name per entry under the same string and the gate sees one context.
func expand(name string, matrix []map[string]string, file string) ([]string, error) {
	if !strings.Contains(name, "${{") {
		return []string{name}, nil
	}
	if len(matrix) == 0 {
		// The name is left as it stands, expression included, and Judge
		// refuses it by name. Returning an error here would stop the whole
		// run on one file, and a refusal naming the workflow is what a reader
		// can act on.
		return []string{name}, nil
	}
	var out []string
	seen := map[string]bool{}
	for _, entry := range matrix {
		expanded := name
		for key, value := range entry {
			expanded = strings.ReplaceAll(expanded, "${{ matrix."+key+" }}", value)
			expanded = strings.ReplaceAll(expanded, "${{matrix."+key+"}}", value)
		}
		if seen[expanded] {
			continue
		}
		seen[expanded] = true
		out = append(out, expanded)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: the name %q expanded to nothing", file, name)
	}
	return out, nil
}

func unquote(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
