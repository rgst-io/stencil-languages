// Copyright (C) 2025 stencil-languages contributors
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

// Package githubactions contains language specific helpers for working
// with Github Action workflows.
package githubactions

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-github/v80/github"
)

// PinnedAction represents a pinned action.
type PinnedAction struct {
	// Action is the action pinned minus the version.
	Action string

	// Tag is the tag the tag requested.
	Tag string

	// Commit is the commit SHA [PinnedAction.Tag] pointed to at the time
	// of pinning.
	Commit string
}

// Action represents a Github Action (e.g., jdx/mise-action@v3) after it
// has been parsed.
type Action struct {
	// Original is the original string for this action _without_ the
	// version.
	Original string

	// Host is the Github host. Defaults to github.com
	Host string

	// Org is the Github org
	Org string

	// Repo is the Github repository name
	Repo string

	// Tag is the tag of the action
	Tag string
}

// parsePath mutates the provided [Action] based on the provided Github
// Action. "path" should be a normalized Github Action path, meaning it
// follows the following format: org/action@v0
func parsePath(a *Action, actPath string) error {
	spl := strings.Split(actPath, "/")
	if len(spl) == 1 {
		return fmt.Errorf("invalid action string %q", actPath)
	}

	a.Org = spl[0]
	a.Repo = spl[1]

	return nil
}

// MarshalAction marshals an action string into an [Action].
func MarshalAction(action string) (*Action, error) {
	a := &Action{}

	vSpl := strings.Split(action, "@")
	if len(vSpl) == 1 {
		return nil, fmt.Errorf("invalid action string (missing @) %q", action)
	}

	a.Tag = vSpl[len(vSpl)-1]
	action = strings.TrimSuffix(vSpl[0], "@"+a.Tag)

	a.Original = action

	// Support Github enterprise, or potentially forgejo some day.
	u, err := url.Parse(action)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action as URL: %w", err)
	}

	if u.Scheme != "" && u.Host != "" {
		a.Host = u.Scheme + "://" + u.Host
	}

	if err := parsePath(a, strings.TrimPrefix(u.Path, "/")); err != nil {
		return nil, fmt.Errorf("failed to parse action: %w", err)
	}

	return a, nil
}

// PinAction converts an action string into a [PinnedAction].
func PinAction(ctx context.Context, gh *github.Client, action string) (*PinnedAction, error) {
	a, err := MarshalAction(action)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action: %w", err)
	}

	if a.Host != "" {
		var err error
		gh, err = gh.WithEnterpriseURLs(a.Host, a.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to create github client scoped to %q for action %q", a.Host, action)
		}
	}

	// Attempt to resolve a.Tag as a tag or a branch
	refNames := []string{path.Join("tags", a.Tag), path.Join("heads", a.Tag)}
	var ref *github.Reference
	for _, refName := range refNames {
		_ref, resp, err := gh.Git.GetRef(ctx, a.Org, a.Repo, refName)
		if err != nil {
			if resp.StatusCode != 404 {
				continue
			}

			return nil, fmt.Errorf("failed to resolve ref %q: %w", refName, err)
		}

		ref = _ref
		break
	}
	if ref == nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", a.Tag, err)
	}

	return &PinnedAction{
		Action: a.Original,
		Tag:    a.Tag,
		Commit: ref.GetObject().GetSHA(),
	}, nil
}
