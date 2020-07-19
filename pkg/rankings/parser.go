package rankings

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
)

// KeywordData contains all SERP data for a group of keywords
type KeywordData map[string]SERP

// New creates a new KeywordData instance
func New() KeywordData {
	return make(map[string]SERP)
}

// SERP contains a set of related serps
type SERP struct {
	Keyword string
	Members []SERPMember
}

// Length determines the length of the members in the SERP
func (s SERP) Length() int {
	return len(s.Members)
}

// SERPMember represents a ranked member in prominent entries in a SERP
type SERPMember struct {
	Keyword    string `json:"keyword"`
	Prominence int    `json:"prominence"`
	Domain     string `json:"competitor"`
}

// Parse builds a SERP by parsing from JSON data
func Parse(rdr io.Reader) (SERP, error) {
	var serp SERP
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return serp, fmt.Errorf("could not read input: %v", err)
	}

	var entries []SERPMember
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return serp, fmt.Errorf("could not parse JSON: %v", err)
	}

	if len(entries) == 0 {
		return serp, nil
	}

	serp.Keyword = entries[0].Keyword
	for _, e := range entries {
		serp.Members = append(serp.Members, e)
	}

	return serp, nil
}

// Build builds a KeywordData set by parsing and adding SERP members
func (kd KeywordData) Build(paths []string) error {
	var errors []error
	if len(paths) == 0 {
		return nil
	}

	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(paths))

	for _, path := range paths {
		go func(path string) {
			defer wg.Done()
			f, err := os.Open(path)
			if err != nil {
				errors = append(errors, fmt.Errorf("could not open %s: %v", path, err))
				return
			}
			defer f.Close()
			serp, err := Parse(f)
			if err != nil {
				errors = append(errors, fmt.Errorf("could not parse %s: %v", path, err))
			}

			lock.Lock()
			kd[serp.Keyword] = serp
			lock.Unlock()
		}(path)
	}

	wg.Wait()
	if len(errors) > 0 {
		return BuildError{Errors: errors}
	}

	return nil
}

// BuildError represents one or more errors that could occur as the result
// processing a directory
type BuildError struct {
	Errors []error
}

func (be BuildError) Error() string {
	return fmt.Sprintf("got %d error(s) - first was: %v",
		len(be.Errors),
		be.Errors[0])
}

// ProcessDirectory scans a directory for files containing SERP data and builds
// a KeywordData from their contents
func ProcessDirectory(directory string) (KeywordData, error) {
	dir, err := os.Open(directory)
	if err != nil {
		return nil, fmt.Errorf("could not open directory: %v", err)
	}
	defer dir.Close()

	children, err := dir.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %v", err)
	}

	var paths []string
	for _, p := range children {
		paths = append(paths, path.Join(directory, p.Name()))
	}

	kd := New()
	err = kd.Build(paths)
	if err != nil {
		log.Fatalf("could not build keyword data: %v", err)
	}

	return kd, nil
}
