// The leg that holds record 0011's contract to itself.
//
// The numbers are declared in more than one place on purpose. A code lives next
// to the thing that can return it, which is what keeps the code the hardware
// harness returns out of the runner rather than written there as a branch
// nothing can reach. What that shape costs is that each declaration is right
// about itself and nothing was right about all of them together: a number that
// moved in one of them compiled, passed the tests of the package it moved in,
// and changed the meaning of every reader keyed on it without their files being
// edited.
//
// This leg is the other half of that shape rather than a replacement for it.
// The declarations stay where their producers are and this reads them all and
// requires them to say one thing.

package invariants

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// goExtension is what this leg reads. Everything else the scan carries is prose
// and declares no constants.
const goExtension = ".go"

// exitCodeName is how a declaration of one of record 0011's codes is
// recognised. It is a naming convention read out of the tree rather than a list
// of files, so a fourth declaration written somewhere nobody has thought of yet
// is held to the same contract as the three that exist, and this leg does not
// have to be edited to notice it.
//
// The cost of a convention is that it reaches only what follows it. A code
// declared under a name that does not read like one is invisible here, and
// nothing in this tree can see that it is a code, which is why the convention is
// stated at each declaration rather than only in this file.
var exitCodeName = regexp.MustCompile(`^[Ee]xit[A-Z]`)

// An exitCodeDeclaration is one exit-code constant as one file declares it.
type exitCodeDeclaration struct {
	// name is the constant, which is what groups two declarations of one code.
	name string

	// value is the number, which is what the contract fixes.
	value int

	// path is the file as the walk reached it, which is what a reader opens.
	path string

	// line is where in that file, because a reader hitting this refusal is
	// looking for two numbers and needs to be sent to both of them.
	line int
}

// exitCodesLeg holds every declaration of record 0011's contract to every other
// one, in both directions a disagreement can run.
//
// THE BOUND, and it is the one this package's harness already carries for the
// notice leg. Both directions refuse under one property, because the repair is
// one repair: make the declarations say what the record says. A case comparing
// property sets cannot tell them apart, so a change that loses one of the two
// stays green on the other's case. Each direction has its own case anyway, and
// that is a convention here rather than something the harness requires.
func exitCodesLeg(texts []textFile) (Leg, []Refusal) {
	declarations, examined := exitCodeDeclarations(texts)
	leg := Leg{Name: "the exit codes", Asked: true, Examined: examined}

	var refusals []Refusal
	refusals = append(refusals, oneCodeWithTwoNumbers(declarations)...)
	refusals = append(refusals, oneNumberWithTwoCodes(declarations)...)
	return leg, refusals
}

// oneCodeWithTwoNumbers refuses a code whose declarations do not agree on the
// number. This is the drift the shape invites: the constant is spelled the same
// everywhere, so a reader comparing two files by eye is looking at the one
// character that differs.
func oneCodeWithTwoNumbers(declarations []exitCodeDeclaration) []Refusal {
	first := make(map[string]exitCodeDeclaration)
	var refusals []Refusal

	for _, declaration := range declarations {
		reference, seen := first[declaration.name]
		if !seen {
			first[declaration.name] = declaration
			continue
		}
		if declaration.value == reference.value {
			continue
		}
		refusals = append(refusals, Refusal{
			Property: ExitCodeDeclarationsDisagree,
			Subject:  declaration.path,
			Detail: fmt.Sprintf("it declares %s as %d at line %d, and %s declares the same code as %d, so the contract record 0011 fixes has two numbers in this tree",
				declaration.name, declaration.value, declaration.line, reference.path, reference.value),
		})
	}
	return refusals
}

// oneNumberWithTwoCodes refuses one number declared under two codes. This is
// the direction that reaches a code declared exactly once, which the direction
// above cannot: a number written into a new constant collides with the meaning
// the record already gave it, and there is no second declaration of that
// constant for anything to compare it against.
func oneNumberWithTwoCodes(declarations []exitCodeDeclaration) []Refusal {
	first := make(map[int]exitCodeDeclaration)
	var refusals []Refusal

	for _, declaration := range declarations {
		reference, seen := first[declaration.value]
		if !seen {
			first[declaration.value] = declaration
			continue
		}
		if declaration.name == reference.name {
			continue
		}
		refusals = append(refusals, Refusal{
			Property: ExitCodeDeclarationsDisagree,
			Subject:  declaration.path,
			Detail: fmt.Sprintf("it declares %s as %d at line %d, and %s declares %s as the same number, so one code in record 0011's contract answers to two names",
				declaration.name, declaration.value, declaration.line, reference.path, reference.name),
		})
	}
	return refusals
}

// exitCodeDeclarations reads every Go file the scan carries and returns the
// exit-code constants it declares, with how many files were parsed.
//
// It reads the text the scan already read rather than the filesystem again, so
// the declarations it compares are the ones in the tree the report is about.
func exitCodeDeclarations(texts []textFile) ([]exitCodeDeclaration, int) {
	var found []exitCodeDeclaration
	parsed := 0

	for _, file := range texts {
		if !strings.HasSuffix(file.relative, goExtension) {
			continue
		}
		fset := token.NewFileSet()
		syntax, err := parser.ParseFile(fset, file.relative, file.text, 0)
		if err != nil {
			// A file the parser cannot read is not judged here. The compiler
			// refuses it, and a leg that refused it too would be a second copy
			// of a rule that already has a better mechanism.
			continue
		}
		parsed++

		for _, declaration := range syntax.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.CONST {
				continue
			}
			for _, spec := range group.Specs {
				values, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range values.Names {
					if i >= len(values.Values) || !exitCodeName.MatchString(name.Name) {
						continue
					}
					number, ok := integerLiteral(values.Values[i])
					if !ok {
						continue
					}
					found = append(found, exitCodeDeclaration{
						name:  name.Name,
						value: number,
						path:  file.path,
						line:  fset.Position(name.Pos()).Line,
					})
				}
			}
		}
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].path != found[j].path {
			return found[i].path < found[j].path
		}
		return found[i].line < found[j].line
	})
	return found, parsed
}

// integerLiteral reads a constant's value where that value is a plain number,
// and says no for everything else.
//
// A constant whose name reads like an exit code and whose value is not a number
// is not an exit code. That is not a hypothetical: the property identifier this
// leg refuses under is spelled ExitCodeDeclarationsDisagree, it sits in this
// package, and a leg that took every matching name would refuse this repository
// for declaring the rule it is refusing under.
func integerLiteral(expression ast.Expr) (int, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return 0, false
	}
	number, err := strconv.Atoi(literal.Value)
	if err != nil {
		return 0, false
	}
	return number, true
}
