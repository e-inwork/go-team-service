package mocks

import (
	"time"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/google/uuid"
)

type TeamMemberModel struct{}

func (m TeamMemberModel) Insert(teamMember *data.TeamMember) error {
	teamMember.ID = MockFirstUUID()
	teamMember.CreatedAt = time.Now()

	return nil
}

func (m TeamMemberModel) GetByID(id uuid.UUID) (*data.TeamMember, error) {
	teamMemberId := MockFirstUUID()

	if id == teamMemberId {
		var teamMember = &data.TeamMember{
			ID:                      teamMemberId,
			CreatedAt:               time.Now(),
			TeamMemberTeam:          teamMemberId,
			TeamMemberTeamName:      "Doe's Team",
			TeamMemberUser:          MockSecondUUID(),
			TeamMemberUserFirstName: "Nina",
			TeamMemberUserLastName:  "Doe",
		}

		return teamMember, nil
	}

	return nil, data.ErrRecordNotFound
}

func (m TeamMemberModel) ListByOwner(teamMemberTeam uuid.UUID) ([]*data.TeamMember, error) {
	teamMemberTeamId := MockFirstUUID()

	teamMembers := []*data.TeamMember{}

	if teamMemberTeam == teamMemberTeamId {
		var teamMember = &data.TeamMember{
			ID:                      MockFirstUUID(),
			CreatedAt:               time.Now(),
			TeamMemberTeam:          teamMemberTeam,
			TeamMemberTeamName:      "Doe's Team",
			TeamMemberUser:          MockSecondUUID(),
			TeamMemberUserFirstName: "Nina",
			TeamMemberUserLastName:  "Doe",
		}

		teamMembers = append(teamMembers, teamMember)
	}

	return teamMembers, nil
}

func (m TeamMemberModel) Delete(teamMember *data.TeamMember) error {
	id := MockFirstUUID()

	if teamMember.ID != id {
		return data.ErrRecordNotFound
	}

	return nil
}
