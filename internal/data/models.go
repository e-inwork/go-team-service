// Copyright 2022, e-inwork.com. All rights reserved.

package data

import (
	"database/sql"
	"errors"

	dataUser "github.com/e-inwork-com/go-user-service/pkg/data"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Teams TeamModel
	Users dataUser.UserModel
}

func InitModels(db *sql.DB) Models {
	return Models{
		Teams: TeamModel{DB: db},
		Users: dataUser.UserModel{DB: db},
	}
}
