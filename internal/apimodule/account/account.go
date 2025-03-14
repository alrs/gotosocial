/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package account

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	idKey                 = "id"
	basePath              = "/api/v1/accounts"
	basePathWithID        = basePath + "/:" + idKey
	verifyPath            = basePath + "/verify_credentials"
	updateCredentialsPath = basePath + "/update_credentials"
)

type accountModule struct {
	config       *config.Config
	db           db.DB
	oauthServer  oauth.Server
	mediaHandler media.MediaHandler
	log          *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, oauthServer oauth.Server, mediaHandler media.MediaHandler, log *logrus.Logger) apimodule.ClientAPIModule {
	return &accountModule{
		config:       config,
		db:           db,
		oauthServer:  oauthServer,
		mediaHandler: mediaHandler,
		log:          log,
	}
}

// Route attaches all routes from this module to the given router
func (m *accountModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, basePath, m.accountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, basePathWithID, m.muxHandler)
	return nil
}

func (m *accountModule) CreateTables(db db.DB) error {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.FollowRequest{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
		&model.MediaAttachment{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}

func (m *accountModule) muxHandler(c *gin.Context) {
	ru := c.Request.RequestURI
	if strings.HasPrefix(ru, verifyPath) {
		m.accountVerifyGETHandler(c)
	} else if strings.HasPrefix(ru, updateCredentialsPath) {
		m.accountUpdateCredentialsPATCHHandler(c)
	} else {
		m.accountGETHandler(c)
	}
}
