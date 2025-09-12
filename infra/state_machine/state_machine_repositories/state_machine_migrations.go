package state_machine_repositories

import (
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/logger-go/logger"
	"gorm.io/gorm"
)

var CreateStateMachinesTable1721984741191 = database.Migration{
	ID: "1721984741191_CreateStateMachinesTable",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
            CREATE TABLE IF NOT EXISTS state_machines (
                id UUID PRIMARY KEY,
                machine_id VARCHAR(255) UNIQUE NOT NULL,
                blob JSONB NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                deleted_at TIMESTAMP WITH TIME ZONE
            );
            CREATE INDEX IF NOT EXISTS idx_state_machines_deleted_at ON state_machines(deleted_at);
        `).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
            DROP TABLE IF EXISTS state_machines;
        `).Error
	},
}

var AdaptMachineStateBlob1725028128419 = database.Migration{
	ID: "1725028128419_AdaptMachineStateBlob",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			-- Create a temporary function to convert the old format to the new format
			CREATE OR REPLACE FUNCTION convert_state_machine(old_data jsonb, entity_updated_at timestamp with time zone) RETURNS jsonb AS $$
			DECLARE
				new_data jsonb;
				visited_states jsonb;
			BEGIN
				-- Check if current_state_at is null or undefined
				IF old_data->>'current_state_at' IS NULL THEN
					-- Initialize the new data structure
					new_data := jsonb_build_object(
						'id', old_data->>'id',
						'states', old_data->'states',
						'transitions', old_data->'transitions',
						'current_state', old_data->>'current_state',
						'current_state_at', entity_updated_at,
						'visited_states', '[]'::jsonb
					);
			
					-- Convert visited_states to the new format
					visited_states := (
						SELECT jsonb_agg(
							jsonb_build_object(
								'state', state_name,
								'at', entity_updated_at
							)
						)
						FROM jsonb_array_elements_text(old_data->'visited_states') AS state_name
					);
			
					-- Add the converted visited_states to the new data structure
					new_data := new_data || jsonb_build_object('visited_states', visited_states);
			
					RETURN new_data;
				ELSE
					-- If current_state_at is not null, return the original data unchanged
					RETURN old_data;
				END IF;
			END;
			$$ LANGUAGE plpgsql;
			
			-- Update the state_machines table
			UPDATE state_machines
			SET blob = convert_state_machine(blob::jsonb, updated_at);
			
			-- Drop the temporary function
			DROP FUNCTION convert_state_machine(jsonb, timestamp with time zone);`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		logger.Logger.Error("rollback of %s not available; previous version is incompatible with current code", "1725028128419_AdaptMachineStateBlob")
		return nil
	},
}
