// Media Rating Overlay - A tool for adding rating information to media posters
// Copyright (C) 2025 Pietro Pollarolo
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package main

import (
	"log"

	"go.uber.org/zap"

	core "github.com/zepollabot/media-rating-overlay/internal/app"
	env "github.com/zepollabot/media-rating-overlay/internal/environment"
)

func main() {
	// Load environment variables
	env.Load()

	// Create and initialize the application
	app, err := core.NewApp()
	if err != nil {
		log.Fatalf("Error initializing application: %v", err)
	}

	defer app.Logger.Sync()

	// Run the application
	if err := app.Run(); err != nil {
		app.Logger.Fatal("Error during execution", zap.Error(err))
	}

	// Shutdown gracefully
	app.Shutdown()
}
