// Reading the module table out of what the toolchain recorded.
//
// WHY IT IS HERE RATHER THAN IN THE COMMAND THAT NEEDED IT FIRST. Two documents
// are rendered from one input now, the notices and the bill of materials, and
// both of them are statements about what is inside the same binary. A second
// reader of the module table would let the two disagree about that, which is
// the one failure neither document is allowed to have: an operator holding both
// files would have no way to tell which of them was wrong. So the table is read
// once, here, and both commands call this.

package notices

import "runtime/debug"

// BuildOf turns what the toolchain recorded into what a render reads.
//
// THE REPLACEMENT IS CARRIED RATHER THAN FLATTENED. A replaced module is
// recorded twice by the toolchain: the module the build asked for, carrying the
// replacement, and the replacement itself. What is in the binary is the
// replacement's code, so that is what the licence has to come from, and a
// reader comparing a document against go.mod is looking for the module that was
// asked for. Both are written down.
func BuildOf(info *debug.BuildInfo) Build {
	build := Build{
		Main: Module{Path: info.Main.Path, Version: info.Main.Version},
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			build.Revision = setting.Value
		case "vcs.modified":
			build.Modified = setting.Value == "true"
		}
	}
	for _, dep := range info.Deps {
		if dep == nil {
			continue
		}
		module := Module{Path: dep.Path, Version: dep.Version}
		if dep.Replace != nil {
			module = Module{
				Path:            dep.Replace.Path,
				Version:         dep.Replace.Version,
				ReplacedPath:    dep.Path,
				ReplacedVersion: dep.Version,
			}
		}
		build.Deps = append(build.Deps, module)
	}
	return build
}
