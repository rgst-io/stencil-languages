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

package plugin_test

import (
	"context"
	"testing"

	"github.com/rgst-io/stencil-languages/internal/plugin"
	"go.rgst.io/stencil/v2/pkg/extensions/apiv1"
	"gotest.tools/v3/assert"
)

// TestGolangMergeGoMod ensures that we can call the GolangMergeGoMod
// template function.
func TestGolangMergeGoMod(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// TODO(jaredallard): There's no package exported by stencil for
	// testing the over the wire functionality. Best we can do is mimic
	// stencil by using [apiv1.Implementation].
	var impl apiv1.Implementation = plugin.New(ctx)

	resp, err := impl.ExecuteTemplateFunction(&apiv1.TemplateFunctionExec{
		Name:      "GolangMergeGoMod",
		Arguments: []any{"go 1.16", "go 1.17"},
	})
	assert.NilError(t, err, "failed to execute template function")
	assert.Equal(t, resp, "go 1.17\n", "expected go1.17")
}

// TestGithubActionsPinAction ensures that we can call the
// GithubActionsPinAction template function.
func TestGithubActionsPinAction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var impl apiv1.Implementation = plugin.New(ctx)

	resp, err := impl.ExecuteTemplateFunction(&apiv1.TemplateFunctionExec{
		Name:      "GithubActionsPinAction",
		Arguments: []any{"jdx/mise-action@v3.5.1"},
	})
	assert.NilError(t, err, "failed to execute template function")
	assert.Equal(t, resp, "jdx/mise-action@146a28175021df8ca24f8ee1828cc2a60f980bd5 # v3.5.1")
}
