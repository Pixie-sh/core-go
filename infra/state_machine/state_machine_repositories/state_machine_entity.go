package state_machine_repositories

import "github.com/pixie-sh/core-go/pkg/models/database_models"

type StateMachine struct {
	database_models.SoftDeletable

	ID        string `gorm:"type:uuid;primaryKey"`
	MachineID string `gorm:"type:text;uniqueIndex"`
	Blob      database_models.JSONB
} //@name StateMachine

func (StateMachine) TableName() string {
	return "state_machines"
}
