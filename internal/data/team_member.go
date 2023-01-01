// Copyright 2022, e-inwork.com. All rights reserved.

package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/e-inwork-com/go-team-service/internal/validator"

	"github.com/google/uuid"
)

type TeamMember struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	TeamMemberTeam uuid.UUID `json:"team_member_team"`
	TeamMemberUser uuid.UUID `json:"team_member_user"`
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
        SELECT id, created_at, team_member_team, team_member_user
        FROM team_members
        WHERE id = $1`

	var teamMember TeamMember

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&teamMember.ID,
		&teamMember.CreatedAt,
		&teamMember.TeamMemberTeam,
		&teamMember.TeamMemberUser,
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
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
