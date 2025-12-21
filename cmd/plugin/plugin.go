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

// Package main wraps the plugin logic so that it can be called by
// stencil.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rgst-io/stencil-languages/internal/plugin"
	"go.rgst.io/stencil/v2/pkg/extensions/apiv1"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// main starts the stencil-languages plugin
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	log := slogext.New()

	err := apiv1.NewExtensionImplementation(plugin.New(ctx), log)
	if err != nil {
		log.WithError(err).Error("failed to create extension")
	}

	// close the context
	cancel()

	if err != nil {
		os.Exit(1)
	}
}
