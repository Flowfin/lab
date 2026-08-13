package check

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// HarnessBuildTag is the build constraint the integration-hardware harness's
// tests sit behind, and HarnessTestSuffix is what a file holding one is called.
//
// Both are written here rather than imported from internal/hardware, because
// that package imports testing and the runner is a program somebody runs rather
// than a test binary. TestTheHarnessNamesAreTheHarnessesOwn imports it from the
// suite and holds the two equal, so the copy cannot drift in silence.
const (
	HarnessBuildTag   = "integration_hardware"
	HarnessTestSuffix = "_" + HarnessBuildTag + "_test.go"
)

// The properties a record's hardware declaration can be refused for.
const (
	// RecordHardwareDeclarationIsEmpty refuses a record that declares
	// Needs-Hardware and writes nothing after the colon. Record 0013 makes an
	// absent field legal and an empty declaration a different statement, and
	// this is the second one: it names no hardware and it does not say none,
	// so neither direction of the rule below has anything to compare against.
	// Reading it as one of the two would be a guess written into a checker.
	RecordHardwareDeclarationIsEmpty = "record-hardware-declaration-is-empty"

	// RecordHardwareDeclarationDisagreesWithItsTests refuses a declaration the
	// experiment's own directory contradicts, at two sites.
	//
	// A record naming hardware whose directory registers no test with the
	// harness is the ordinary half. The tests moved into the default run once
	// they stopped needing the device and the record still sends a reader
	// looking for one.
	//
	// A record saying none whose directory registers a test with the harness is
	// the half worth the fixture. It passes every other rule in this tree, it
	// reads as a result anybody can reproduce, and the person who finds out
	// otherwise is whoever cloned the repository to check the answer.
	RecordHardwareDeclarationDisagreesWithItsTests = "record-hardware-declaration-disagrees-with-its-tests"
)

// refuseHardware holds a record's hardware declaration to what its directory
// registers with the integration-hardware harness.
//
// WHAT IT DOES NOT JUDGE, and this is most of what the field says. Whether the
// words name the hardware the tests actually need is a reader's judgement, and
// a declaration of "a laptop" satisfies every rule here. What a green run says
// is that the record and the directory do not contradict each other.
//
// WHERE IT DOES NOT REACH. An absent field is never refused, which record 0013
// fixes, so an experiment carrying hardware tests and declaring nothing at all
// passes this. That is the price of a format that can grow without turning
// older records red, it is paid deliberately, and the template is what closes
// the gap for records written from here on rather than a refusal.
//
// A record whose bytes do not parse as a record is not judged here, for the
// reason refuseState and refuseHeaderDates both give: nothing can read a field
// out of a file that has no header.
func refuseHardware(fsys fs.FS, inside, experiment, path string, data []byte) ([]Refusal, error) {
	record, err := ParseRecord(data)
	if err != nil {
		return nil, nil
	}

	declared, present := record.Field(FieldNeedsHardware)
	if !present {
		return nil, nil
	}

	if strings.TrimSpace(declared) == "" {
		return []Refusal{{
			Property: RecordHardwareDeclarationIsEmpty,
			Subject:  path,
			Detail: fmt.Sprintf("it declares %s and writes nothing after the colon, so it names no hardware and does not say %s",
				FieldNeedsHardware, HardwareNone),
		}}, nil
	}

	registered, err := harnessTestsUnder(fsys, inside, experiment)
	if err != nil {
		return nil, err
	}

	if declared == HardwareNone {
		if len(registered) == 0 {
			return nil, nil
		}
		return []Refusal{{
			Property: RecordHardwareDeclarationDisagreesWithItsTests,
			Subject:  path,
			Detail: fmt.Sprintf("it says %s: %s and %s is registered with the %s harness, so it reads as a result anybody can reproduce",
				FieldNeedsHardware, HardwareNone, strings.Join(registered, ", "), HarnessBuildTag),
		}}, nil
	}

	if len(registered) > 0 {
		return nil, nil
	}
	return []Refusal{{
		Property: RecordHardwareDeclarationDisagreesWithItsTests,
		Subject:  path,
		Detail: fmt.Sprintf("it says %s: %s and no file under %s is named %s, so nothing in it is registered with that harness. a record that needs nothing says %s",
			FieldNeedsHardware, declared, filepath.ToSlash(experiment), "*"+HarnessTestSuffix, HardwareNone),
	}}, nil
}

// harnessTestsUnder returns the files in an experiment directory that register
// a test with the integration-hardware harness, in the order the walk reached
// them, relative to that directory.
//
// Registration is read from the file's name and from nothing else. That is the
// convention internal/hardware already holds its own files to, and the
// alternative is a Go parser pointed at whatever somebody put in an experiment
// directory, on a runner whose input is untrusted. A file carrying the build
// constraint under some other name is not registered with the harness by any
// route this repository has, so there is nothing there to miss.
//
// The walk is bounded below the experiment directory by the same depth the
// stray-record walk uses, for the same reason: the cost of reading a tree is
// whatever the tree makes it, and a bound is what keeps a run finite against a
// directory nobody wrote by hand. A directory below the bound is not descended
// into, and TheTreeIsDeeperThanTheWalkReads is what refuses the tree that
// reaches it.
func harnessTestsUnder(fsys fs.FS, inside, experiment string) ([]string, error) {
	var registered []string

	// A tree walk does not follow symbolic links, so a link pointing out of
	// the experiment is reported as a link and never descended into. What
	// refuses such a link is the stray-record walk, which reaches every path
	// under experiments/; this reading leaves it alone and registers nothing
	// for it, because a name is all this reads and a link's name says nothing
	// about what it points at.
	err := fs.WalkDir(fsys, inside, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if depthOf(name)-depthOf(inside) > WalkDepthBound {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), HarnessTestSuffix) {
			return nil
		}
		registered = append(registered, strings.TrimPrefix(name, inside+"/"))
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot walk %s: %w", experiment, err)
	}
	return registered, nil
}
