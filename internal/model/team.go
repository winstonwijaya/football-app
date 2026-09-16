package model

type Team struct {
	ID              int64   `gorm:"column:id;primaryKey"`
	Name            string  `gorm:"column:name;not null"`
	LogoURL         *string `gorm:"column:logo_url"`
	EstablishedYear *int16  `gorm:"column:established_year"`
	Address         *string `gorm:"column:address"`
	City            *string `gorm:"column:city"`

	Audit
}

func (Team) TableName() string {
	return "teams"
}
