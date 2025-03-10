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
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"golang.org/x/crypto/bcrypt"
)

type login struct {
	Email    string `form:"username"`
	Password string `form:"password"`
}

// signInGETHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The form will then POST to the sign in page, which will be handled by SignInPOSTHandler
func (m *authModule) signInGETHandler(c *gin.Context) {
	m.log.WithField("func", "SignInGETHandler").Trace("serving sign in html")
	c.HTML(http.StatusOK, "sign-in.tmpl", gin.H{})
}

// signInPOSTHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (m *authModule) signInPOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "SignInPOSTHandler")
	s := sessions.Default(c)
	form := &login{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("parsed form: %+v", form)

	userid, err := m.validatePassword(form.Email, form.Password)
	if err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	s.Set("userid", userid)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Trace("redirecting to auth page")
	c.Redirect(http.StatusFound, oauthAuthorizePath)
}

// validatePassword takes an email address and a password.
// The goal is to authenticate the password against the one for that email
// address stored in the database. If OK, we return the userid (a uuid) for that user,
// so that it can be used in further Oauth flows to generate a token/retreieve an oauth client from the db.
func (m *authModule) validatePassword(email string, password string) (userid string, err error) {
	l := m.log.WithField("func", "ValidatePassword")

	// make sure an email/password was provided and bail if not
	if email == "" || password == "" {
		l.Debug("email or password was not provided")
		return incorrectPassword()
	}

	// first we select the user from the database based on email address, bail if no user found for that email
	gtsUser := &model.User{}

	if err := m.db.GetWhere("email", email, gtsUser); err != nil {
		l.Debugf("user %s was not retrievable from db during oauth authorization attempt: %s", email, err)
		return incorrectPassword()
	}

	// make sure a password is actually set and bail if not
	if gtsUser.EncryptedPassword == "" {
		l.Warnf("encrypted password for user %s was empty for some reason", gtsUser.Email)
		return incorrectPassword()
	}

	// compare the provided password with the encrypted one from the db, bail if they don't match
	if err := bcrypt.CompareHashAndPassword([]byte(gtsUser.EncryptedPassword), []byte(password)); err != nil {
		l.Debugf("password hash didn't match for user %s during login attempt: %s", gtsUser.Email, err)
		return incorrectPassword()
	}

	// If we've made it this far the email/password is correct, so we can just return the id of the user.
	userid = gtsUser.ID
	l.Tracef("returning (%s, %s)", userid, err)
	return
}

// incorrectPassword is just a little helper function to use in the ValidatePassword function
func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}
