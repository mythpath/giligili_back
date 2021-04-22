package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func shouldSaveAssociations(tx *gorm.DB) bool {
	if saveAssociations, ok := tx.Get("gorm:save_associations"); ok && !saveAssociations.(bool) {
		return false
	}
	return tx.Error == nil
}

func changeableField(tx *gorm.DB, field *schema.Field) bool {
	if selectAttrs := tx.Statement.Selects; len(selectAttrs) > 0 {
		for _, attr := range selectAttrs {
			if field.Name == attr || field.DBName == attr {
				return true
			}
		}
		return false
	}

	for _, attr := range tx.Statement.Omits {
		if field.Name == attr || field.DBName == attr {
			return false
		}
	}

	return true
}
