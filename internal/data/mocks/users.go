package mocks

import (
	"time"

	"github.com/e-inwork-com/go-team-service/internal/data"
	"github.com/google/uuid"
)

type UserModel struct{}

func (m UserModel) GetByID(id uuid.UUID) (*data.User, error) {
	if MockFirstUUID() == id {
		var user = &data.User{
			ID:        id,
			CreatedAt: time.Now(),
			Email:     "jon@doe.com",
			FirstName: "Jon",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}

		return user, nil
	}

	if MockSecondUUID() == id {
		var user = &data.User{
			ID:        id,
			CreatedAt: time.Now(),
			Email:     "nina@doe.com",
			FirstName: "nina",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}

		return user, nil
	}

	return nil, data.ErrRecordNotFound
}
