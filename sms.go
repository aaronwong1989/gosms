package gosms

import (
	"time"
)

type TextSms struct {
	Phones   []string  `json:"phones"   validate:"required"`
	Content  string    `json:"content"  validate:"required"`
	Priority int       `json:"priority" validate:"omitempty"`
	AtTime   time.Time `json:"at-time"  validate:"omitempty"`
}

type TemplateSms struct {
	Phones []string          `json:"phones"   validate:"required"`
	TempID string            `json:"temp-id"  validate:"required"`
	Args   map[string]string `json:"args"     validate:"required"`
	AtTime time.Time         `json:"at-time"  validate:"omitempty"`
}
