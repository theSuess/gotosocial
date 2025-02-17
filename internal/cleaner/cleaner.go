// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cleaner

import (
	"context"
	"errors"
	"time"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-sched"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

const (
	selectLimit = 50
)

type Cleaner struct {
	state *state.State
	emoji Emoji
	media Media
}

func New(state *state.State) *Cleaner {
	c := new(Cleaner)
	c.state = state
	c.emoji.Cleaner = c
	c.media.Cleaner = c
	scheduleJobs(c)
	return c
}

// Emoji returns the emoji set of cleaner utilities.
func (c *Cleaner) Emoji() *Emoji {
	return &c.emoji
}

// Media returns the media set of cleaner utilities.
func (c *Cleaner) Media() *Media {
	return &c.media
}

// checkFiles checks for each of the provided files, and calls onMissing() if any of them are missing. Returns true if missing.
func (c *Cleaner) checkFiles(ctx context.Context, onMissing func() error, files ...string) (bool, error) {
	for _, file := range files {
		// Check whether each file exists in storage.
		have, err := c.state.Storage.Has(ctx, file)
		if err != nil {
			return false, gtserror.Newf("error checking storage for %s: %w", file, err)
		} else if !have {
			// Missing files, perform hook.
			return true, onMissing()
		}
	}
	return false, nil
}

// removeFiles removes the provided files, returning the number of them returned.
func (c *Cleaner) removeFiles(ctx context.Context, files ...string) (int, error) {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return len(files), nil
	}

	var errs gtserror.MultiError

	for _, path := range files {
		// Remove each provided storage path.
		log.Debugf(ctx, "removing file: %s", path)
		err := c.state.Storage.Delete(ctx, path)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			errs.Appendf("error removing %s: %v", path, err)
		}
	}

	// Calculate no. files removed.
	diff := len(files) - len(errs)

	// Wrap the combined error slice.
	if err := errs.Combine(); err != nil {
		return diff, gtserror.Newf("error(s) removing files: %w", err)
	}

	return diff, nil
}

func scheduleJobs(c *Cleaner) {
	const day = time.Hour * 24

	// Calculate closest midnight.
	now := time.Now()
	midnight := now.Round(day)

	if midnight.Before(now) {
		// since <= 11:59am rounds down.
		midnight = midnight.Add(day)
	}

	// Get ctx associated with scheduler run state.
	done := c.state.Workers.Scheduler.Done()
	doneCtx := runners.CancelCtx(done)

	// TODO: we'll need to do some thinking to make these
	// jobs restartable if we want to implement reloads in
	// the future that make call to Workers.Stop() -> Workers.Start().

	// Schedule the cleaning tasks to execute every day at midnight.
	c.state.Workers.Scheduler.Schedule(sched.NewJob(func(start time.Time) {
		log.Info(nil, "starting media clean")
		c.Media().All(doneCtx, config.GetMediaRemoteCacheDays())
		c.Emoji().All(doneCtx)
		log.Infof(nil, "finished media clean after %s", time.Since(start))
	}).EveryAt(midnight, day))
}
