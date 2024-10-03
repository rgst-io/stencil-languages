// Copyright (C) 2024 stencil-languages contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0

// Package golang implements Golang specific helpers for stencil
// templates generating Golang code.
package golang

import (
	"fmt"

	"github.com/blang/semver/v4"
	"golang.org/x/mod/modfile"
)

// MergeGoMod merges two go.mod files together. This function operates
// differently than a standard merge would. It is designed to
// conditionally merge an existing go.mod file with a templated go.mod.
//
// ## Behaviour
//
// Noted as the "left" and "right" go.mod files (first and second values
// provided). They are merged with the following rules:
//   - Versions from the right go.mod file will be used if the version
//     is greater than the version in the left go.mod file or the module
//     is not present in the left go.mod file. If a module in the left
//     go.mod is newer than the module in the right go.mod, the left
//     version will be used.
//   - Replacements from the right go.mod file will be kept if they are
//     not in the left go.mod file. If a replacement in the right go.mod
//     has the same path as a replacement in the left go.mod, the left
//     replacement will be kept. Replacements existing in the left go.mod
//     but not in the right go.mod will be kept.
//   - The go and toolchain statements from the right go.mod file will
//     always be used over the left go.mod file.
//
// This is heavily based on getoutreach/stencil-golang, which is
// licensed under the Apache-2.0 license. Link to the original as it
// appeared in the stencil-golang repository:
// https://github.com/getoutreach/stencil-golang/blob/993a3fc484e5631dd9e7004bdd304cbacac7cccd/internal/plugin/merge.go
func MergeGoMod(leftB, rightB []byte) ([]byte, error) {
	leftMod, err := modfile.Parse("go.left.mod", leftB, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse left go.mod: %w", err)
	}

	// Build a map of the left hand module paths to their version.
	leftMods := make(map[string]semver.Version)
	for _, mod := range leftMod.Require {
		v, err := semver.ParseTolerant(mod.Mod.Version)
		if err != nil {
			continue
		}

		leftMods[mod.Mod.Path] = v
	}

	// Build a map of the replaces in the left hand go.mod.
	leftReplaces := make(map[string]*modfile.Replace)
	for _, repl := range leftMod.Replace {
		leftReplaces[repl.Old.Path] = repl
	}

	rightMod, err := modfile.Parse("go.right.mod", rightB, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse right go.mod: %w", err)
	}

	// Change the left hand module versions if the right hand versions
	// are greater than the left hand ones.
	for _, req := range rightMod.Require {
		rv, err := semver.ParseTolerant(req.Mod.Version)
		if err != nil {
			// Invalid, skip. Go would be yelling about this anyways.
			continue
		}

		// Check if it exists in the left hand go.mod.
		if lv, ok := leftMods[req.Mod.Path]; ok {
			// If the right version is less than the left, skip. We don't want
			// to downgrade.
			if rv.LT(lv) {
				continue
			}
		}

		// The right hand version is either greater than left or isn't
		// present in the left go.mod. Add it to the left hand go.mod.
		if err := leftMod.AddRequire(req.Mod.Path, req.Mod.Version); err != nil {
			return nil, fmt.Errorf("failed to add/update dependency '%s': %w", req.Mod.Path, err)
		}
	}

	// Add any modules that exist in the right go.mod, but not in the
	// left.
	for _, repl := range rightMod.Replace {
		// If the left go.mod already has a replacement for this module,
		// don't add a replacement for it from the right.
		if _, ok := leftReplaces[repl.Old.Path]; ok {
			continue
		}

		if err := leftMod.AddReplace(repl.Old.Path, repl.Old.Version, repl.New.Path, repl.New.Version); err != nil {
			return nil, fmt.Errorf("failed to add replacement '%s': %w", repl.Old.Path, err)
		}
	}

	// Always use the go version from the right hand go.mod if present.
	if rightMod.Go != nil && rightMod.Go.Version != "" {
		if err := leftMod.AddGoStmt(rightMod.Go.Version); err != nil {
			return nil, fmt.Errorf("failed to set go version: %w", err)
		}
	}

	// Always use the toolchain from the right hand go.mod if it is set.
	if rightMod.Toolchain != nil && rightMod.Toolchain.Name != "" {
		if err := leftMod.AddToolchainStmt(rightMod.Toolchain.Name); err != nil {
			return nil, fmt.Errorf("failed to set toolchain: %w", err)
		}
	}

	newBytes, err := leftMod.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to format go.mod: %w", err)
	}
	return newBytes, nil
}
