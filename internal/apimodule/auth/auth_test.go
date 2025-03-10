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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"golang.org/x/crypto/bcrypt"
)

type AuthTestSuite struct {
	suite.Suite
	oauthServer     oauth.Server
	db              db.DB
	testAccount     *model.Account
	testApplication *model.Application
	testUser        *model.User
	testClient      *oauth.Client
	config          *config.Config
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AuthTestSuite) SetupSuite() {
	c := config.Empty()
	// we're running on localhost without https so set the protocol to http
	c.Protocol = "http"
	// just for testing
	c.Host = "localhost:8080"
	// because go tests are run within the test package directory, we need to fiddle with the templateconfig
	// basedir in a way that we wouldn't normally have to do when running the binary, in order to make
	// the templates actually load
	c.TemplateConfig.BaseDir = "../../../web/template/"
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	suite.config = c

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logrus.Panicf("error encrypting user pass: %s", err)
	}

	acctID := uuid.NewString()

	suite.testAccount = &model.Account{
		ID:       acctID,
		Username: "test_user",
	}
	suite.testUser = &model.User{
		EncryptedPassword: string(encryptedPassword),
		Email:             "user@example.org",
		AccountID:         acctID,
	}
	suite.testClient = &oauth.Client{
		ID:     "a-known-client-id",
		Secret: "some-secret",
		Domain: fmt.Sprintf("%s://%s", c.Protocol, c.Host),
	}
	suite.testApplication = &model.Application{
		Name:         "a test application",
		Website:      "https://some-application-website.com",
		RedirectURI:  "http://localhost:8080",
		ClientID:     "a-known-client-id",
		ClientSecret: "some-secret",
		Scopes:       "read",
		VapidKey:     uuid.NewString(),
	}
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *AuthTestSuite) SetupTest() {

	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	db, err := db.New(context.Background(), suite.config, log)
	if err != nil {
		logrus.Panicf("error creating database connection: %s", err)
	}

	suite.db = db

	models := []interface{}{
		&oauth.Client{},
		&oauth.Token{},
		&model.User{},
		&model.Account{},
		&model.Application{},
	}

	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}

	suite.oauthServer = oauth.New(suite.db, log)

	if err := suite.db.Put(suite.testAccount); err != nil {
		logrus.Panicf("could not insert test account into db: %s", err)
	}
	if err := suite.db.Put(suite.testUser); err != nil {
		logrus.Panicf("could not insert test user into db: %s", err)
	}
	if err := suite.db.Put(suite.testClient); err != nil {
		logrus.Panicf("could not insert test client into db: %s", err)
	}
	if err := suite.db.Put(suite.testApplication); err != nil {
		logrus.Panicf("could not insert test application into db: %s", err)
	}

}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *AuthTestSuite) TearDownTest() {
	models := []interface{}{
		&oauth.Client{},
		&oauth.Token{},
		&model.User{},
		&model.Account{},
		&model.Application{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.db = nil
}

func (suite *AuthTestSuite) TestAPIInitialize() {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)

	r, err := router.New(suite.config, log)
	if err != nil {
		suite.FailNow(fmt.Sprintf("error mapping routes onto router: %s", err))
	}

	api := New(suite.oauthServer, suite.db, log)
	if err := api.Route(r); err != nil {
		suite.FailNow(fmt.Sprintf("error mapping routes onto router: %s", err))
	}

	r.Start()
	time.Sleep(60 * time.Second)
	if err := r.Stop(context.Background()); err != nil {
		suite.FailNow(fmt.Sprintf("error stopping router: %s", err))
	}
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
