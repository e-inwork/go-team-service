// Copyright 2023, e-inwork.com. All rights reserved.

package api

import (
	"io"
	"net/http"
	"testing"

	"github.com/e-inwork-com/go-team-service/internal/data/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	app := testApplication(t)

	ts := testServer(t, app.Routes())
	defer ts.Close()

	firstToken := app.testFirstToken(t)
	secondToken := app.testSecondToken(t)
	tBodyTeam, tContentTypeTeam := app.testFormTeam(t)
	tJSONTeamMember := app.testJSONTeamMember(t)

	tests := []struct {
		name         string
		method       string
		urlPath      string
		contentType  string
		token        string
		body         io.Reader
		expectedCode int
	}{
		{
			name:         "Create Team",
			method:       "POST",
			urlPath:      "/service/teams",
			contentType:  tContentTypeTeam,
			token:        firstToken,
			body:         tBodyTeam,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Get Team",
			method:       "GET",
			urlPath:      "/service/teams/me",
			contentType:  "",
			token:        firstToken,
			body:         nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Team Picture",
			method:       "GET",
			urlPath:      "/service/teams/pictures/" + mocks.MockFirstUUID().String() + ".jpg",
			contentType:  "",
			token:        "",
			body:         nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Patch Team",
			method:       "PATCH",
			urlPath:      "/service/teams/" + mocks.MockFirstUUID().String(),
			contentType:  tContentTypeTeam,
			token:        firstToken,
			body:         tBodyTeam,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Patch Team Forbidden",
			method:       "PATCH",
			urlPath:      "/service/teams/" + mocks.MockFirstUUID().String(),
			contentType:  tContentTypeTeam,
			token:        secondToken,
			body:         tBodyTeam,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Create Team Member",
			method:       "POST",
			urlPath:      "/service/teams/members",
			contentType:  "application/json",
			token:        firstToken,
			body:         tJSONTeamMember,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Get Team Member",
			method:       "GET",
			urlPath:      "/service/teams/members/" + mocks.MockFirstUUID().String(),
			contentType:  "",
			token:        firstToken,
			body:         nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get List Team Members",
			method:       "GET",
			urlPath:      "/service/teams/members",
			contentType:  "",
			token:        firstToken,
			body:         nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Delete Team Members",
			method:       "DELETE",
			urlPath:      "/service/teams/members/" + mocks.MockFirstUUID().String(),
			contentType:  "",
			token:        firstToken,
			body:         nil,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCode, _, _ := ts.request(t, tt.method, tt.urlPath, tt.contentType, tt.token, tt.body)
			assert.Equal(t, tt.expectedCode, actualCode)
		})
	}
}
