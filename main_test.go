package main

import (
	"net/http"
	"testing"

	"multicarrier-email-api/internal/email/testutils"
)

var databaseTestFacade *testutils.EmailDatabaseFacade

func init() {
	databaseTestFacade = testutils.NewEmailDatabaseFacade()
}

var fixtures []testutils.EmailTestFixtureKeys

func tearDown() {
	_ = databaseTestFacade.RemoveFixtures(fixtures)
	fixtures = []testutils.EmailTestFixtureKeys{}
}

func mockNewAppServerFn(server *http.Server) func() *http.Server {
	return func() *http.Server {
		return server
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode")
		return
	}

	defer tearDown()

	server := newAppServer()
	defer server.Close()

	newAppServerFn = mockNewAppServerFn(server)

	go main()

	// TODO make calls and test
}
