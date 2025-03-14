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

package oauth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/errors"
	"github.com/superseriousbusiness/oauth2/v4/manage"
	"github.com/superseriousbusiness/oauth2/v4/server"
)

const (
	SessionAuthorizedToken = "authorized_token"
	// SessionAuthorizedUser is the key set in the gin context for the id of
	// a User who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.User
	SessionAuthorizedUser = "authorized_user"
	// SessionAuthorizedAccount is the key set in the gin context for the Account
	// of a User who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.Account
	SessionAuthorizedAccount = "authorized_account"
	// SessionAuthorizedAccount is the key set in the gin context for the Application
	// of a Client who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.Application
	SessionAuthorizedApplication = "authorized_app"
)

// Server wraps some oauth2 server functions in an interface, exposing only what is needed
type Server interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) error
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
	GenerateUserAccessToken(ti oauth2.TokenInfo, clientSecret string, userID string) (accessToken oauth2.TokenInfo, err error)
}

// s fulfils the Server interface using the underlying oauth2 server
type s struct {
	server *server.Server
	log    *logrus.Logger
}

type Authed struct {
	Token       oauth2.TokenInfo
	Application *model.Application
	User        *model.User
	Account     *model.Account
}

// GetAuthed is a convenience function for returning an Authed struct from a gin context.
// In essence, it tries to extract a token, application, user, and account from the context,
// and then sets them on a struct for convenience.
//
// If any are not present in the context, they will be set to nil on the returned Authed struct.
//
// If *ALL* are not present, then nil and an error will be returned.
//
// If something goes wrong during parsing, then nil and an error will be returned (consider this not authed).
func GetAuthed(c *gin.Context) (*Authed, error) {
	ctx := c.Copy()
	a := &Authed{}
	var i interface{}
	var ok bool

	i, ok = ctx.Get(SessionAuthorizedToken)
	if ok {
		parsed, ok := i.(oauth2.TokenInfo)
		if !ok {
			return nil, errors.New("could not parse token from session context")
		}
		a.Token = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedApplication)
	if ok {
		parsed, ok := i.(*model.Application)
		if !ok {
			return nil, errors.New("could not parse application from session context")
		}
		a.Application = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedUser)
	if ok {
		parsed, ok := i.(*model.User)
		if !ok {
			return nil, errors.New("could not parse user from session context")
		}
		a.User = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedAccount)
	if ok {
		parsed, ok := i.(*model.Account)
		if !ok {
			return nil, errors.New("could not parse account from session context")
		}
		a.Account = parsed
	}

	if a.Token == nil && a.Application == nil && a.User == nil && a.Account == nil {
		return nil, errors.New("not authorized")
	}

	return a, nil
}

// MustAuth is like GetAuthed, but will fail if one of the requirements is not met.
func MustAuth(c *gin.Context, requireToken bool, requireApp bool, requireUser bool, requireAccount bool) (*Authed, error) {
	a, err := GetAuthed(c)
	if err != nil {
		return nil, err
	}
	if requireToken && a.Token == nil {
		return nil, errors.New("token not supplied")
	}
	if requireApp && a.Application == nil {
		return nil, errors.New("application not supplied")
	}
	if requireUser && a.User == nil {
		return nil, errors.New("user not supplied")
	}
	if requireAccount && a.Account == nil {
		return nil, errors.New("account not supplied")
	}
	return a, nil
}

// HandleTokenRequest wraps the oauth2 library's HandleTokenRequest function
func (s *s) HandleTokenRequest(w http.ResponseWriter, r *http.Request) error {
	return s.server.HandleTokenRequest(w, r)
}

// HandleAuthorizeRequest wraps the oauth2 library's HandleAuthorizeRequest function
func (s *s) HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) error {
	return s.server.HandleAuthorizeRequest(w, r)
}

// ValidationBearerToken wraps the oauth2 library's ValidationBearerToken function
func (s *s) ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	return s.server.ValidationBearerToken(r)
}

// GenerateUserAccessToken shortcuts the normal oauth flow to create an user-level
// bearer token *without* requiring that user to log in. This is useful when we
// need to create a token for new users who haven't validated their email or logged in yet.
//
// The ti parameter refers to an existing Application token that was used to make the upstream
// request. This token needs to be validated and exist in database in order to create a new token.
func (s *s) GenerateUserAccessToken(ti oauth2.TokenInfo, clientSecret string, userID string) (oauth2.TokenInfo, error) {

	authToken, err := s.server.Manager.GenerateAuthToken(context.Background(), oauth2.Code, &oauth2.TokenGenerateRequest{
		ClientID:     ti.GetClientID(),
		ClientSecret: clientSecret,
		UserID:       userID,
		RedirectURI:  ti.GetRedirectURI(),
		Scope:        ti.GetScope(),
	})
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %s", err)
	}
	if authToken == nil {
		return nil, errors.New("generated auth token was empty")
	}
	s.log.Tracef("obtained auth token: %+v", authToken)

	accessToken, err := s.server.Manager.GenerateAccessToken(context.Background(), oauth2.AuthorizationCode, &oauth2.TokenGenerateRequest{
		ClientID:     authToken.GetClientID(),
		ClientSecret: clientSecret,
		RedirectURI:  authToken.GetRedirectURI(),
		Scope:        authToken.GetScope(),
		Code:         authToken.GetCode(),
	})

	if err != nil {
		return nil, fmt.Errorf("error generating user-level access token: %s", err)
	}
	if accessToken == nil {
		return nil, errors.New("generated user-level access token was empty")
	}
	s.log.Tracef("obtained user-level access token: %+v", accessToken)
	return accessToken, nil
}

func New(database db.DB, log *logrus.Logger) Server {
	ts := newTokenStore(context.Background(), database, log)
	cs := newClientStore(database)

	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	sc := &server.Config{
		TokenType: "Bearer",
		// Must follow the spec.
		AllowGetAccessRequest: false,
		// Support only the non-implicit flow.
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code},
		// Allow:
		// - Authorization Code (for first & third parties)
		// - Client Credentials (for applications)
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.ClientCredentials,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{oauth2.CodeChallengePlain},
	}

	srv := server.NewServer(sc, manager)
	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		log.Errorf("internal oauth error: %s", err)
		return nil
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Errorf("internal response error: %s", re.Error)
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		userID := r.FormValue("userid")
		if userID == "" {
			return "", errors.New("userid was empty")
		}
		return userID, nil
	})
	srv.SetClientInfoHandler(server.ClientFormHandler)
	return &s{
		server: srv,
		log:    log,
	}
}
