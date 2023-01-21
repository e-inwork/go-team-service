package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrCreateConflict = errors.New("create conflict")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Teams       TeamModelInterface
	Users       UserModelInterface
	TeamMembers TeamMemberModelInterface
}

func InitModels(db *sql.DB) Models {
	return Models{
		Teams:       TeamModel{DB: db},
		Users:       UserModel{DB: db},
		TeamMembers: TeamMemberModel{DB: db},
	}
}
