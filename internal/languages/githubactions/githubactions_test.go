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

package githubactions

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v80/github"
)

func TestMarshalAction(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		action  string
		want    *Action
		wantErr bool
	}{
		{
			name:   "should parse standard action path",
			action: "jdx/mise-action@v3",
			want:   &Action{Original: "jdx/mise-action", Org: "jdx", Repo: "mise-action", Tag: "v3"},
		},
		{
			name:   "should parse host action path",
			action: "https://github.com/jdx/mise-action@v3",
			want:   &Action{Original: "https://github.com/jdx/mise-action", Host: "https://github.com", Org: "jdx", Repo: "mise-action", Tag: "v3"},
		},
		{
			name:   "should parse nested action path",
			action: "anchore/sbom-action/download-syft@v0",
			want:   &Action{Original: "anchore/sbom-action/download-syft", Org: "anchore", Repo: "sbom-action", Tag: "v0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := MarshalAction(tt.action)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshalAction() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshalAction() succeeded unexpectedly")
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error("MarshalAction() =", diff)
			}
		})
	}
}

func TestPinAction(t *testing.T) {
	gh := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	tests := []struct {
		name    string // description of this test case
		action  string
		want    *PinnedAction
		wantErr bool
	}{
		{
			name:   "should pin an action",
			action: "jdx/mise-action@v3.5.1",
			want: &PinnedAction{
				Action: "jdx/mise-action",
				Tag:    "v3.5.1",
				Commit: "146a28175021df8ca24f8ee1828cc2a60f980bd5",
			},
		},
		{
			name:    "should fail on invalid action",
			action:  "not-a-real-action/but-here-we-go-anyways@v3.1.1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := PinAction(context.Background(), gh, tt.action)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("PinAction() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("PinAction() succeeded unexpectedly")
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error("PinAction() =", diff)
			}
		})
	}
}
