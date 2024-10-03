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

package golang_test

import (
	"strings"
	"testing"

	"github.com/rgst-io/stencil-languages/internal/languages/golang"
	"gotest.tools/v3/assert"
)

func TestMergeGoMod(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{
			name: "should do nothing when both are empty",
			args: []string{"", ""},
		},
		{
			name: "should replace left when empty",
			args: []string{"", "go 1.16"},
			want: "go 1.16",
		},
		{
			name: "should use version from right if newer than left",
			args: []string{"require foo v1.0.0", "require foo v1.1.0"},
			want: "require foo v1.1.0",
		},
		{
			name: "should add version from right if not in left",
			args: []string{"require foo v1.0.0", "require bar v1.1.0"},
			want: strings.Join([]string{
				"require (",
				"	foo v1.0.0",
				"	bar v1.1.0",
				")",
			}, "\n"),
		},
		{
			name: "should keep version from left if newer than right",
			args: []string{"require foo v1.1.0", "require foo v1.0.0"},
			want: "require foo v1.1.0",
		},
		{
			name: "should keep replacements from left if not in right",
			args: []string{"replace foo => bar v1.0.0", ""},
			want: "replace foo => bar v1.0.0",
		},
		{
			name: "should add replacements from right if not in left",
			args: []string{"", "replace foo => bar v1.0.0"},
			want: "replace foo => bar v1.0.0",
		},
		{
			name: "should add go statement from right if present",
			args: []string{"", "go 1.16"},
			want: "go 1.16",
		},
		{
			name: "should replace go statement from left if present",
			args: []string{"go 1.15", "go 1.16"},
			want: "go 1.16",
		},
		{
			name: "should add toolchain statement from right if present",
			args: []string{"", "toolchain go1.23.2"},
			want: "toolchain go1.23.2",
		},
		{
			name: "should replace toolchain statement from left if present",
			args: []string{"toolchain go1.23.1", "toolchain go1.23.2"},
			want: "toolchain go1.23.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Assert(t, len(tt.args) == 2, "expected exactly 2 arguments")

			gotB, err := golang.MergeGoMod([]byte(tt.args[0]), []byte(tt.args[1]))
			got := string(gotB)
			if got != "" && strings.HasSuffix(got, "\n") {
				// Ensure we have no trailing newline for testing.
				got = got[:len(got)-1]
			}
			if tt.wantErr != "" {
				assert.Assert(t, err != nil, "expected an error")
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
