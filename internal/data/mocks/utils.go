package mocks

import "github.com/google/uuid"

func MockFirstUUID() uuid.UUID {
	id, _ := uuid.Parse("77134e81-0cbe-4148-bb41-f0eecd56ac1d")
	return id
}

func MockSecondUUID() uuid.UUID {
	id, _ := uuid.Parse("77134e81-0cbe-4148-bb41-f0eecd56ac11")
	return id
}
