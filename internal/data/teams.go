package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/e-inwork-com/go-team-service/internal/grpc/teams"
	"github.com/e-inwork-com/go-team-service/internal/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/google/uuid"
)

type TeamModelInterface interface {
	Insert(gRPCTeam string, team *Team) error
	GetByID(id uuid.UUID) (*Team, error)
	GetByTeamUser(teamUser uuid.UUID) (*Team, error)
	Update(gRPCTeam string, team *Team) error
}

type Team struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	TeamUser    uuid.UUID `json:"team_user"`
	TeamName    string    `json:"team_name"`
	TeamPicture string    `json:"team_picture"`
	Version     int       `json:"-"`
}

type TeamModel struct {
	DB *sql.DB
}

func ValidateTeam(v *validator.Validator, team *Team) {
	v.Check(team.TeamName != "", "team_name", "must be provided")
}

func gRPCTeamIndexing(gRPCTeam string, team *Team) error {
	// Send to gRPC - Go Team Indexing Service
	con, err := grpc.Dial(gRPCTeam, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer con.Close()

	c := teams.NewTeamServiceClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = c.WriteTeam(ctx, &teams.TeamRequest{
		TeamEntry: &teams.Team{
			Id: team.ID.String(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m TeamModel) Insert(gRPCTeam string, team *Team) error {
	query := `
        INSERT INTO teams (team_user, team_name, team_picture)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, version`

	args := []interface{}{team.TeamUser, team.TeamName, team.TeamPicture}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&team.ID, &team.CreatedAt, &team.Version)
	if err != nil {
		return err
	}

	err = gRPCTeamIndexing(gRPCTeam, team)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (m TeamModel) GetByID(id uuid.UUID) (*Team, error) {
	query := `
        SELECT id, created_at, team_user, team_name, team_picture, version
        FROM teams
        WHERE id = $1`

	var team Team

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.TeamUser,
		&team.TeamName,
		&team.TeamPicture,
		&team.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &team, nil
}

func (m TeamModel) GetByTeamUser(teamUser uuid.UUID) (*Team, error) {
	// Select query by owner
	query := `
        SELECT id, created_at, team_user, team_name, team_picture, version
        FROM teams
        WHERE team_user = $1`

	// Define a record variable
	var team Team

	// Create a context background
	// to use it with a query to database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Query by owner to the database,
	// and the assign the row result to the profile variable
	err := m.DB.QueryRowContext(ctx, query, teamUser).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.TeamUser,
		&team.TeamName,
		&team.TeamPicture,
		&team.Version,
	)

	// Check error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Return the result
	return &team, nil
}

func (m TeamModel) Update(gRPCTeam string, team *Team) error {
	// SQL Update
	query := `
        UPDATE teams
        SET team_name = $1, team_picture = $2, version = version + 1, is_indexed = false
        WHERE id = $3 AND version = $4
        RETURNING version`

	// Assign arguments
	args := []interface{}{
		team.TeamName,
		team.TeamPicture,
		team.ID,
		team.Version,
	}

	// Create a context of the SQL Update
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run SQL Update
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&team.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	err = gRPCTeamIndexing(gRPCTeam, team)
	if err != nil {
		log.Println(err)
	}

	return nil
}
