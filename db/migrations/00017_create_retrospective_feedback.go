package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00017, Down00017)
}

// Up00017 ...
func Up00017(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.RetrospectiveFeedback{})

	gormdb.Model(&models.RetrospectiveFeedback{}).AddForeignKey("retrospective_id", "retrospectives(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.RetrospectiveFeedback{}).AddForeignKey("created_by_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.RetrospectiveFeedback{}).AddForeignKey("assignee_id", "users(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.RetrospectiveFeedback{}).AddIndex("idx_sprints_created_by_id_retro_id_type", "created_by_id", "retrospective_id", "type")
	gormdb.Model(&models.RetrospectiveFeedback{}).AddIndex("idx_sprints_assignee_id_retro_id_type", "assignee_id", "retrospective_id", "type")
	gormdb.Model(&models.RetrospectiveFeedback{}).AddIndex("idx_sprints_type_retro_id_created_at_resolved_at", "type", "retrospective_id", "created_at", "resolved_at")
	gormdb.Model(&models.RetrospectiveFeedback{}).AddIndex("idx_sprints_retro_id_created_at_resolved_at", "retrospective_id", "created_at", "resolved_at")

	return nil
}

// Down00017 ...
func Down00017(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveIndex("idx_sprints_retro_id_created_at_resolved_at")
	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveIndex("idx_sprints_type_retro_id_created_at_resolved_at")
	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveIndex("idx_sprints_assignee_id_retro_id_type")
	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveIndex("idx_sprints_created_by_id_retro_id_type")

	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveForeignKey("assignee_id", "users(id)")
	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveForeignKey("created_by_id", "users(id)")
	gormdb.Model(&models.RetrospectiveFeedback{}).RemoveForeignKey("retrospective_id", "retrospectives(id)")

	gormdb.DropTable(&models.RetrospectiveFeedback{})

	return nil
}
