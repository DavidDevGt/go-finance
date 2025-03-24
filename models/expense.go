package models

import (
	"time"

	"gorm.io/gorm"
)

type FormattedDate time.Time

// MarshalJSON muestra la fecha en formato "DD-MM-2006" al enviarla en la respuesta.
func (f FormattedDate) MarshalJSON() ([]byte, error) {
	t := time.Time(f)
	formatted := t.Format("02-01-2006")
	return []byte(`"` + formatted + `"`), nil
}

func (f *FormattedDate) UnmarshalJSON(data []byte) error {
	s := string(data)
	// Eliminar comillas
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse("02-01-2006", s)
		if err != nil {
			return err
		}
	}
	*f = FormattedDate(t)
	return nil
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
