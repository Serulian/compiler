// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packageloader

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilerutil"

	"github.com/stretchr/testify/assert"

	cmap "github.com/streamrail/concurrent-map"
)

var _ = fmt.Printf

type testFile struct {
	Imports []string
}

type testTracker struct {
	pathsImported cmap.ConcurrentMap
}

func (tt *testTracker) PackageFileExtension() string {
	return ".json"
}

func (tt *testTracker) Kind() string {
	return ""
}

func (tt *testTracker) NewParser() SourceHandlerParser {
	return &testParser{tt}
}

func (tt *testTracker) createHandler() SourceHandler {
	return tt
}

type testParser struct {
	tt *testTracker
}

func (tt *testParser) Cancel() {
}

func (tt *testParser) Apply(packageMap LoadedPackageMap, sourceTracker SourceTracker, cancelationHandle compilerutil.CancelationHandle) {

}

func (tt *testParser) Verify(errorReporter ErrorReporter, warningReporter WarningReporter, cancelationHandle compilerutil.CancelationHandle) {
}

func (tt *testParser) Parse(source compilercommon.InputSource, input string, importHandler ImportHandler) {
	tt.tt.pathsImported.Set(string(source), true)

	file := testFile{}
	json.Unmarshal([]byte(input), &file)

	for _, importPath := range file.Imports {
		importHandler("", importPath, ImportTypeLocal, source, 0)
	}
}

func TestBasicLoading(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/basic/somefile.json", tt.createHandler()))
	result := loader.Load()
	if !result.Status || len(result.Errors) > 0 {
		t.Errorf("Expected success, found: %v", result.Errors)
	}

	assertFileImported(t, tt, result, "tests/basic/somefile.json")
	assertFileImported(t, tt, result, "tests/basic/anotherfile.json")
	assertFileImported(t, tt, result, "tests/basic/somesubdir/subdirfile.json")

	// Ensure that the PATH map contains an entry for package imported.
	for key := range tt.pathsImported.Items() {
		if _, ok := result.PackageMap.Get("", key); !ok {
			t.Errorf("Expected package %s in packages map", key)
		}
	}
}

func TestRelativeImportSuccess(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/relative/entrypoint.json", tt.createHandler()))
	result := loader.Load()
	if !result.Status || len(result.Errors) > 0 {
		t.Errorf("Expected success, found: %v", result.Errors)
	}

	assertFileImported(t, tt, result, "tests/relative/entrypoint.json")
	assertFileImported(t, tt, result, "tests/relative/subdir/subfile.json")
	assertFileImported(t, tt, result, "tests/relative/relativelyimported.json")

	// Ensure that the PATH map contains an entry for package imported.
	for key := range tt.pathsImported.Items() {
		if _, ok := result.PackageMap.Get("", key); !ok {
			t.Errorf("Expected package %s in packages map", key)
		}
	}
}

func TestRelativeImportFailureAboveVCS(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/vcsabove/fail.json", tt.createHandler()))
	result := loader.Load()
	if !assert.False(t, result.Status, "Expected failure for relative import VCS above") {
		return
	}

	if !assert.Equal(t, 1, len(result.Errors), "Expected one error for relative import VCS above") {
		return
	}

	assert.Equal(t, "Import of package '../basic/foo' crosses VCS boundary at package 'tests/vcsabove'", result.Errors[0].Error(), "Error message mismatch")
}

func TestRelativeImportFailureBelowVCS(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/vcsbelow/fail.json", tt.createHandler()))
	result := loader.Load()
	if !assert.False(t, result.Status, "Expected failure for relative import VCS below") {
		return
	}

	if !assert.Equal(t, 1, len(result.Errors), "Expected one error for relative import VCS below") {
		return
	}

	assert.Equal(t, "Import of package 'somesubdir' crosses VCS boundary at package 'tests/vcsbelow/somesubdir'", result.Errors[0].Error(), "Error message mismatch")
}

func TestUnknownPath(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/unknownimport/importsunknown.json", tt.createHandler()))
	result := loader.Load()
	if result.Status || len(result.Errors) != 1 {
		t.Errorf("Expected error")
		return
	}

	if !strings.Contains(result.Errors[0].Error(), "someunknownpath") {
		t.Errorf("Expected error referencing someunknownpath")
		return
	}

	assertFileImported(t, tt, result, "tests/unknownimport/importsunknown.json")
}

func TestListSubModulesAndPackages(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/basic/somefile.json", tt.createHandler()))

	modulesOrPackages, err := loader.ListSubModulesAndPackages("tests/basic")
	if !assert.Nil(t, err, "Expected successful listing") {
		return
	}

	assert.Equal(t, len(modulesOrPackages), 3)
	assert.Equal(t, modulesOrPackages[0], ModuleOrPackage{"anotherfile", "tests/basic/anotherfile.json", ""})
	assert.Equal(t, modulesOrPackages[1], ModuleOrPackage{"somefile", "tests/basic/somefile.json", ""})
	assert.Equal(t, modulesOrPackages[2], ModuleOrPackage{"somesubdir", "tests/basic/somesubdir", ""})
}

type localPackageInfoForPathTest struct {
	path            string
	sourceKind      string
	isVCSPath       bool
	expectedSuccess bool
	expectedInfo    PackageInfo
}

var localPackageInfoForPathTests = []localPackageInfoForPathTest{
	localPackageInfoForPathTest{"basic", "", false, true, PackageInfo{
		kind:        "",
		referenceID: "tests/basic",
		modulePaths: []compilercommon.InputSource{"tests/basic/anotherfile.json", "tests/basic/somefile.json"},
	}},

	localPackageInfoForPathTest{"basic/anotherfile", "", false, true, PackageInfo{
		kind:        "",
		referenceID: "tests/basic/anotherfile.json",
		modulePaths: []compilercommon.InputSource{"tests/basic/anotherfile.json"},
	}},

	localPackageInfoForPathTest{"relative", "", false, true, PackageInfo{
		kind:        "",
		referenceID: "tests/relative",
		modulePaths: []compilercommon.InputSource{"tests/relative/entrypoint.json", "tests/relative/relativelyimported.json"},
	}},

	localPackageInfoForPathTest{"someinvalid", "", false, false, PackageInfo{}},

	// Note: since we don't have a valid VCS cache, this should fail.
	localPackageInfoForPathTest{"vcsabove", "", true, false, PackageInfo{}},
}

func TestLocalPackageInfoForPath(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/basic/somefile.json", tt.createHandler()))
	for _, test := range localPackageInfoForPathTests {
		packageInfo, err := loader.LocalPackageInfoForPath("tests/"+test.path, test.sourceKind, test.isVCSPath)
		if !assert.Equal(t, err == nil, test.expectedSuccess, "Expected %v success for local package info test %s", test.expectedSuccess, test.path) {
			continue
		}

		if !test.expectedSuccess {
			continue
		}

		if !assert.Equal(t, packageInfo, test.expectedInfo, "Mismatch in expected package info for test %s", test.path) {
			continue
		}
	}
}

func TestLibraryPath(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(NewBasicConfig("tests/basic/somefile.json", tt.createHandler()))
	result := loader.Load(Library{"tests/libtest", false, "", "testlib"})
	if !result.Status || len(result.Errors) > 0 {
		t.Errorf("Expected success, found: %v", result.Errors)
		return
	}

	assertFileImported(t, tt, result, "tests/basic/somefile.json")
	assertFileImported(t, tt, result, "tests/basic/anotherfile.json")
	assertFileImported(t, tt, result, "tests/basic/somesubdir/subdirfile.json")

	assertFileImported(t, tt, result, "tests/libtest/libfile1.json")
	assertFileImported(t, tt, result, "tests/libtest/libfile2.json")
}

func TestEntrypointDir(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(Config{
		Entrypoint:                Entrypoint("tests/libtest/"),
		SourceHandlers:            []SourceHandler{tt.createHandler()},
		VCSDevelopmentDirectories: []string{},
		PathLoader:                LocalFilePathLoader{},
	})

	result := loader.Load()
	if !result.Status || len(result.Errors) > 0 {
		t.Errorf("Expected success, found: %v", result.Errors)
		return
	}

	assertFileImported(t, tt, result, "tests/libtest/libfile1.json")
	assertFileImported(t, tt, result, "tests/libtest/libfile2.json")
}

func assertFileImported(t *testing.T, tt *testTracker, result LoadResult, filePath string) {
	if !tt.pathsImported.Has(filePath) {
		t.Errorf("Expected import of file %s", filePath)
	}

	_, exists := result.SourceTracker.LoadedContents(compilercommon.InputSource(filePath))
	if !exists {
		t.Errorf("Expected tracking of imported file %s", filePath)
	}
}

type TestPathLoader struct{}

func (tpl TestPathLoader) VCSPackageDirectory(entrypoint Entrypoint) string {
	return ""
}

func (tpl TestPathLoader) Exists(path string) (bool, error) {
	_, err := tpl.LoadSourceFile(path)
	return err == nil, nil
}

func (tpl TestPathLoader) LoadSourceFile(path string) ([]byte, error) {
	if path == "startingfile.json" {
		return []byte(`{
				"Imports": ["anotherfile"]
			}
			`), nil
	}

	if path == "anotherfile.json" {
		return []byte("{}"), nil
	}

	return []byte{}, fmt.Errorf("Could not find file: %s", path)
}

func (tpl TestPathLoader) IsSourceFile(path string) bool {
	return path == "startingfile.json" || path == "anotherfile.json"
}

func (tpl TestPathLoader) LoadDirectory(path string) ([]DirectoryEntry, error) {
	return []DirectoryEntry{}, fmt.Errorf("Invalid path: %s", path)
}

func (tpl TestPathLoader) GetRevisionID(path string) (int64, error) {
	return 1, nil
}

func TestLocalLoader(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	loader := NewPackageLoader(Config{
		Entrypoint:                Entrypoint("startingfile.json"),
		SourceHandlers:            []SourceHandler{tt.createHandler()},
		VCSDevelopmentDirectories: []string{},
		PathLoader:                TestPathLoader{},
	})

	result := loader.Load()
	if !result.Status || len(result.Errors) > 0 {
		t.Errorf("Expected success, found: %v", result.Errors)
		return
	}

	assertFileImported(t, tt, result, "startingfile.json")
	assertFileImported(t, tt, result, "anotherfile.json")
}

type BlockingPathLoader struct{}

func (tpl BlockingPathLoader) VCSPackageDirectory(entrypoint Entrypoint) string {
	return ""
}

func (tpl BlockingPathLoader) Exists(path string) (bool, error) {
	_, err := tpl.LoadSourceFile(path)
	return err == nil, nil
}

func (tpl BlockingPathLoader) LoadSourceFile(path string) ([]byte, error) {
	time.Sleep(30 * time.Millisecond)
	if path == "startingfile.json" {
		return []byte(`{
				"Imports": ["anotherfile"]
			}
			`), nil
	}

	return []byte{}, fmt.Errorf("Could not find file: %s", path)
}

func (tpl BlockingPathLoader) IsSourceFile(path string) bool {
	return path == "startingfile.json" || path == "anotherfile.json"
}

func (tpl BlockingPathLoader) LoadDirectory(path string) ([]DirectoryEntry, error) {
	return []DirectoryEntry{}, fmt.Errorf("Invalid path: %s", path)
}

func (tpl BlockingPathLoader) GetRevisionID(path string) (int64, error) {
	return 1, nil
}

func TestCancelation(t *testing.T) {
	tt := &testTracker{
		pathsImported: cmap.New(),
	}

	config := Config{
		Entrypoint:                Entrypoint("startingfile.json"),
		SourceHandlers:            []SourceHandler{tt.createHandler()},
		VCSDevelopmentDirectories: []string{},
		PathLoader:                BlockingPathLoader{},
	}

	cancelable, cancel := config.WithCancel()
	loader := NewPackageLoader(cancelable)

	go func() {
		// Cancel the load.
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()

	result := loader.Load()
	assert.False(t, result.Status, "Expected cancelation")
}
