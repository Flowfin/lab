package check

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// producibleProperties reads this package's own source and returns every
// property a refusal site names.
//
// It is derived rather than listed. A list of properties written by hand is a
// list that stops being true the moment somebody adds a refusal site and
// forgets it, and the whole point of the requirement below is that it reaches
// refusals nobody has written yet. What the source says a site can produce is
// the only thing that cannot fall behind the source.
//
// It reads the package's non-test files, finds every composite literal field
// named Property, and resolves the constant it names to the string that
// constant holds. A site naming a string literal directly is resolved too. A
// site naming anything else fails the test rather than being skipped, because
// a property this cannot resolve is a property nothing below can require a
// fixture for.
func producibleProperties(t *testing.T) []string {
	t.Helper()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("cannot parse this package: %v", err)
	}

	constants := make(map[string]string)
	var sites []ast.Expr

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.ValueSpec:
					for i, name := range node.Names {
						if i >= len(node.Values) {
							continue
						}
						if lit, ok := node.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							value, err := strconv.Unquote(lit.Value)
							if err == nil {
								constants[name.Name] = value
							}
						}
					}
				case *ast.KeyValueExpr:
					if key, ok := node.Key.(*ast.Ident); ok && key.Name == "Property" {
						sites = append(sites, node.Value)
					}
				}
				return true
			})
		}
	}

	if len(sites) == 0 {
		t.Fatal("no refusal site in this package names a property, which cannot be right while the package refuses anything")
	}

	seen := make(map[string]bool)
	var props []string
	for _, site := range sites {
		var value string
		switch expr := site.(type) {
		case *ast.Ident:
			v, ok := constants[expr.Name]
			if !ok {
				t.Fatalf("a refusal site names %s and this cannot resolve it to a property", expr.Name)
			}
			value = v
		case *ast.BasicLit:
			v, err := strconv.Unquote(expr.Value)
			if err != nil {
				t.Fatalf("a refusal site names %s and this cannot read it: %v", expr.Value, err)
			}
			value = v
		default:
			t.Fatalf("a refusal site at %s names a property this cannot resolve", fset.Position(site.Pos()))
		}
		if !seen[value] {
			seen[value] = true
			props = append(props, value)
		}
	}
	sort.Strings(props)
	return props
}

// TestEveryRefusalHasAFixtureThatProvesItBites is the standing requirement. It
// reaches refusals nobody has written yet, because the property list is read
// out of the source rather than out of a list somebody maintains.
//
// Three legs, and a refusal is only proved when it has all three.
//
// The first leg is a fixture whose whole refusal set is exactly that one
// property. Comparing the whole set is what makes the leg mean anything: a
// fixture that also trips a second refusal proves neither of them cleanly.
//
// The second leg is a near neighbour of that fixture, differing by the
// smallest change that makes it legal, refusing nothing. Without it, a check
// that refuses everything passes its own test.
//
// The third leg is that the refusal names its subject, which
// TestARefusalNamesItsSubject holds over every case in the tree.
//
// THE BOUND. Every leg compares which properties were refused, and never
// which line inside the runner refused them. Two refusal sites producing one
// property are indistinguishable to all three, so an operator holding several
// of them can lose one and stay green. That is not hypothetical here:
// record-is-not-text is produced at two sites, one for a null byte and one for
// an invalid sequence, and a fixture for either satisfies this test for both.
// What catches the loss of one arm is the case for that arm existing, which is
// a fixture somebody remembered rather than a thing this test can require.
func TestEveryRefusalHasAFixtureThatProvesItBites(t *testing.T) {
	cases := loadCases(t)

	provers := make(map[string][]string)
	for name, exp := range cases {
		if len(exp.refusals) == 1 {
			provers[exp.refusals[0]] = append(provers[exp.refusals[0]], name)
		}
	}

	for _, property := range producibleProperties(t) {
		names := provers[property]
		if len(names) == 0 {
			t.Errorf("%s can be refused and no case declares exactly that refusal and no other", property)
			continue
		}
		sort.Strings(names)

		for _, name := range names {
			neighbour := readNearNeighbour(t, name)
			if neighbour == "" {
				t.Errorf("case %s proves %s and declares no near neighbour, so nothing here would notice a check that refuses everything", name, property)
				continue
			}
			exp, ok := cases[neighbour]
			if !ok {
				t.Errorf("case %s names %s as its near neighbour and there is no such case", name, neighbour)
				continue
			}
			if len(exp.refusals) != 0 {
				t.Errorf("case %s names %s as its near neighbour, and that case expects %v rather than nothing", name, neighbour, exp.refusals)
			}
		}
	}
}

// TestANearNeighbourIsDeclaredOnlyWhereItMeansSomething refuses a near
// neighbour declared by a case that refuses nothing. Such a declaration reads
// as a proof and is not one, since a fixture that refuses nothing has no
// refusal for a neighbour to be near to.
func TestANearNeighbourIsDeclaredOnlyWhereItMeansSomething(t *testing.T) {
	for name, exp := range loadCases(t) {
		if len(exp.refusals) == 0 && readNearNeighbour(t, name) != "" {
			t.Errorf("case %s refuses nothing and declares a near neighbour", name)
		}
	}
}

// readNearNeighbour returns the case a refusing case declares as its near
// neighbour, or the empty string when it declares none.
func readNearNeighbour(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(casesDir, name, "near-neighbour"))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("case %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}
