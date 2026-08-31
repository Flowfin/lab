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

	// AQuarantineOutsideThePlaceQuarantinesLive refuses a directory named
	// borrowed inside an experiment that is not the one record 0019 fixes.
	// That record puts the quarantine at experiments/<slug>/borrowed and
	// allows one per experiment, and both halves of that are this one
	// property: a second quarantine is necessarily somewhere the record does
	// not put one.
	//
	// What it protects is the reason the quarantine exists at all. A person
	// promoting the work walks the tree and reads the boundary off the layout,
	// so a boundary somewhere below the first means the first one was not the
	// boundary, and the reader who stopped at it learned something untrue.
	//
	// It is refused whatever the record says, for the same reason the licence
	// arm is. The declaration names one source and one licence, so a second
	// quarantine is undeclared by construction however carefully the header
	// was written.
	AQuarantineOutsideThePlaceQuarantinesLive = "quarantine-outside-the-place-quarantines-live"

	// BorrowedDirectoryTheRecordDoesNotDeclare refuses an experiment that holds
	// a quarantine while its record declares no Borrowed field. That is the
	// second direction of the disagreement above: the tree says the experiment
	// borrows and the record says nothing, so the person promoting the work
	// reads a boundary and finds no source and no licence named anywhere they
	// would look for one.
	//
	// IT IS KEYED ON THE DIRECTORY AND NOT ON THE ABSENT FIELD, and the
	// difference is the whole reason it may exist. Record 0013 says a field
	// added to the format after it is optional and that an absent field is
	// never a refusal, which is why refuseHardware declines the identical shape
	// one field over. Nothing here reads an absence and refuses it: what is
	// refused is a directory that is present, and a tree carrying one has
	// already said it borrows without the header being consulted at all. So
	// every experiment that borrows nothing goes on writing nothing, which is
	// the property record 0013 bought and this does not spend.
	//
	// WHERE IT DOES NOT REACH. A quarantine somewhere other than
	// experiments/<slug>/borrowed is not this arm's subject - it is refused by
	// the arm above, whose repair is to move it, and this one then reads the
	// moved directory. So an experiment whose only borrowed directory is in the
	// wrong place is refused once rather than twice, and the second refusal
	// arrives with the repair rather than beside the defect.
	BorrowedDirectoryTheRecordDoesNotDeclare = "borrowed-directory-the-record-does-not-declare"
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
// WHERE IT DOES NOT REACH, in two places rather than three.
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

	// Asked whether or not the experiment holds a quarantine of its own, and
	// asked before the declaration is read, because the two arms below return
	// early on an unparseable record and a directory in the wrong place is not
	// a thing a header could excuse.
	elsewhere, err := refuseQuarantineElsewhere(fsys, root, inside, quarantine)
	if err != nil {
		return nil, err
	}
	refusals = append(refusals, elsewhere...)

	parsed, err := ParseRecord(data)
	if err != nil {
		return refusals, nil
	}
	_, declared := parsed.Field(FieldBorrowed)

	switch {
	case declared && !held:
		refusals = append(refusals, Refusal{
			Property: RecordBorrowedDeclarationNamesNoDirectory,
			Subject:  record,
			Detail: fmt.Sprintf("it declares %s and there is no %s directory in %s, so the record says the experiment borrows and the tree says it does not",
				FieldBorrowed, BorrowedDir, inside),
		})
	case held && !declared:
		refusals = append(refusals, Refusal{
			Property: BorrowedDirectoryTheRecordDoesNotDeclare,
			Subject:  at(root, quarantine),
			Detail: fmt.Sprintf("it holds code under somebody else's terms and %s declares no %s, so the tree says the experiment borrows and the record says nothing. record 0019 puts the source and the licence in that field",
				record, FieldBorrowed),
		})
	}

	return refusals, nil
}

// refuseQuarantineElsewhere walks an experiment for a directory named borrowed
// that is not the one record 0019 fixes. It takes the allowed one as a path
// rather than deriving it a second time, so the place this walks past and the
// place refuseBorrowed judges cannot drift apart.
//
// IT STOPS AT THE QUARANTINE RATHER THAN JUDGING PAST IT. What is inside the
// allowed directory is somebody else's code laid out somebody else's way, and a
// directory named borrowed in there is that code's own business. Descending
// would refuse honest work for the shape of a name this board does not own, and
// the whole point of the quarantine is that this repository's rules stop at its
// edge.
//
// WHERE IT DOES NOT REACH, in three places.
//
// Below WalkDepthBound nothing is examined, which is the bound the stray-record
// walk already carries. A tree deep enough to hide a quarantine there is refused
// by that walk for its depth, so the repair is the same one either way, and a
// second bound here would give one tree two answers about how far a walk goes.
//
// A symbolic link named borrowed is not a directory to fs.WalkDir, so it is
// neither followed nor refused here. That is the same position isDirectory
// takes at its own declaration for the same reason: a link under experiments/
// is already refused where the stray-record walk meets it, by reading the entry
// rather than the target.
//
// Nothing here opens a licence file or reads a word of one, so a quarantine in
// the right place with the wrong terms inside it is not this arm's subject and
// is not any other arm's either.
func refuseQuarantineElsewhere(fsys fs.FS, root, inside, quarantine string) ([]Refusal, error) {
	var refusals []Refusal

	err := fs.WalkDir(fsys, inside, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if name == quarantine {
			return fs.SkipDir
		}
		if depthOf(name) > WalkDepthBound {
			return fs.SkipDir
		}
		if entry.Name() != BorrowedDir {
			return nil
		}
		// Refused and not descended into. What is under it is code this board
		// has already said it will not judge, and the repair is to move the
		// directory rather than anything inside it.
		refusals = append(refusals, Refusal{
			Property: AQuarantineOutsideThePlaceQuarantinesLive,
			Subject:  at(root, name),
			Detail: fmt.Sprintf("a quarantine lives at %s and this one is at %s, so %s holds a second boundary below the first and record 0019 allows one",
				quarantine, name, inside),
		})
		return fs.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %w", at(root, inside), err)
	}
	return refusals, nil
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
