// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package tester implements support for testing Serulian code.
package tester

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/serulian/compiler/builder"
	"github.com/serulian/compiler/bundle"
	"github.com/serulian/compiler/compilerutil"
	"github.com/serulian/compiler/graphs/scopegraph"
	"github.com/serulian/compiler/packageloader"
	"github.com/serulian/compiler/sourceshape"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// testingRootPath is the path of the testing directory, in which all test runner dependencies will be installed.
const testingRootPath = ".test"

// runners defines the map of test runners by name.
var runners = map[string]TestRunner{}

// TestRunner defines an interface for the test runner.
type TestRunner interface {
	// Title is a human-readable title for the test runner.
	Title() string

	// DecorateCommand decorates the cobra command for the runner with the runner-specific
	// options.
	DecorateCommand(command *cobra.Command)

	// SetupIfNecessary is run before any test runs occur to run the setup process
	// for the runner (if necessary). This method should no-op if all necessary
	// dependencies are in place.
	SetupIfNecessary(testingEnvDirectoryPath string) error

	// Run runs the test runner over the generated ES path.
	Run(testingEnvDirectoryPath string, generatedFilePath string) (bool, error)
}

// runTestsViaRunner runs all the tests at the given source path via the runner.
func runTestsViaRunner(runner TestRunner, path string, vcsDevelopmentDirectories []string) bool {
	log.Printf("Starting test run of %s via %v runner", path, runner.Title())

	// Ensure the testing root path exists.
	if _, serr := os.Stat(testingRootPath); serr != nil && os.IsNotExist(serr) {
		os.Mkdir(testingRootPath, 0777)
	}

	// Run setup for the runner.
	err := runner.SetupIfNecessary(testingRootPath)
	if err != nil {
		errHighlight := color.New(color.FgRed, color.Bold)
		errHighlight.Print("ERROR: ")

		text := color.New(color.FgWhite)
		text.Printf("Could not setup %s runner: %v\n", runner.Title(), err)
		return false
	}

	// Iterate over each test file in the source path. For each, compile the code into
	// JS at a temporary location and then pass the temporary location to the test
	// runner.
	overallSuccess := true
	filesWalked, err := compilerutil.WalkSourcePath(path, func(currentPath string, info os.FileInfo) (bool, error) {
		if !strings.HasSuffix(info.Name(), packageloader.SerulianTestSuffix+sourceshape.SerulianFileExtension) {
			return false, nil
		}

		success := buildAndRunTests(currentPath, vcsDevelopmentDirectories, runner)
		overallSuccess = overallSuccess && success
		return true, nil
	}, packageloader.SerulianPackageDirectory)

	if filesWalked == 0 {
		compilerutil.LogToConsole(compilerutil.WarningLogLevel, nil, "No valid test source files found for path `%s`", path)
		return false
	}

	return overallSuccess && err == nil
}

// buildAndRunTests builds the source found at the given path and then runs its tests via the runner.
func buildAndRunTests(filePath string, vcsDevelopmentDirectories []string, runner TestRunner) bool {
	log.Printf("Building %s...", filePath)

	filename := path.Base(filePath)

	scopeResult, err := scopegraph.ParseAndBuildScopeGraph(filePath,
		vcsDevelopmentDirectories,
		builder.CORE_LIBRARY)

	if err != nil {
		compilerutil.LogToConsole(compilerutil.ErrorLogLevel, nil, "%s", fmt.Errorf("Error running test %s: %v", filePath, err))
		return false
	}

	builder.OutputWarnings(scopeResult.Warnings)

	if !scopeResult.Status {
		builder.OutputErrors(scopeResult.Errors)
		return false
	}

	// Create a temp directory for the outputting bundle.
	dir, err := ioutil.TempDir("", "testing")
	if err != nil {
		log.Fatal(err)
	}

	// Clean up once complete.
	defer os.RemoveAll(dir)

	// Generate the source.
	sourceBundle := builder.GenerateSourceAndBundle(scopeResult)

	// Save the source (with an adjusted call), in a temporary directory.
	moduleName := filename[0 : len(filename)-len(sourceshape.SerulianFileExtension)]
	sourceFilename := moduleName + ".seru.js"

	adjusted := fmt.Sprintf(`%s

		window.Serulian.then(function(global) {
			global.%s.TEST().then(function(a) {
			}).catch(function(err) {
		    throw err;     
		  })
		})

		//# sourceMappingURL=/%s.map
	`, sourceBundle.Source(), moduleName, sourceFilename)

	fullBundle := sourceBundle.BundleWithSource(sourceFilename, "")
	adjustedBundle := bundle.WithFile(fullBundle, bundle.FileFromString(sourceFilename, bundle.Script, adjusted))

	// Write the source and map into the directory.
	err = bundle.WriteToFileSystem(adjustedBundle, dir)
	if err != nil {
		log.Fatal(err)
	}

	// Call the runner with the test file.
	success, err := runner.Run(testingRootPath, path.Join(dir, sourceFilename))
	if err != nil {
		log.Fatal(err)
	}

	return success
}

// DecorateRunners decorates the test command with a command for each runner.
func DecorateRunners(command *cobra.Command, vcsDevelopmentDirectories *[]string) {
	for name, runner := range runners {
		var runnerCmd = &cobra.Command{
			Use:   fmt.Sprintf("%s [source path]", name),
			Short: "Runs the tests defined at the given source path via " + runner.Title(),
			Long:  fmt.Sprintf("Runs the tests found in any *%s.seru files at the given source path", packageloader.SerulianTestSuffix),
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					fmt.Println("Expected source path")
					os.Exit(-1)
				}

				if runTestsViaRunner(runner, args[0], *vcsDevelopmentDirectories) {
					os.Exit(0)
				} else {
					os.Exit(1)
				}
			},
		}

		runner.DecorateCommand(runnerCmd)
		command.AddCommand(runnerCmd)
	}
}

// RegisterRunner registers a test runner with the specific name.
func RegisterRunner(name string, runner TestRunner) {
	if name == "" {
		panic("Test runner must have a name")
	}

	if runner == nil {
		panic("Cannot register nil runner")
	}

	if _, exists := runners[name]; exists {
		panic(fmt.Sprintf("Test runner with name %s already exists", name))
	}

	runners[name] = runner
}
