package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

var testModels []interface{} = []interface{}{
	&model.Account{},
	&model.Application{},
	&model.Block{},
	&model.DomainBlock{},
	&model.EmailDomainBlock{},
	&model.Follow{},
	&model.FollowRequest{},
	&model.MediaAttachment{},
	&model.Mention{},
	&model.Status{},
	&model.Tag{},
	&model.User{},
	&oauth.Token{},
	&oauth.Client{},
}

// StandardDBSetup populates a given db with all the necessary tables/models for perfoming tests.
func StandardDBSetup(db db.DB) error {
	for _, m := range testModels {
		if err := db.CreateTable(m); err != nil {
			return err
		}
	}

	for _, v := range TestTokens() {
		if err := db.Put(v); err != nil {
			return err
		}
	}

	for _, v := range TestClients() {
		if err := db.Put(v); err != nil {
			return err
		}
	}

	for _, v := range TestApplications() {
		if err := db.Put(v); err != nil {
			return err
		}
	}

	for _, v := range TestUsers() {
		if err := db.Put(v); err != nil {
			return err
		}
	}

	for _, v := range TestAccounts() {
		if err := db.Put(v); err != nil {
			return err
		}
	}

	return nil
}

// StandardDBTeardown drops all the standard testing tables/models from the database to ensure it's clean for the next test.
func StandardDBTeardown(db db.DB) error {
	for _, m := range testModels {
		if err := db.DropTable(m); err != nil {
			return err
		}
	}
	return nil
}
