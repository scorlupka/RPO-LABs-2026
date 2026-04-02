package models

import "time"

type Terminal struct {
	ID           int64
	SerialNumber string
	Address      string
	Name         string
	Status       bool
	CreatedAt    time.Time
}

type Card struct {
	ID         int64
	Number     string
	Balance    int64
	IsBlocked  bool
	OwnerName  string
	ExpireDate time.Time
	KeyID      int64
	CreatedAt  time.Time
}

type Transaction struct {
	ID         int64
	Amount     int64
	CardID     int64
	TerminalID int64
	CreatedAt  time.Time
}

type Key struct {
	ID        int64
	KeyValue  string
	CreatedAt time.Time
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
}
