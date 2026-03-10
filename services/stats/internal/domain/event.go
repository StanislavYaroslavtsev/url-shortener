package domain

import "time"

type ClickEvent struct {
	Code      string
	IP        string
	Country   string
	ClickedAt time.Time
}
