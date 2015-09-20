// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Some code based on the Golang VCS portion of the cmd package:
// https://golang.org/src/cmd/go/vcs.go

//go:generate stringer -type=VCSKind

// vcs package defines helpers functions and interfaces for working with Version Control Systems
// such as git, including discovery of VCS information based on the Golang VCS discovery protocol.
//
// The discovery protocol is documented loosely here: https://golang.org/cmd/go/#hdr-Remote_import_paths
package vcs

import (
	"fmt"
	"log"
	"os"
	"path"
)

// PerformVCSCheckout performs the checkout and updating of the given VCS path and returns
// the local system directory at which the package was checked out.
//
// pkgCacheRootPath holds the path of the root directory that forms the package cache.
func PerformVCSCheckout(vcsPath string, pkgCacheRootPath string) (string, error, string) {
	// Parse the VCS path.
	parsedPath, perr := parseVCSPath(vcsPath)
	if perr != nil {
		return "", perr, ""
	}

	var err error
	var warning string

	// Conduct the checkout or pull.
	fullCacheDirectory := path.Join(pkgCacheRootPath, parsedPath.cacheDirectory())
	err, warning = checkCacheAndPull(parsedPath, fullCacheDirectory)

	// Warn if the package is a HEAD checkout.
	if err == nil && warning == "" && parsedPath.isHEAD() {
		warning = fmt.Sprintf("Package '%s' points to HEAD of a branch or commit and will be updated on every build", parsedPath.String())
	}

	// If the parsed path is a subdirectory of the checkout, return a reference to it.
	if err == nil && parsedPath.subpackage != "" {
		subpackageCacheDirectory := path.Join(fullCacheDirectory, parsedPath.subpackage)
		if _, serr := os.Stat(subpackageCacheDirectory); os.IsNotExist(serr) {
			return "", fmt.Errorf("Subpackage '%s' does not exist under VCS package '%s'", parsedPath.subpackage, parsedPath.url), warning
		}
		return subpackageCacheDirectory, nil, warning
	}

	return fullCacheDirectory, err, warning
}

// checkCacheAndPull conducts the cache check and necessary pulls.
func checkCacheAndPull(parsedPath vcsPackagePath, fullCacheDirectory string) (error, string) {
	// TODO(jschorr): Should we delete the package cache directory here if there was an error?

	// Check the package cache for the path.
	log.Printf("Checking cache directory %s", fullCacheDirectory)
	if _, err := os.Stat(fullCacheDirectory); os.IsNotExist(err) {
		// Do a full checkout.
		return performFullCheckout(parsedPath, fullCacheDirectory)
	}

	// If the cache exists, we only perform an update to the VCS package if the package
	// is marked as pointing to HEAD of a branch or commit. Tagged VCS packages are always left alone.
	log.Printf("Cache directory %s exists", fullCacheDirectory)
	if parsedPath.isHEAD() {
		return performUpdateCheckout(parsedPath, fullCacheDirectory)
	}

	log.Printf("Cache directory %s exists and points to tag %s; no update needed", fullCacheDirectory, parsedPath.tag)
	return nil, ""
}

// performFullCheckout performs a full VCS checkout of the given package path.
func performFullCheckout(path vcsPackagePath, fullCacheDirectory string) (error, string) {
	// Lookup the VCS discovery information.
	discovery, err := DiscoverVCSInformation(path.url)
	if err != nil {
		return err, ""
	}

	// Perform a full checkout.
	handler, ok := vcsByKind[discovery.Kind]
	if !ok {
		panic("Unknown VCS handler")
	}

	warning, err := handler.checkout(path, discovery, fullCacheDirectory)
	return err, warning
}

// performUpdateCheckout performs a VCS update of the given package path.
func performUpdateCheckout(path vcsPackagePath, fullCacheDirectory string) (error, string) {
	// Detect the kind of VCS based on the checkout.
	handler, ok := detectHandler(fullCacheDirectory)
	if !ok {
		return fmt.Errorf("Could not detect VCS for directory: %s", fullCacheDirectory), ""
	}

	// If the checkout has changes, warn, but nothing more to do.
	if handler.check(fullCacheDirectory) {
		warning := fmt.Sprintf("VCS Package '%s' has changes on the local file system and will therefore not be updated", path.String())
		return nil, warning
	}

	// Otherwise, perform a pull to update.
	err := handler.update(fullCacheDirectory)
	return err, ""
}