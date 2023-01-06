// Copyright 2022, e-inwork.com. All rights reserved.

package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/e-inwork-com/go-team-service/grpc/teams"
	"github.com/e-inwork-com/go-team-service/pkg/data"
	"github.com/e-inwork-com/go-team-service/pkg/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (app *Application) createTeamHandler(w http.ResponseWriter, r *http.Request) {
	// Get a name
	teamName := r.FormValue("team_name")

	// Read a file attachment
	file, fileHeader, err := r.FormFile("team_picture")
	if err == nil {
		defer file.Close()
	}

	// Get the current user
	user := app.contextGetUser(r)

	// Set profile picture
	teamPicture := ""
	if file != nil {
		teamPicture = fmt.Sprintf("%s%s", user.ID.String(), filepath.Ext(fileHeader.Filename))
	}

	// Set a Team
	team := &data.Team{
		TeamUser:    user.ID,
		TeamName:    teamName,
		TeamPicture: teamPicture,
	}

	// Validate Profile
	v := validator.New()
	if data.ValidateTeam(v, team); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check type of file
	if teamPicture != "" {
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			http.Error(w, "Please upload a JPEG or PNG image", http.StatusBadRequest)
			return
		}

		// Read from the beginning of the file offset
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create an uploading folder if it doesn't
		// already exist
		err = os.MkdirAll(app.Config.Uploads, os.ModePerm)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create a new file in the uploads directory
		dst, err := os.Create(fmt.Sprintf("%s/%s", app.Config.Uploads, teamPicture))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		defer dst.Close()

		// Copy the uploaded file to the filesystem
		// at the specified destination
		_, err = io.Copy(dst, file)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Insert data to Team
	err = app.Models.Teams.Insert(team)
	if err != nil {
		switch {
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send to gRPC - Go Team Search Service
	conn, err := grpc.Dial("go-team-search-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer conn.Close()

	c := teams.NewTeamServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteTeam(ctx, &teams.TeamRequest{
		TeamEntry: &teams.Team{
			Id: team.ID.String(),
		},
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a data as response of the HTTP request
	err = app.writeJSON(w, http.StatusCreated, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getOwnTeamHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current user as the owner of the record
	user := app.contextGetUser(r)

	// Get team by user
	team, err := app.Models.Teams.GetByTeamUser(user.ID)

	// Check error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send a request response
	err = app.writeJSON(w, http.StatusOK, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) patchTeamHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID from the request parameters
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Get a record from the database
	team, err := app.Models.Teams.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Get the current user
	user := app.contextGetUser(r)

	// Check if the record has a related to the current user
	// Only the owner of the record can update the own record
	if team.TeamUser != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	// Get a profile name
	teamName := r.FormValue("team_name")

	// Read a file attachment
	file, fileHeader, err := r.FormFile("team_picture")
	if err == nil {
		defer file.Close()
	}

	// Set picture
	teamPicture := ""
	if file != nil {
		teamPicture = fmt.Sprintf("%s%s", user.ID.String(), filepath.Ext(fileHeader.Filename))
	}

	// Set a new Profile
	newTeam := &data.Team{
		TeamName:    teamName,
		TeamPicture: teamPicture,
	}

	if teamPicture != "" {
		// Check type of file
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			http.Error(w, "Please upload a JPEG or PNG image", http.StatusBadRequest)
			return
		}

		// Read a file from the beginning offset
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create an uploading folder if it doesn't
		// already exist
		err = os.MkdirAll(app.Config.Uploads, os.ModePerm)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Delete the old profile picture
		if _, err := os.Stat(fmt.Sprintf("%s/%s", app.Config.Uploads, team.TeamPicture)); err == nil {
			err = os.Remove(fmt.Sprintf("%s/%s", app.Config.Uploads, team.TeamPicture))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		// Create a new file in the uploads directory
		dst, err := os.Create(fmt.Sprintf("%s/%s", app.Config.Uploads, teamPicture))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the filesystem
		// at the specified destination
		_, err = io.Copy(dst, file)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Update the old profile picture with a new one
	if newTeam.TeamName != "" {
		team.TeamName = newTeam.TeamName
	}

	if newTeam.TeamPicture != "" {
		team.TeamPicture = newTeam.TeamPicture
	}

	// Update the Profile
	err = app.Models.Teams.Update(team)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send back the record to the request response
	err = app.writeJSON(w, http.StatusOK, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	// Get file from the request parameters
	file, err := app.readFileParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Read file
	buffer, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", app.Config.Uploads, file))
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Check type of file
	filetype := http.DetectContentType(buffer)

	w.Header().Set("Content-Type", filetype)
	w.Write(buffer)
}
