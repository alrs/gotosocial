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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// accountVerifyGETHandler serves a user's account details to them IF they reached this
// handler while in possession of a valid token, according to the oauth middleware.
// It should be served as a GET at /api/v1/accounts/verify_credentials
func (m *accountModule) accountVerifyGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountVerifyGETHandler")
	authed, err := oauth.MustAuth(c, true, false, false, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("retrieved account %+v, converting to mastosensitive...", authed.Account.ID)
	acctSensitive, err := m.db.AccountToMastoSensitive(authed.Account)
	if err != nil {
		l.Tracef("could not convert account into mastosensitive account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("conversion successful, returning OK and mastosensitive account %+v", acctSensitive)
	c.JSON(http.StatusOK, acctSensitive)
}
