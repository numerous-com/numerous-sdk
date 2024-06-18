package app

import "time"

type App struct {
	ID          string
	Name        string
	Description string
	PublicURL   string
	SharedURL   string
	CreatedAt   time.Time
}
