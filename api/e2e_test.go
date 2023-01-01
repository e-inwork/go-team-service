// Copyright 2022, e-inwork.com. All rights reserved.

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/e-inwork-com/go-team-service/internal/jsonlog"
	apiUser "github.com/e-inwork-com/go-user-service/api"
	dataUser "github.com/e-inwork-com/go-user-service/pkg/data"
	jsonLogUser "github.com/e-inwork-com/go-user-service/pkg/jsonlog"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

var cfgUser apiUser.Config

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	// Tell docker to hard kill the container in 120 seconds
	resource.Expire(120)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Server Setup
	cfgUser.Db.Dsn = databaseUrl
	cfgUser.Auth.Secret = "secret"
	cfgUser.Db.MaxOpenConn = 25
	cfgUser.Db.MaxIdleConn = 25
	cfgUser.Db.MaxIdleTime = "15m"
	cfgUser.Limiter.Enabled = true
	cfgUser.Limiter.Rps = 2
	cfgUser.Limiter.Burst = 4

	db, _ := apiUser.OpenDB(cfgUser)
	defer db.Close()

	_, _ = db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
			created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
			email text UNIQUE NOT NULL,
			password_hash bytea NOT NULL,
			first_name char varying(100) NOT NULL,
			last_name char varying(100) NOT NULL,
			activated bool NOT NULL DEFAULT false,
			version integer NOT NULL DEFAULT 1
		);
	`)

	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS teams (
			id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
			created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
			team_user UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE UNIQUE,
			team_name char varying(100) NOT NULL,
			team_picture char varying(512),
			version integer NOT NULL DEFAULT 1
		);
	`)

	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS team_members (
			id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
			created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
			team_member_team UUID NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
			team_member_user UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
			UNIQUE(team_member_team, team_member_user)
		);
	`)

	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	// User Service
	loggerUser := jsonLogUser.New(os.Stdout, jsonLogUser.LevelInfo)

	db, _ := apiUser.OpenDB(cfgUser)
	defer db.Close()

	appUser := &apiUser.Application{
		Config: cfgUser,
		Logger: loggerUser,
		Models: dataUser.InitModels(db),
	}

	tsUser := httptest.NewTLSServer(appUser.Routes())
	defer tsUser.Close()

	// Register
	email := "jon@doe.com"
	password := "pa55word"
	user := fmt.Sprintf(`{"email": "%v", "password":  "%v", "first_name": "Jon", "last_name": "Doe"}`, email, password)
	res, err := tsUser.Client().Post(tsUser.URL+"/service/users", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusAccepted)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var mUser map[string]dataUser.User
	err = json.Unmarshal(body, &mUser)
	assert.Nil(t, err)
	assert.Equal(t, mUser["user"].Email, email)

	// User sign in to get a token
	login := fmt.Sprintf(`{"email": "%v", "password":  "%v"}`, email, password)
	res, err = tsUser.Client().Post(tsUser.URL+"/service/users/authentication", "application/json", bytes.NewReader([]byte(login)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	type authType struct {
		Token string `json:"token"`
	}
	var authResult authType
	err = json.Unmarshal(body, &authResult)

	assert.Nil(t, err)
	assert.NotNil(t, authResult.Token)

	// Team Service
	var cfgTeam Config
	cfgTeam.Db.Dsn = cfgUser.Db.Dsn
	cfgTeam.Auth.Secret = "secret"
	cfgTeam.Db.MaxOpenConn = 25
	cfgTeam.Db.MaxIdleConn = 25
	cfgTeam.Db.MaxIdleTime = "15m"
	cfgTeam.Limiter.Enabled = true
	cfgTeam.Limiter.Rps = 2
	cfgTeam.Limiter.Burst = 6
	cfgTeam.Uploads = "../uploads"

	loggerTeam := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	appTeam := Application{
		Config: cfgTeam,
		Logger: loggerTeam,
		Models: data.InitModels(db),
	}

	// Server Routes API
	tsTeam := httptest.NewTLSServer(appTeam.routes())
	defer tsTeam.Close()

	// Create Team
	// Create body buffer
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// Add team name
	bodyWriter.WriteField("team_name", "Doe's Team")

	// Add team picture
	filename := "./test/team.jpg"
	fileWriter, err := bodyWriter.CreateFormFile("team_picture", filename)
	if err != nil {
		fmt.Println("Error writing to buffer")
	}

	// Open file
	fileHandler, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file")
	}

	// Copy file
	_, err = io.Copy(fileWriter, fileHandler)
	if err != nil {
		fmt.Println("Error copy file")
	}

	// Put on body
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// Post a new team
	req, _ := http.NewRequest("POST", tsTeam.URL+"/service/teams", bodyBuf)
	req.Header.Add("Content-Type", contentType)

	bearer := fmt.Sprintf("Bearer %v", authResult.Token)
	req.Header.Set("Authorization", bearer)

	res, err = tsTeam.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Read response
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var mTeam map[string]data.Team
	err = json.Unmarshal(body, &mTeam)
	assert.Nil(t, err)
	assert.Equal(t, mTeam["team"].TeamUser, mUser["user"].ID)

	// Get a picture
	req, _ = http.NewRequest("GET", tsTeam.URL+"/service/teams/pictures/"+mTeam["team"].TeamPicture, nil)

	res, err = tsTeam.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Read response
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	// Check type of file
	filetype := http.DetectContentType(body)
	assert.NotEqual(t, filetype, "")

	// Register a user for the team member
	user = `{"email": "nina@doe.com", "password":  "pa55word", "first_name": "Nina", "last_name": "Doe"}`
	res, err = tsUser.Client().Post(tsUser.URL+"/service/users", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusAccepted)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var mNewUser map[string]dataUser.User
	err = json.Unmarshal(body, &mNewUser)
	assert.Nil(t, err)

	// Create a new team member
	teamMember := fmt.Sprintf(`{"team_member_team": "%v", "team_member_user":  "%v"}`, mTeam["team"].ID, mNewUser["user"].ID)
	req, _ = http.NewRequest("POST", tsTeam.URL+"/service/teams/members", bytes.NewReader([]byte(teamMember)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = tsTeam.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Read response
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var mTeamMember map[string]data.TeamMember
	err = json.Unmarshal(body, &mTeamMember)
	assert.Nil(t, err)
	assert.Equal(t, mTeamMember["team_member"].TeamMemberTeam, mTeam["team"].ID)

	// Get a list team members of the current user
	req, _ = http.NewRequest("GET", tsTeam.URL+"/service/teams/members", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = tsTeam.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Delete a team member
	req, _ = http.NewRequest("DELETE", tsTeam.URL+"/service/teams/members/"+mTeamMember["team_member"].ID.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	res, err = tsTeam.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
}
