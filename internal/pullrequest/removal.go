package pullrequest

// The other half of the rule this board's central claim rests on. A record is
// not rewritten to hide what it said, which record.go holds, and it is not
// deleted, which is here.
//
// Deleting was the cheaper of the two and nothing refused it. A change that
// removes an EXPERIMENT.md, or the directory it sits in, passes every check
// that walks a tree, because the runner walks what is there and a record that
// is gone is not a record it can meet. So the rule held against editing and not
// against removing, and removing is the way to make an experiment stop having
// existed.
//
// WHY IT IS HERE AND NOT IN THE RUNNER, which is record.go's reason one step
// further on. Telling a removal from a record that was never there needs the
// version at the base of the range, which is history rather than a tree, and
// giving the runner a history reader would cost it the dependency surface
// record 0001 chose and the claim that it opens no connection leans on. This
// check already holds both ends of the range, so the comparison belongs beside
// the answer rule it completes.

import "fmt"

// RecordAlreadyLandedWasRemoved refuses a change that removes a record which
// was already on the branch the change lands on.
//
// The failure is ordinary rather than malicious, and that is what makes it
// likely. An experiment answered no, its prototype was removed under record
// 0004, and what is left looks like an empty directory somebody forgot to tidy.
// Tidying it takes the record with it, the tree is green afterwards, and the
// only thing that said the work happened at all is gone.
//
// What stays allowed is exactly what record 0004 already permits: the code
// goes, the record stays, and it gains the line naming the commit that removed
// the code.
const RecordAlreadyLandedWasRemoved = "record-already-landed-was-removed"

// ExperimentAlreadyLandedWasRenamed refuses a change that moves a record which
// was already on the branch the change lands on.
//
// A slug renamed to something better months later is a removal and an addition
// as far as the tree is concerned. Every pointer at the old name stops
// resolving, including a promotion section a reader is following from another
// board, and nothing about the tree afterwards says the old name was ever used.
//
// It is refused rather than repaired, and the refusal names both paths, because
// the repair is a choice this check may not make: keep the slug, or write a new
// experiment with its own question and let the old record say where the work
// went. Renaming in silence is the only option this removes.
const ExperimentAlreadyLandedWasRenamed = "experiment-already-landed-was-renamed"

// judgeRemovals holds every record that was on the branch at the base of the
// range to still being there at the head.
//
// THE BOUNDARIES, WRITTEN HERE RATHER THAN DISCOVERED.
//
// A record created and removed inside one branch is not covered, and it is not
// covered twice over. Such a record never reached the branch this change lands
// on, so nothing about it was made permanent, and a directory added by mistake
// and taken out again in the same pull request is ordinary work. The range is
// read from base to head, so the file is not in the diff at all, and even where
// something put it there the record would carry no version at the base.
//
// A rename is separated from a removal by what git reported, which is a
// similarity judgement rather than a fact about the change. Where the move is
// not reported as one, and a rename made together with a rewrite of the record
// is the case where it will not be, the refusal is the removal rather than the
// rename. That is red for the right reason and names the wrong repair, which is
// the residual worth knowing about before somebody meets it.
//
// WHAT THIS CANNOT DO is stop history being rewritten on the branch itself. A
// force push that removes the commit the record arrived in is refused by the
// ruleset, which refuses a non-fast-forward push and carries no bypass actors,
// and it is named here so that a green run is not read as covering it.
func judgeRemovals(change Change) Verdict {
	if !change.RecordsRead {
		return Verdict{Skips: []Skip{
			{
				Rule: RecordAlreadyLandedWasRemoved,
				Why:  "this run was given no records, so nothing was read at the base of the range to be missing at the head",
			},
			{
				Rule: ExperimentAlreadyLandedWasRenamed,
				Why:  "this run was given no records, so no record was looked for at the two ends of a move",
			},
		}}
	}

	var verdict Verdict
	if !change.FilesRead {
		verdict.Skips = append(verdict.Skips, Skip{
			Rule: ExperimentAlreadyLandedWasRenamed,
			Why:  "this run was given no changed paths, so a record that moved cannot be told from one that went, and a move is refused as a removal",
		})
	}

	moves := renames(change)
	for _, record := range change.Records {
		if !record.BeforePresent || record.AfterPresent {
			continue
		}
		if to, moved := moves[record.Path]; moved {
			verdict.Refusals = append(verdict.Refusals, Refusal{
				Property: ExperimentAlreadyLandedWasRenamed,
				Subject:  record.Path,
				Detail:   fmt.Sprintf("it was on the branch this lands on and this change moves it to %s, so every pointer at the old path stops resolving. Keep the slug, or leave the record where it is and let it say where the work went", to),
			})
			continue
		}
		verdict.Refusals = append(verdict.Refusals, Refusal{
			Property: RecordAlreadyLandedWasRemoved,
			Subject:  record.Path,
			Detail:   "it was on the branch this lands on and is not at the head of this range, so an experiment that happened stops having happened. Record 0004 removes the code and keeps the record, which gains the line naming the commit that removed it",
		})
	}
	return verdict
}

// renames returns where each moved path went, read out of what git reported
// rather than guessed from the content at the two ends.
//
// Guessing was the alternative and it is worse in both directions. Comparing
// the bytes of a record that went with the bytes of one that arrived calls a
// move made together with an edit a removal, and calls two unrelated records
// with the same words a move.
func renames(change Change) map[string]string {
	moves := make(map[string]string)
	for _, file := range change.Files {
		if file.From != "" {
			moves[file.From] = file.Path
		}
	}
	return moves
}
