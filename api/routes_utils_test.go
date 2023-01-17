// Copyright 2023, e-inwork.com. All rights reserved.

package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/e-inwork-com/go-team-service/internal/data/mocks"
	"github.com/e-inwork-com/go-team-service/internal/jsonlog"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func testApplication(t *testing.T) *Application {

	var cfg Config
	cfg.Auth.Secret = "secret"
	cfg.Uploads = "../local/test/uploads"

	return &Application{
		Config: cfg,
		Logger: jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		Models: data.Models{
			Teams:       &mocks.TeamModel{},
			Users:       &mocks.UserModel{},
			TeamMembers: &mocks.TeamMemberModel{},
		},
	}

}

type httpTestServer struct {
	*httptest.Server
}

func testServer(t *testing.T, h http.Handler) *httpTestServer {
	ts := httptest.NewTLSServer(h)

	return &httpTestServer{ts}
}

func (ts *httpTestServer) request(t *testing.T, method string, urlPath string, contentType string, authToken string, body io.Reader) (int, http.Header, string) {
	rq, _ := http.NewRequest(method, ts.URL+urlPath, body)

	if contentType != "" {
		rq.Header.Add("Content-Type", contentType)
	}

	if authToken != "" {
		rq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", authToken))
	}

	rs, err := ts.Client().Do(rq)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	bd, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(bd)

	return rs.StatusCode, rs.Header, string(bd)
}

func (app *Application) testCreateToken(t *testing.T, id uuid.UUID) string {
	// Set Signing Key from the Config Environment
	signingKey := []byte(app.Config.Auth.Secret)

	// Set an expired time for a week
	expirationTime := time.Now().Add((24 * 7) * time.Hour)

	// Set the ID of the user in the Claim token
	claims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create a signed token
	signed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := signed.SignedString(signingKey)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func (app *Application) testFirstToken(t *testing.T) string {
	// Create UUID
	id := mocks.MockFirstUUID()

	return app.testCreateToken(t, id)
}

func (app *Application) testSecondToken(t *testing.T) string {
	// Create UUID
	id := mocks.MockSecondUUID()

	return app.testCreateToken(t, id)
}

func (app *Application) testFormTeam(t *testing.T) (io.Reader, string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// Add team name
	bodyWriter.WriteField("team_name", "Doe's Team")

	// Add team picture
	filename := "./test/images/team.jpg"
	fileWriter, err := bodyWriter.CreateFormFile("team_picture", filename)
	if err != nil {
		t.Fatal(err)
	}

	// Open file
	fileHandler, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	// Copy file
	_, err = io.Copy(fileWriter, fileHandler)
	if err != nil {
		t.Fatal(err)
	}

	// Put on body
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	return bodyBuf, contentType
}

func (app *Application) testJSONTeamMember(t *testing.T) io.Reader {
	teamMember := fmt.Sprintf(
		`{"team_member_team": "%v", "team_member_user":  "%v"}`,
		mocks.MockFirstUUID(),
		mocks.MockSecondUUID())

	return bytes.NewReader([]byte(teamMember))
}
