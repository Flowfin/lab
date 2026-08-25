package check

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
)

// FieldBorrowed names where borrowed code came from and the licence it arrives
// under. Record 0019 adds it and record 0013 makes it optional, as it makes
// every field added after it, so an experiment that borrows nothing writes
// nothing and that is never a refusal.
//
// It is declared here rather than beside the four record 0008 fixes, which is
// the shape FieldMeasurementCommit already argues for at its own declaration: a
// field a later record adds arrives with the check that reads it, so a checker
// built before the field is unaware of it and a field with no check has nowhere
// to hide.
const FieldBorrowed = "Borrowed"

// BorrowedDir is the one directory inside an experiment that may hold code
// under a licence that is not this board's, and BorrowedLicenceName is the file
// that says which licence that is. Record 0019 fixes both names and puts the
// directory inside the experiment rather than at the root, because record 0002
// refuses a root directory it does not name.
const (
	BorrowedDir         = "borrowed"
	BorrowedLicenceName = "LICENSE"
)

// The properties a borrowed quarantine can be refused for.
const (
	// BorrowedDirectoryCarriesNoLicence refuses a borrowed directory with no
	// licence file in it. That is the quarantine without the thing that makes
	// it one: a directory a reader walking the tree takes for code under
	// somebody else's terms, declaring none of them, so the person promoting
	// the work later reads the boundary and learns nothing from it.
	//
	// It is refused whatever the record says, because the layout is what
	// record 0019 buys and a header field is visible only to somebody who
	// opened the header.
	BorrowedDirectoryCarriesNoLicence = "borrowed-directory-carries-no-licence"

	// RecordBorrowedDeclarationNamesNoDirectory refuses a record that declares
	// Borrowed while the experiment holds no borrowed directory. The record
	// then says code under other terms is in the experiment and the tree says
	// it is not, and which of the two is wrong decides the repair, so the
	// message names both sides rather than restating the rule.
	//
	// The declaration having been made is what this reads, so record 0013 is
	// untouched by it: an absent field says nothing and is not refused here.
	RecordBorrowedDeclarationNamesNoDirectory = "record-borrowed-declaration-names-no-directory"
)

// refuseBorrowed holds an experiment's borrowed quarantine and its record's
// Borrowed declaration to each other and to record 0019's layout.
//
// WHAT A GREEN RUN DOES NOT SAY, and it is most of what somebody wants when
// they read the word licence. Whether the licence file names the licence the
// code is actually under is a judgement about the world rather than about the
// tree, and nothing here opens that file or reads a word of it. So is whether
// the borrowed code may be promoted into a board under other terms, which
// record 0019 explicitly leaves undecided. What passes here is a layout and a
// declaration that do not contradict each other, and nothing further.
//
// WHERE IT DOES NOT REACH, in three places rather than one.
//
// A borrowed directory in an experiment whose record declares no Borrowed field
// passes. That is the second direction of the disagreement above and it is a
// refusal on an absent field, which record 0013 forbids in the words
// refuseHardware already declines the same shape in. Closing it is a change to
// that record rather than a wider check here.
//
// A Borrowed declaration written with nothing after the colon is a declaration,
// so the directory half above is read against it, and nothing here asks whether
// it names a source or a licence. Record 0019 says the field names both;
// whether it does is prose a reader judges.
//
// An experiment whose record the walk did not reach - absent, unreadable, above
// the size bound - is not judged here at all, because this is called with the
// record's bytes in hand and those runs never get that far.
//
// A record whose bytes do not parse as a record is judged for its directory and
// not for its declaration, for the reason refuseState and refuseHardware both
// give: nothing can read a field out of a file that has no header.
func refuseBorrowed(fsys fs.FS, root, inside, record string, data []byte) ([]Refusal, error) {
	quarantine := path.Join(inside, BorrowedDir)

	held, err := isDirectory(fsys, quarantine)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", at(root, quarantine), err)
	}

	var refusals []Refusal

	if held {
		carried, err := isRegularFile(fsys, path.Join(quarantine, BorrowedLicenceName))
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", at(root, path.Join(quarantine, BorrowedLicenceName)), err)
		}
		if !carried {
			refusals = append(refusals, Refusal{
				Property: BorrowedDirectoryCarriesNoLicence,
				Subject:  at(root, quarantine),
				Detail: fmt.Sprintf("it holds no %s, so it reads as code under somebody else's terms and names none of them. record 0019 puts that file at %s",
					BorrowedLicenceName, path.Join(quarantine, BorrowedLicenceName)),
			})
		}
	}

	parsed, err := ParseRecord(data)
	if err != nil {
		return refusals, nil
	}
	if _, declared := parsed.Field(FieldBorrowed); !declared || held {
		return refusals, nil
	}

	return append(refusals, Refusal{
		Property: RecordBorrowedDeclarationNamesNoDirectory,
		Subject:  record,
		Detail: fmt.Sprintf("it declares %s and there is no %s directory in %s, so the record says the experiment borrows and the tree says it does not",
			FieldBorrowed, BorrowedDir, inside),
	}), nil
}

// isDirectory says whether a name in the walked filesystem is a directory. A
// name that is not there is not one, and that is the ordinary answer rather
// than an error: almost every experiment borrows nothing.
//
// THIS FOLLOWS A SYMBOLIC LINK AND DOES NOT REFUSE ONE. fs.Stat resolves a
// link, so a link named borrowed that points at a directory reads as a
// quarantine here and its target is stated to be inside the experiment by
// nothing. That is deliberately not repaired at this site: a link anywhere
// under experiments/ is already refused by the stray-record walk, which reads
// the entry rather than the target, and a second rule deciding what a link
// points at is the resolution that walk exists to avoid.
func isDirectory(fsys fs.FS, name string) (bool, error) {
	info, err := fs.Stat(fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// isRegularFile says whether a name in the walked filesystem is a file a reader
// can open. A licence that is a directory, or a link, is not one, and it is the
// same answer as a licence that is not there: in both cases nothing at that
// path states terms.
func isRegularFile(fsys fs.FS, name string) (bool, error) {
	info, err := fs.Stat(fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.Mode().IsRegular(), nil
}
