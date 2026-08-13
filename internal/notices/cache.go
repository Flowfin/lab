// The source that reads licence texts out of a module cache.
//
// The cache is where the toolchain already put every module it downloaded, so
// reading it costs no fetch and no network. That matters twice: this repository
// claims its runner opens no connection and holds that claim to a test, and a
// notices file that could only be produced with a working network is a notices
// file that cannot be produced from an archive of the build.

package notices

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// A Cache reads licence texts from a directory laid out the way the module
// cache is: one directory per module and version, under the escaped module
// path.
//
// It is a directory rather than the cache itself so that a case can write one
// out in full. A test resolving against whichever modules the machine running
// the suite happens to have downloaded proves the state of that machine on the
// day it ran, not the reader.
type Cache struct {
	// Root is the directory the modules are under.
	Root string
}

// Licence reads the text a module shipped.
//
// It tries the conventional filenames in order and takes the first that holds
// something. A module carrying two of them is reporting the same licence twice
// in this repository's experience, and a document reproducing both would double
// its length for no reader.
func (c Cache) Licence(module Module) (Licence, error) {
	dir, err := c.moduleDir(module)
	if err != nil {
		return Licence{}, err
	}

	var tried []string
	for _, filename := range LicenceFilenames {
		text, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			tried = append(tried, filename)
			continue
		}
		if strings.TrimSpace(string(text)) == "" {
			return Licence{}, fmt.Errorf("%s is present and holds no text, so there is nothing to reproduce", LicencePath(module, filename))
		}
		return Licence{File: LicencePath(module, filename), Text: string(text)}, nil
	}
	return Licence{}, fmt.Errorf("no licence file under %s; the names tried were %s", ModuleDir(module), strings.Join(tried, ", "))
}

// moduleDir is where the cache holds this module, and it refuses a module path
// that does not name a directory under the root.
//
// THE INPUT IS NOT THIS REPOSITORY'S. A module path arrives from the module
// table inside a binary, which is written by whoever built it, so a path
// carrying a parent-directory element would send this reader outside the cache
// and reproduce whatever it found there as a licence. Joining and cleaning is
// what makes that possible, so the join is checked rather than trusted.
func (c Cache) moduleDir(module Module) (string, error) {
	root, err := filepath.Abs(c.Root)
	if err != nil {
		return "", fmt.Errorf("cannot resolve the cache root %s: %w", c.Root, err)
	}
	dir, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(ModuleDir(module))))
	if err != nil {
		return "", fmt.Errorf("cannot resolve %s under the cache root: %w", ModuleDir(module), err)
	}
	relative, err := filepath.Rel(root, dir)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("the module path %q does not name a directory under the cache root, so nothing there is this module's licence", module.Path)
	}
	return dir, nil
}
