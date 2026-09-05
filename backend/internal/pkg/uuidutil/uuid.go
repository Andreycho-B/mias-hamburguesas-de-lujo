package uuidutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func FromPG(pgID pgtype.UUID) uuid.UUID {
	return uuid.UUID(pgID.Bytes)
}
