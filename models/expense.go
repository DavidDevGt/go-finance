package models

import (
	"time"

	"gorm.io/gorm"
)

type FormattedDate time.Time

func (f FormattedDate) MarshalJSON() ([]byte, error) {
	t := time.Time(f)
	formatted := t.Format("02-01-2006")
	return []byte(`"` + formatted + `"`), nil
}

func CalculateWeek(date time.Time) int {
	_, week := date.ISOWeek()
	return week
}

type WeeklyBudget struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Week   int     `json:"week" binding:"required"`
	Year   int     `json:"year" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

type Expense struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description,omitempty"`
	Amount      float64        `json:"amount" binding:"required"`
	Category    string         `json:"category,omitempty"`
	DateRaw     time.Time      `gorm:"column:date_raw" json:"-"`
	Date        FormattedDate  `gorm:"-" json:"date"`
	Week        int            `json:"week"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
