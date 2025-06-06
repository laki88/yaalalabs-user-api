// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Phone     sql.NullString
	Age       sql.NullInt32
	Status    sql.NullString
}
