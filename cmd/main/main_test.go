package main

import (
	"net/http"
	"testing"
)

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

	// Integration test requires MySQL to be running
	// Set MYSQL_DSN environment variable before running
	server := newAppServer()
	if server == nil {
		t.Skip("could not create app server, likely missing MYSQL_DSN")
		return
	}
	defer server.Close()

	newAppServerFn = mockNewAppServerFn(server)

	go main()

	// TODO make calls and test
}
