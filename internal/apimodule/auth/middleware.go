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

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// oauthTokenMiddleware checks if the client has presented a valid oauth Bearer token.
// If so, it will check the User that the token belongs to, and set that in the context of
// the request. Then, it will look up the account for that user, and set that in the request too.
// If user or account can't be found, then the handler won't *fail*, in case the server wants to allow
// public requests that don't have a Bearer token set (eg., for public instance information and so on).
func (m *authModule) oauthTokenMiddleware(c *gin.Context) {
	l := m.log.WithField("func", "ValidatePassword")
	l.Trace("entering OauthTokenMiddleware")

	ti, err := m.server.ValidationBearerToken(c.Request)
	if err != nil {
		l.Trace("no valid token presented: continuing with unauthenticated request")
		return
	}
	c.Set(oauth.SessionAuthorizedToken, ti)
	l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedToken, ti)

	// check for user-level token
	if uid := ti.GetUserID(); uid != "" {
		l.Tracef("authenticated user %s with bearer token, scope is %s", uid, ti.GetScope())

		// fetch user's and account for this user id
		user := &model.User{}
		if err := m.db.GetByID(uid, user); err != nil || user == nil {
			l.Warnf("no user found for validated uid %s", uid)
			return
		}
		c.Set(oauth.SessionAuthorizedUser, user)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedUser, user)

		acct := &model.Account{}
		if err := m.db.GetByID(user.AccountID, acct); err != nil || acct == nil {
			l.Warnf("no account found for validated user %s", uid)
			return
		}
		c.Set(oauth.SessionAuthorizedAccount, acct)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedAccount, acct)
	}

	// check for application token
	if cid := ti.GetClientID(); cid != "" {
		l.Tracef("authenticated client %s with bearer token, scope is %s", cid, ti.GetScope())
		app := &model.Application{}
		if err := m.db.GetWhere("client_id", cid, app); err != nil {
			l.Tracef("no app found for client %s", cid)
		}
		c.Set(oauth.SessionAuthorizedApplication, app)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedApplication, app)
	}
}
