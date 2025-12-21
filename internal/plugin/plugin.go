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

// Package plugin implements the entrypoint for the stencil-languages
// plugin.
package plugin

import (
	"context"
	"fmt"

	"github.com/google/go-github/v80/github"
	"github.com/jaredallard/vcs"
	"github.com/jaredallard/vcs/token"
	"github.com/rgst-io/stencil-languages/internal/languages/githubactions"
	"github.com/rgst-io/stencil-languages/internal/languages/golang"
	"go.rgst.io/stencil/v2/pkg/extensions/apiv1"
)

// _ ensures that StencilGolangPlugin implements the [apiv1.Implementation] interface.
var _ apiv1.Implementation = &Instance{}

// Instance contains a [apiv1.Implementation] satisfying plugin.
type Instance struct {
	ctx context.Context
	gh  *github.Client
}

// New creates a new [Instance].
func New(ctx context.Context) *Instance {
	// Create an authenticated token, if possible
	var tstr string
	t, err := token.Fetch(ctx, vcs.ProviderGithub, false, &token.Options{AllowUnauthenticated: true})
	if err == nil {
		tstr = t.Value
	}

	gh := github.NewClient(nil)
	if tstr != "" {
		gh = gh.WithAuthToken(tstr)
	}

	return &Instance{ctx: ctx, gh: gh}
}

// GetConfig returns a [apiv1.Config] for the [Instance].
func (*Instance) GetConfig() (*apiv1.Config, error) {
	return &apiv1.Config{}, nil
}

// GetTemplateFunctions returns the [apiv1.TemplateFunction]s for the
// [Instance].
func (*Instance) GetTemplateFunctions() ([]*apiv1.TemplateFunction, error) {
	return []*apiv1.TemplateFunction{
		// GolangMergeGoMod calls [golang.MergeGoMod].
		{
			Name:              "GolangMergeGoMod",
			NumberOfArguments: 2,
		},
		// GithubActionsPinAction calls [githubactions.PinAction].
		{
			Name:              "GithubActionsPinAction",
			NumberOfArguments: 1,
		},
	}, nil
}

// ExecuteTemplateFunction executes a template function for the [Instance].
func (i *Instance) ExecuteTemplateFunction(exec *apiv1.TemplateFunctionExec) (any, error) {
	switch exec.Name { //nolint:gocritic // Why: Will add more cases soon.
	case "GolangMergeGoMod":
		// Safe because of the NumberOfArguments check.
		left, ok := exec.Arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("argument 0 invalid, expected string got %T", exec.Arguments[0])
		}

		right, ok := exec.Arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("argument 1 invalid, expected string got %T", exec.Arguments[1])
		}

		resp, err := golang.MergeGoMod([]byte(left), []byte(right))
		if err != nil {
			return "", err
		}
		return string(resp), nil
	case "GithubActionsPinAction":
		action, ok := exec.Arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("argument 0 invalid, expected string got %T", exec.Arguments[0])
		}

		pa, err := githubactions.PinAction(i.ctx, i.gh, action)
		if err != nil {
			return "", err
		}

		// org/action@<commit> # <tag>
		return pa.Action + "@" + pa.Commit + " # " + pa.Tag, nil
	}

	return nil, fmt.Errorf("unknown template function: %s", exec.Name)
}
