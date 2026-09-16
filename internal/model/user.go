package model

type User struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	FirstName    string `gorm:"column:first_name;not null"`
	LastName     string `gorm:"column:last_name;not null"`
	Username     string `gorm:"column:username;not null"`
	PasswordHash string `gorm:"column:password_hash;not null"`

	Audit
}

func (User) TableName() string {
	return "users"
}
