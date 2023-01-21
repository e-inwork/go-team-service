package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/e-inwork-com/go-team-service/internal/validator"

	"github.com/google/uuid"
)

type TeamMemberModelInterface interface {
	Insert(teamMember *TeamMember) error
	GetByID(id uuid.UUID) (*TeamMember, error)
	ListByOwner(teamMemberTeam uuid.UUID) ([]*TeamMember, error)
	Delete(teamMember *TeamMember) error
}

type TeamMember struct {
	ID                      uuid.UUID `json:"id"`
	CreatedAt               time.Time `json:"created_at"`
	TeamMemberTeam          uuid.UUID `json:"team_member_team"`
	TeamMemberTeamName      string    `json:"team_member_team_name"`
	TeamMemberUser          uuid.UUID `json:"team_member_user"`
	TeamMemberUserFirstName string    `json:"team_member_user_first_name"`
	TeamMemberUserLastName  string    `json:"team_member_user_last_name"`
}

func ValidateTeamMember(v *validator.Validator, teamMember *TeamMember) {
	v.Check(teamMember.TeamMemberTeam != uuid.Nil, "team_member_team", "must be provided")
	v.Check(teamMember.TeamMemberUser != uuid.Nil, "team_member_user", "must be provided")
}

type TeamMemberModel struct {
	DB *sql.DB
}

func (m TeamMemberModel) Insert(teamMember *TeamMember) error {
	query := `
        INSERT INTO team_members (team_member_team, team_member_user)
        VALUES ($1, $2)
        RETURNING id, created_at`

	args := []interface{}{teamMember.TeamMemberTeam, teamMember.TeamMemberUser}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&teamMember.ID, &teamMember.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (m TeamMemberModel) GetByID(id uuid.UUID) (*TeamMember, error) {
	query := `
    SELECT
			team_members.id,
			team_members.created_at,
			team_member_team,
			teams.team_name as team_member_team_name,
			team_member_user,
			users.first_name as team_member_user_first_name,
			users.last_name as team_member_user_last_name
    FROM team_members, teams, users
		WHERE team_members.id = $1
		AND team_member_team = teams.id
		AND team_member_user = users.id
	`

	var teamMember TeamMember

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&teamMember.ID,
		&teamMember.CreatedAt,
		&teamMember.TeamMemberTeam,
		&teamMember.TeamMemberTeamName,
		&teamMember.TeamMemberUser,
		&teamMember.TeamMemberUserFirstName,
		&teamMember.TeamMemberUserLastName,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &teamMember, nil
}

func (m TeamMemberModel) ListByOwner(teamMemberTeam uuid.UUID) ([]*TeamMember, error) {
	query := `
    SELECT
			team_members.id,
			team_members.created_at,
			team_member_team,
			teams.team_name as team_member_team_name,
			team_member_user,
			users.first_name as team_member_user_first_name,
			users.last_name as team_member_user_last_name
    FROM team_members, teams, users
		WHERE team_member_team = $1
		AND team_member_team = teams.id
		AND team_member_user = users.id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, teamMemberTeam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teamMembers := []*TeamMember{}

	for rows.Next() {
		var teamMember TeamMember

		err = rows.Scan(
			&teamMember.ID,
			&teamMember.CreatedAt,
			&teamMember.TeamMemberTeam,
			&teamMember.TeamMemberTeamName,
			&teamMember.TeamMemberUser,
			&teamMember.TeamMemberUserFirstName,
			&teamMember.TeamMemberUserLastName,
		)
		if err != nil {
			return nil, err
		}

		teamMembers = append(teamMembers, &teamMember)
	}

	return teamMembers, nil
}

func (m TeamMemberModel) Delete(teamMember *TeamMember) error {
	query := `
        DELETE FROM team_members
        WHERE id = $1`

	args := []interface{}{
		teamMember.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
