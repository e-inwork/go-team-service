// Copyright 2023, e-inwork.com. All rights reserved.

package mocks

import (
	"time"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/google/uuid"
)

type TeamModel struct{}

func (m TeamModel) Insert(gRPCTeam string, team *data.Team) error {
	team.ID = MockFirstUUID()
	team.CreatedAt = time.Now()
	team.Version = 1

	return nil
}

func (m TeamModel) GetByID(id uuid.UUID) (*data.Team, error) {
	teamId := MockFirstUUID()

	if teamId == id {
		var team = &data.Team{
			ID:          teamId,
			CreatedAt:   time.Now(),
			TeamUser:    teamId,
			TeamName:    "Doe's Team",
			TeamPicture: "77134e81-0cbe-4148-bb41-f0eecd56ac1d.jpg",
			Version:     1,
		}

		return team, nil
	}

	return nil, data.ErrRecordNotFound
}

func (m TeamModel) GetByTeamUser(teamUser uuid.UUID) (*data.Team, error) {
	teamUserId := MockFirstUUID()

	if teamUserId == teamUser {
		var team = &data.Team{
			ID:          teamUserId,
			CreatedAt:   time.Now(),
			TeamUser:    teamUserId,
			TeamName:    "Doe's Team",
			TeamPicture: "77134e81-0cbe-4148-bb41-f0eecd56ac1d.jpg",
			Version:     1,
		}

		return team, nil
	}

	return nil, data.ErrRecordNotFound
}

func (m TeamModel) Update(gRPCTeam string, team *data.Team) error {
	team.Version = team.Version + 1

	return nil
}
