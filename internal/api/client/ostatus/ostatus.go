/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package ostatus

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// AuthorizeInteractionPath is the path the user is redirected to when initiating a follow request via the ostatus subscribe action
	AuthorizeInteractionPath = "/authorize_interaction"
)

// Module implements the ClientAPIModule interface
type Module struct {
	processor processing.Processor
}

// New returns a new user module
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodGet, AuthorizeInteractionPath, m.AuthorizeInteractionGETHandler)
	return nil
}
