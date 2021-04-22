package orm

import (
	"gorm.io/gorm"
)

// DBService
type DBService interface {
	GetDB() *gorm.DB
}
