// Copyright 2022, e-inwork.com. All rights reserved.

package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Teams       TeamModel
	Users       UserModel
	TeamMembers TeamMemberModel
}

func InitModels(db *sql.DB) Models {
	return Models{
		Teams:       TeamModel{DB: db},
		Users:       UserModel{DB: db},
		TeamMembers: TeamMemberModel{DB: db},
	}
}
