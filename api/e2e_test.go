package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/e-inwork-com/go-team-service/internal/jsonlog"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	// Team Service
	var cfg Config
	cfg.Db.Dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	cfg.Auth.Secret = "secret"
	cfg.Db.MaxOpenConn = 25
	cfg.Db.MaxIdleConn = 25
	cfg.Db.MaxIdleTime = "15m"
	cfg.Limiter.Enabled = true
	cfg.Limiter.Rps = 2
	cfg.Limiter.Burst = 6
	cfg.GRPCTeam = "localhost:5001"
	cfg.Uploads = "../local/test/uploads"

	// Set logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Set Database
	db, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	// Set Applcation
	app := Application{
		Config: cfg,
		Logger: logger,
		Models: data.InitModels(db),
	}

	// Server Routes API
	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	// Read a SQL file for the deleting all records
	script, err := os.ReadFile("./test/sql/delete_all.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Delete all records
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Delete all indexing in the Solr Team Collection
	res, err := http.Post(
		"http://localhost:8983/solr/teams/update?commit=true",
		"application/json",
		bytes.NewReader([]byte("{'delete': {'query': '*:*'}}")))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatal(res)
	}

	// Register
	email := "jon@doe.com"
	password := "pa55word"
	user := fmt.Sprintf(
		`{"email": "%v", "password":  "%v", "first_name": "Jon", "last_name": "Doe"}`,
		email, password)
	res, err = http.Post(
		"http://localhost:8000/service/users",
		"application/json",
		bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusAccepted)

	// Read response
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	assert.Nil(t, err)

	var mUser map[string]data.User
	err = json.Unmarshal(body, &mUser)
	assert.Nil(t, err)
	assert.Equal(t, mUser["user"].Email, email)

	// Sign in
	login := fmt.Sprintf(
		`{"email": "%v", "password":  "%v"}`,
		email, password)
	res, err = http.Post(
		"http://localhost:8000/service/users/authentication",
		"application/json",
		bytes.NewReader([]byte(login)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Read a token
	body, err = io.ReadAll(res.Body)
	defer res.Body.Close()
	assert.Nil(t, err)

	type authType struct {
		Token string `json:"token"`
	}
	var authResult authType
	err = json.Unmarshal(body, &authResult)
	assert.Nil(t, err)
	assert.NotNil(t, authResult.Token)

	// Create Team
	// Create body buffer
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

	// Post a new team
	req, _ := http.NewRequest("POST", ts.URL+"/service/teams", bodyBuf)
	req.Header.Add("Content-Type", contentType)

	bearer := fmt.Sprintf("Bearer %v", authResult.Token)
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Read response
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	assert.Nil(t, err)

	var mTeam map[string]data.Team
	err = json.Unmarshal(body, &mTeam)
	assert.Nil(t, err)
	assert.Equal(t, mTeam["team"].TeamUser, mUser["user"].ID)

	// Get a picture
	req, _ = http.NewRequest(
		"GET",
		ts.URL+"/service/teams/pictures/"+mTeam["team"].TeamPicture,
		nil)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Read response
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	assert.Nil(t, err)

	// Check type of file
	filetype := http.DetectContentType(body)
	assert.NotEqual(t, filetype, "")

	// Register a user for the team member
	user = `{"email": "nina@doe.com", "password":  "pa55word", "first_name": "Nina", "last_name": "Doe"}`
	res, err = http.Post(
		"http://localhost:8000/service/users",
		"application/json",
		bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusAccepted)

	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	assert.Nil(t, err)

	var mNewUser map[string]data.User
	err = json.Unmarshal(body, &mNewUser)
	assert.Nil(t, err)

	// Create a new team member
	teamMember := fmt.Sprintf(
		`{"team_member_team": "%v", "team_member_user":  "%v"}`,
		mTeam["team"].ID,
		mNewUser["user"].ID)
	req, _ = http.NewRequest(
		"POST",
		ts.URL+"/service/teams/members",
		bytes.NewReader([]byte(teamMember)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Read response
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	assert.Nil(t, err)

	var mTeamMember map[string]data.TeamMember
	err = json.Unmarshal(body, &mTeamMember)
	assert.Nil(t, err)
	assert.Equal(t, mTeamMember["team_member"].TeamMemberTeam, mTeam["team"].ID)

	// Get a list team members of the current user
	req, _ = http.NewRequest("GET", ts.URL+"/service/teams/members", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Read response
	var mTeamMembers map[string][]data.TeamMember
	err = json.NewDecoder(res.Body).Decode(&mTeamMembers)
	assert.Nil(t, err)

	// Should be more than 0
	assert.NotEqual(t, len(mTeamMembers["team_members"]), 0)

	// Get a team member
	req, _ = http.NewRequest(
		"GET",
		ts.URL+"/service/teams/members/"+mTeamMembers["team_members"][0].ID.String(),
		nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Delete a team member
	req, _ = http.NewRequest(
		"DELETE",
		ts.URL+"/service/teams/members/"+mTeamMember["team_member"].ID.String(),
		nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
}
