package state_machine_repositories

import (
	"fmt"
	"time"

	"github.com/pixie-sh/database-helpers-go/database"
	"gorm.io/gorm/clause"

	"github.com/pixie-sh/core-go/infra/state_machine"
	"github.com/pixie-sh/core-go/pkg/models/database_models"
	mapper "github.com/pixie-sh/core-go/pkg/models/serializer"
)

type StateMachineRepository struct {
	database.Repository[StateMachineRepository]
}

func NewStateMachineRepository(db *database.DB) StateMachineRepository {
	return StateMachineRepository{database.NewRepository(db, NewStateMachineRepository)}
}

func (r StateMachineRepository) GetByMachineID(machineID string) (sm StateMachine, e error) {
	return sm, r.DB.Model(&StateMachine{}).
		Where("machine_id", machineID).
		First(&sm).
		Error
}

func (r StateMachineRepository) SaveByMachineID(guid string, machineID string, blob database_models.JSONB) (StateMachine, error) {
	now := time.Now()
	data := StateMachine{
		ID:        guid,
		MachineID: machineID,
		Blob:      blob,
		SoftDeletable: database_models.SoftDeletable{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	err := r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "machine_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"id", "blob", "updated_at"}),
	}).Create(&data).Error

	return data, err
}

// Store implements state_machine.StateMachineStorage
func (r StateMachineRepository) Store(m state_machine.MachineModel) error {
	blob, err := mapper.ToJSONB(m)
	if err != nil {
		return err
	}

	_, err = r.SaveByMachineID(m.Guid, m.ID, blob)
	return err
}

// Get implements state_machine.StateMachineStorage
func (r StateMachineRepository) Get(machineID string) (state_machine.MachineModel, error) {
	machineEntity, err := r.GetByMachineID(machineID)
	if err != nil {
		return state_machine.MachineModel{}, err
	}

	var machineModel state_machine.MachineModel
	err = mapper.ToStruct(machineEntity.Blob, &machineModel)
	if err != nil {
		return machineModel, err
	}

	return machineModel, nil
}

func (r StateMachineRepository) CountByEntityNameAndStateTransition(entityName string, state string, transition string) (count int64, err error) {
	jsonQuery := fmt.Sprintf("blob->'transitions'->'%s' ? '%s'", state, transition)
	return count, r.DB.Model(&StateMachine{}).
		Unscoped().
		Where("machine_id LIKE ?", entityName+"%").
		Where(jsonQuery).
		Count(&count).Error
}
