// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package packageloader defines functions and types for loading and parsing source from disk or VCS.
package packageloader

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilerutil"
	"github.com/serulian/compiler/vcs"

	cmap "github.com/streamrail/concurrent-map"
)

// SerulianPackageDirectory is the directory under the root directory holding cached packages.
const SerulianPackageDirectory = ".pkg"

// SerulianTestSuffix is the suffix for all testing modules. Testing modules will not be loaded
// when loading a package.
const SerulianTestSuffix = "_test"

// PackageLoader helps to fully and recursively load a Serulian package and its dependencies
// from a directory or set of directories.
type PackageLoader struct {
	entrypoint Entrypoint         // The entrypoint for the package loader.
	libraries  map[string]Library // The libraries being loaded.

	vcsDevelopmentDirectories []string   // Directories to check for VCS packages before VCS checkout.
	pathLoader                PathLoader // The path loaders to use.
	alwaysValidate            bool       // Whether to always run validation, regardless of errors. Useful to IDE tooling.
	skipVCSRefresh            bool       // Whether to skip VCS refresh if cache exists. Useful to IDE tooling.

	errors   chan compilercommon.SourceError   // Errors are reported on this channel
	warnings chan compilercommon.SourceWarning // Warnings are reported on this channel

	handlers map[string]SourceHandler       // The handlers for each of the supported package kinds.
	parsers  map[string]SourceHandlerParser // The parsers for each of the supported package kinds.

	pathKindsEncountered cmap.ConcurrentMap    // The path+kinds processed by the loader goroutine
	vcsPathsLoaded       cmap.ConcurrentMap    // The VCS paths that have been loaded, mapping to their checkout dir
	vcsLockMap           compilerutil.LockMap  // LockMap for ensuring single loads of all VCS paths.
	packageMap           *mutablePackageMap    // The package map.
	sourceTracker        *mutableSourceTracker // The source tracker.

	workTracker sync.WaitGroup // WaitGroup used to wait until all loading is complete
	finished    chan bool      // Channel used to tell background goroutines to quit

	cancelationHandle compilerutil.CancelationHandle
}

// Library contains a reference to an external library to load, in addition to those referenced
// by the root source file.
type Library struct {
	PathOrURL string // The file location or SCM URL of the library's package.
	IsSCM     bool   // If true, the PathOrURL is treated as a remote SCM package.
	Kind      string // The kind of the library. Leave empty for Serulian files.
	Alias     string // The import alias for this library.
}

// LoadResult contains the result of attempting to load all packages and source files for this
// project.
type LoadResult struct {
	Status        bool                           // True on success, false otherwise
	Errors        []compilercommon.SourceError   // The errors encountered, if any
	Warnings      []compilercommon.SourceWarning // The warnings encountered, if any
	PackageMap    LoadedPackageMap               // Map of packages loaded.
	SourceTracker SourceTracker                  // Tracker of all source loaded.
}

// NewPackageLoader creates and returns a new package loader for the given config.
func NewPackageLoader(config Config) *PackageLoader {
	handlersMap := map[string]SourceHandler{}
	for _, handler := range config.SourceHandlers {
		handlersMap[handler.Kind()] = handler
	}

	pathLoader := config.PathLoader
	if pathLoader == nil {
		pathLoader = LocalFilePathLoader{}
	}

	return &PackageLoader{
		libraries: map[string]Library{},

		entrypoint:                config.Entrypoint,
		vcsDevelopmentDirectories: config.VCSDevelopmentDirectories,
		pathLoader:                pathLoader,
		alwaysValidate:            config.AlwaysValidate,
		skipVCSRefresh:            config.SkipVCSRefresh,

		errors:   make(chan compilercommon.SourceError, 32),
		warnings: make(chan compilercommon.SourceWarning, 32),

		handlers: handlersMap,
		parsers:  nil,

		pathKindsEncountered: cmap.New(),
		packageMap:           newMutablePackageMap(),

		sourceTracker: newMutableSourceTracker(config.PathLoader),

		vcsPathsLoaded: cmap.New(),
		vcsLockMap:     compilerutil.CreateLockMap(),

		finished: make(chan bool, 1),

		cancelationHandle: compilerutil.GetCancelationHandle(config.cancelationHandle),
	}
}

// Load performs the loading of a Serulian package found at the directory path.
// Any libraries specified will be loaded as well.
func (p *PackageLoader) Load(libraries ...Library) LoadResult {
	// Start the parsers for each of the handlers.
	parsersMap := map[string]SourceHandlerParser{}
	for _, handler := range p.handlers {
		parser := handler.NewParser()
		if parser == nil {
			panic(fmt.Sprintf("Got a nil parser from handler `%s`", handler.Kind()))
		}
		parsersMap[handler.Kind()] = parser
	}
	p.parsers = parsersMap

	// Populate the libraries map.
	for _, library := range libraries {
		p.libraries[library.Alias] = library
	}

	// Start the error/warning collection goroutine.
	result := &LoadResult{
		Status:   true,
		Errors:   make([]compilercommon.SourceError, 0),
		Warnings: make([]compilercommon.SourceWarning, 0),
	}

	go p.collectIssues(result)

	// Add the root source file(s) as the first items to be parsed.
	entrypointPaths, err := p.entrypoint.EntrypointPaths(p.pathLoader)
	if err != nil {
		sourceRange := compilercommon.InputSource(string(p.entrypoint)).RangeForRunePosition(0, p.sourceTracker)
		result.Status = false
		result.Errors = append(result.Errors, compilercommon.SourceErrorf(sourceRange, "Could not resolve entrypoint path: %v", err))
		return *result
	}

	for _, path := range entrypointPaths {
		sourceRange := compilercommon.InputSource(path).RangeForRunePosition(0, p.sourceTracker)
		for _, handler := range p.handlers {
			if strings.HasSuffix(path, handler.PackageFileExtension()) {
				p.pushPath(pathSourceFile, handler.Kind(), path, sourceRange)
				break
			}
		}
	}

	// Add the libraries to be parsed.
	for _, library := range libraries {
		sourceRange := compilercommon.InputSource(library.PathOrURL).RangeForRunePosition(0, p.sourceTracker)
		p.pushLibrary(library, library.Kind, sourceRange)
	}

	// Wait for all packages and source files to be completed.
	p.workTracker.Wait()

	// Tell the goroutines to quit.
	p.finished <- true

	// If canceled, return immediately.
	if p.cancelationHandle.WasCanceled() {
		for _, parser := range p.parsers {
			parser.Cancel()
		}

		return LoadResult{
			Status:   false,
			Errors:   make([]compilercommon.SourceError, 0),
			Warnings: make([]compilercommon.SourceWarning, 0),
		}
	}

	// Save the package map.
	result.PackageMap = p.packageMap.Build()
	result.SourceTracker = p.sourceTracker.Freeze()

	// Apply all parser changes.
	for _, parser := range p.parsers {
		parser.Apply(result.PackageMap, result.SourceTracker, p.cancelationHandle)
	}

	// Perform verification in all parsers.
	if p.alwaysValidate || len(result.Errors) == 0 {
		errorReporter := func(err compilercommon.SourceError) {
			result.Errors = append(result.Errors, err)
			result.Status = false
		}

		warningReporter := func(warning compilercommon.SourceWarning) {
			result.Warnings = append(result.Warnings, warning)
		}

		for _, parser := range p.parsers {
			parser.Verify(errorReporter, warningReporter, p.cancelationHandle)
		}
	}

	if p.cancelationHandle.WasCanceled() {
		return LoadResult{
			Status:   false,
			Errors:   make([]compilercommon.SourceError, 0),
			Warnings: make([]compilercommon.SourceWarning, 0),
		}
	}

	return *result
}

// PathLoader returns the path loader used by this package manager.
func (p *PackageLoader) PathLoader() PathLoader {
	return p.pathLoader
}

// ModuleOrPackage defines a reference to a module or package.
type ModuleOrPackage struct {
	// Name is the name of the module or package.
	Name string

	// Path is the on-disk path of the module or package.
	Path string

	// SourceKind is the kind source for the module or package. Packages will always be
	// empty.
	SourceKind string
}

// ListSubModulesAndPackages lists all modules or packages found *directly* under the given path.
func (p *PackageLoader) ListSubModulesAndPackages(packagePath string) ([]ModuleOrPackage, error) {
	directoryContents, err := p.pathLoader.LoadDirectory(packagePath)
	if err != nil {
		return []ModuleOrPackage{}, err
	}

	var modulesOrPackages = make([]ModuleOrPackage, 0, len(directoryContents))
	for _, entry := range directoryContents {
		// Filter any test modules.
		if strings.Contains(entry.Name, SerulianTestSuffix+".") {
			continue
		}

		if entry.IsDirectory {
			modulesOrPackages = append(modulesOrPackages, ModuleOrPackage{entry.Name, path.Join(packagePath, entry.Name), ""})
			continue
		}

		for _, handler := range p.handlers {
			if strings.HasSuffix(entry.Name, handler.PackageFileExtension()) {
				name := entry.Name[0 : len(entry.Name)-len(handler.PackageFileExtension())]
				modulesOrPackages = append(modulesOrPackages, ModuleOrPackage{name, path.Join(packagePath, entry.Name), handler.Kind()})
				break
			}
		}
	}

	return modulesOrPackages, nil
}

// LocalPackageInfoForPath returns the package information for the given path. Note that VCS paths will
// be converted into their local package equivalent. If the path refers to a source file instead of a
// directory, a package containing the single module will be returned.
func (p *PackageLoader) LocalPackageInfoForPath(path string, sourceKind string, isVCSPath bool) (PackageInfo, error) {
	if isVCSPath {
		localPath, err := p.getVCSDirectoryForPath(path)
		if err != nil {
			return PackageInfo{}, err
		}

		path = localPath
	}

	// Find the source handler matching the source kind.
	handler, ok := p.handlers[sourceKind]
	if !ok {
		return PackageInfo{}, fmt.Errorf("Unknown source kind %s", sourceKind)
	}

	// Check for a single module.
	filePath := path + handler.PackageFileExtension()
	if p.pathLoader.IsSourceFile(filePath) {
		return PackageInfo{
			kind:        sourceKind,
			referenceID: filePath,
			modulePaths: []compilercommon.InputSource{compilercommon.InputSource(filePath)},
		}, nil
	}

	// Otherwise, read the contents of the directory.
	return p.packageInfoForPackageDirectory(path, sourceKind)
}

// packageInfoForDirectory returns a PackageInfo for the package found at the given path.
func (p *PackageLoader) packageInfoForPackageDirectory(packagePath string, sourceKind string) (PackageInfo, error) {
	directoryContents, err := p.pathLoader.LoadDirectory(packagePath)
	if err != nil {
		return PackageInfo{}, err
	}

	handler, ok := p.handlers[sourceKind]
	if !ok {
		return PackageInfo{}, fmt.Errorf("Unknown source kind %s", sourceKind)
	}

	packageInfo := &PackageInfo{
		kind:        sourceKind,
		referenceID: packagePath,
		modulePaths: make([]compilercommon.InputSource, 0),
	}

	// Find all source files in the directory and add them to the paths list.
	for _, entry := range directoryContents {
		// Filter any test modules.
		if strings.Contains(entry.Name, SerulianTestSuffix+".") {
			continue
		}

		if !entry.IsDirectory && path.Ext(entry.Name) == handler.PackageFileExtension() {
			filePath := path.Join(packagePath, entry.Name)

			// Add the source file to the package information.
			packageInfo.modulePaths = append(packageInfo.modulePaths, compilercommon.InputSource(filePath))
		}
	}

	return *packageInfo, nil
}

// getVCSDirectoryForPath returns the directory on disk where the given VCS path will be placed, if any.
func (p *PackageLoader) getVCSDirectoryForPath(vcsPath string) (string, error) {
	pkgDirectory := p.pathLoader.VCSPackageDirectory(p.entrypoint)
	return vcs.GetVCSCheckoutDirectory(vcsPath, pkgDirectory, p.vcsDevelopmentDirectories...)
}

// pushLibrary adds a library to be processed by the package loader.
func (p *PackageLoader) pushLibrary(library Library, kind string, sourceRange compilercommon.SourceRange) string {
	if library.IsSCM {
		return p.pushPath(pathVCSPackage, kind, library.PathOrURL, sourceRange)
	}

	return p.pushPath(pathLocalPackage, kind, library.PathOrURL, sourceRange)
}

// pushPath adds a path to be processed by the package loader.
func (p *PackageLoader) pushPath(kind pathKind, sourceKind string, path string, sourceRange compilercommon.SourceRange) string {
	return p.pushPathWithId(path, sourceKind, kind, path, sourceRange)
}

// pushPathWithId adds a path to be processed by the package loader, with the specified ID.
func (p *PackageLoader) pushPathWithId(pathId string, sourceKind string, kind pathKind, path string, sourceRange compilercommon.SourceRange) string {
	if p.cancelationHandle.WasCanceled() {
		return pathId
	}

	p.workTracker.Add(1)
	go p.loadAndParsePath(pathInformation{pathId, kind, path, sourceKind, sourceRange})
	return pathId
}

// loadAndParsePath parses or loads a specific path.
func (p *PackageLoader) loadAndParsePath(currentPath pathInformation) {
	defer p.workTracker.Done()

	if p.cancelationHandle.WasCanceled() {
		return
	}

	// Ensure we have not already seen this path and kind.
	pathKey := currentPath.String()
	if !p.pathKindsEncountered.SetIfAbsent(pathKey, true) {
		return
	}

	// Perform parsing/loading.
	switch currentPath.kind {
	case pathSourceFile:
		p.conductParsing(currentPath)

	case pathLocalPackage:
		p.loadLocalPackage(currentPath)

	case pathVCSPackage:
		p.loadVCSPackage(currentPath)
	}
}

// loadVCSPackage loads the package found at the given VCS path.
func (p *PackageLoader) loadVCSPackage(packagePath pathInformation) {
	if p.cancelationHandle.WasCanceled() {
		return
	}

	// Lock on the package path to ensure no other checkouts occur for this path.
	pathLock := p.vcsLockMap.GetLock(packagePath.path)
	pathLock.Lock()
	defer pathLock.Unlock()

	existingCheckoutDir, exists := p.vcsPathsLoaded.Get(packagePath.path)
	if exists {
		// Note: existingCheckoutDir will be empty if there was an error loading the VCS.
		if existingCheckoutDir != "" {
			// Push the now-local directory onto the package loading channel.
			p.pushPathWithId(packagePath.referenceID, packagePath.sourceKind, pathLocalPackage, existingCheckoutDir.(string), packagePath.sourceRange)
			return
		}
	}

	// Perform the checkout of the VCS package.
	var cacheOption = vcs.VCSFollowNormalCacheRules
	if p.skipVCSRefresh {
		cacheOption = vcs.VCSAlwaysUseCache
	}

	pkgDirectory := p.pathLoader.VCSPackageDirectory(p.entrypoint)
	result, err := vcs.PerformVCSCheckout(packagePath.path, pkgDirectory, cacheOption, p.vcsDevelopmentDirectories...)
	if err != nil {
		p.vcsPathsLoaded.Set(packagePath.path, "")
		p.enqueueError(compilercommon.SourceErrorf(packagePath.sourceRange, "Error loading VCS package '%s': %v", packagePath.path, err))
		return
	}

	p.vcsPathsLoaded.Set(packagePath.path, result.PackageDirectory)
	if result.Warning != "" {
		p.enqueueWarning(compilercommon.NewSourceWarning(packagePath.sourceRange, result.Warning))
	}

	// Check for VCS version different than a library.
	packageVCSPath, _ := vcs.ParseVCSPath(packagePath.path)
	for _, library := range p.libraries {
		if library.IsSCM && library.Kind == packagePath.sourceKind {
			libraryVCSPath, err := vcs.ParseVCSPath(library.PathOrURL)
			if err != nil {
				continue
			}

			if libraryVCSPath.URL() == packageVCSPath.URL() {
				if libraryVCSPath.String() != packageVCSPath.String() {
					p.enqueueWarning(compilercommon.SourceWarningf(packagePath.sourceRange,
						"Library specifies VCS package `%s` but source file is loading `%s`, which could lead to incompatibilities. It is recommended to upgrade the package in the source file.",
						libraryVCSPath.String(), packageVCSPath.String()))
				}
				break
			}
		}
	}

	// Push the now-local directory onto the package loading channel.
	p.pushPathWithId(packagePath.referenceID, packagePath.sourceKind, pathLocalPackage, result.PackageDirectory, packagePath.sourceRange)
}

// loadLocalPackage loads the package found at the path relative to the package directory.
func (p *PackageLoader) loadLocalPackage(packagePath pathInformation) {
	packageInfo, err := p.packageInfoForPackageDirectory(packagePath.path, packagePath.sourceKind)
	if err != nil {
		p.enqueueError(compilercommon.SourceErrorf(packagePath.sourceRange, "Could not load directory '%s'", packagePath.path))
		return
	}

	// Add the module paths to be parsed.
	var moduleFound = false
	for _, modulePath := range packageInfo.ModulePaths() {
		p.pushPath(pathSourceFile, packagePath.sourceKind, string(modulePath), packagePath.sourceRange)
		moduleFound = true
	}

	// Add the package itself to the package map.
	p.packageMap.Add(packagePath.sourceKind, packagePath.referenceID, packageInfo)
	if !moduleFound {
		p.enqueueWarning(compilercommon.SourceWarningf(packagePath.sourceRange, "Package '%s' has no source files", packagePath.path))
		return
	}
}

// conductParsing performs parsing of a source file found at the given path.
func (p *PackageLoader) conductParsing(sourceFile pathInformation) {
	inputSource := compilercommon.InputSource(sourceFile.path)

	// Add the file to the package map as a package of one file.
	p.packageMap.Add(sourceFile.sourceKind, sourceFile.referenceID, PackageInfo{
		kind:        sourceFile.sourceKind,
		referenceID: sourceFile.referenceID,
		modulePaths: []compilercommon.InputSource{inputSource},
	})

	// Load the source file's contents.
	contents, err := p.pathLoader.LoadSourceFile(sourceFile.path)
	if err != nil {
		p.enqueueError(compilercommon.SourceErrorf(sourceFile.sourceRange, "Could not load source file '%s': %v", sourceFile.path, err))
		return
	}

	// Load the source file's revision ID.
	revisionID, err := p.pathLoader.GetRevisionID(sourceFile.path)
	if err != nil {
		p.enqueueError(compilercommon.SourceErrorf(sourceFile.sourceRange, "Could not load source file '%s': %v", sourceFile.path, err))
		return
	}

	// Add the source file to the tracker.
	p.sourceTracker.AddSourceFile(compilercommon.InputSource(sourceFile.path), sourceFile.sourceKind, contents, revisionID)

	// Parse the source file.
	parser, hasParser := p.parsers[sourceFile.sourceKind]
	if !hasParser {
		log.Fatalf("Missing handler for source file of kind: [%v]", sourceFile.sourceKind)
	}

	parser.Parse(inputSource, string(contents), p.handleImport)
}

// verifyNoVCSBoundaryCross does a check to ensure that walking from the given start path
// to the given end path does not cross a VCS boundary. If it does, an error is returned.
func (p *PackageLoader) verifyNoVCSBoundaryCross(startPath string, endPath string, title string, importInformation PackageImport) *compilercommon.SourceError {
	var checkPath = startPath
	for {
		if checkPath == endPath {
			return nil
		}

		if vcs.IsVCSRootDirectory(checkPath) {
			err := compilercommon.SourceErrorf(importInformation.SourceRange,
				"Import of %s '%s' crosses VCS boundary at package '%s'", title,
				importInformation.Path, checkPath)
			return &err
		}

		nextPath := path.Dir(checkPath)
		if checkPath == nextPath {
			return nil
		}

		checkPath = nextPath
	}
}

// handleImport queues an import found in a source file.
func (p *PackageLoader) handleImport(sourceKind string, importPath string, importType PackageImportType, importSource compilercommon.InputSource, runePosition int) string {
	sourceRange := importSource.RangeForRunePosition(runePosition, p.sourceTracker)
	importInformation := PackageImport{sourceKind, importPath, importType, sourceRange}

	handler, hasHandler := p.handlers[importInformation.Kind]
	if !hasHandler {
		p.enqueueError(compilercommon.SourceErrorf(importInformation.SourceRange, "Unknown kind of import '%s'. Did you forgot to install a source plugin?", importInformation.Kind))
		return ""
	}

	// Check for a library alias.
	switch importInformation.ImportType {
	case ImportTypeAlias:
		// Aliases get pushed as their library.
		libraryName := importInformation.Path
		library, found := p.libraries[libraryName]
		if !found {
			p.enqueueError(compilercommon.SourceErrorf(importInformation.SourceRange, "Import alias `%s` not found", libraryName))
			return ""
		}

		return p.pushLibrary(library, importInformation.Kind, importInformation.SourceRange)

	case ImportTypeVCS:
		// VCS paths get added directly.
		return p.pushPath(pathVCSPackage, importInformation.Kind, importInformation.Path, importInformation.SourceRange)
	}

	// Check the path to see if it exists as a single source file. If so, we add it
	// as a source file instead of a local package.
	sourcePath := string(importInformation.SourceRange.Source())
	currentDirectory := path.Dir(sourcePath)
	dirPath := path.Join(currentDirectory, importInformation.Path)
	filePath := dirPath + handler.PackageFileExtension()

	var importedDirectoryPath = dirPath
	var title = "package"

	// Determine if path refers to a single source file. If so, it is imported rather than
	// the entire directory.
	isSourceFile := p.pathLoader.IsSourceFile(filePath)
	if isSourceFile {
		title = "module"
		importedDirectoryPath = path.Dir(filePath)
	}

	// Check to ensure we are not crossing a VCS boundary.
	if currentDirectory != importedDirectoryPath {
		// If the imported directory is underneath the current directory, we need to walk upward.
		if strings.HasPrefix(importedDirectoryPath, currentDirectory) {
			err := p.verifyNoVCSBoundaryCross(importedDirectoryPath, currentDirectory, title, importInformation)
			if err != nil {
				p.enqueueError(*err)
				return ""
			}
		} else {
			// Otherwise, we walk upward from the current directory to the imported directory.
			err := p.verifyNoVCSBoundaryCross(currentDirectory, importedDirectoryPath, title, importInformation)
			if err != nil {
				p.enqueueError(*err)
				return ""
			}
		}
	}

	// Push the imported path.
	if isSourceFile {
		return p.pushPath(pathSourceFile, handler.Kind(), filePath, importInformation.SourceRange)
	}

	return p.pushPath(pathLocalPackage, handler.Kind(), dirPath, importInformation.SourceRange)
}

// enqueueError queues an error to be added to the Errors slice in the Result.
func (p *PackageLoader) enqueueError(err compilercommon.SourceError) {
	p.workTracker.Add(1)
	p.errors <- err
}

// enqueueWarning queues a warning to be added to the Warnings slice in the Result.
func (p *PackageLoader) enqueueWarning(warning compilercommon.SourceWarning) {
	p.workTracker.Add(1)
	p.warnings <- warning
}

// collectIssues watches the errors and warnings channels to collect those issues as they
// are added.
func (p *PackageLoader) collectIssues(result *LoadResult) {
	for {
		select {
		case newError := <-p.errors:
			result.Errors = append(result.Errors, newError)
			result.Status = false
			p.workTracker.Done()

		case newWarnings := <-p.warnings:
			result.Warnings = append(result.Warnings, newWarnings)
			p.workTracker.Done()

		case <-p.finished:
			return
		}
	}
}
