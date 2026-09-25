package db

import (
	"database/sql"
)

type RecoveryToolCall struct {
	ID             string
	RunID          string
	StepID         string
	ToolName       string
	Arguments      string
	Status         string
	Attempt        int
	MaxAttempts    int
	IdempotencyKey string
	Retryable      bool
	LastError      string
}

func (s *Store) FindRunningToolCalls() ([]RecoveryToolCall, error) {
	rows, err := s.db.Query(`
		SELECT id, run_id, step_id, tool_name, arguments, status,
		       attempt, max_attempts, idempotency_key, retryable, error
		FROM tool_calls
		WHERE status = 'RUNNING'
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecoveryToolCall
	for rows.Next() {
		var x RecoveryToolCall
		var retryable int
		if err := rows.Scan(
			&x.ID, &x.RunID, &x.StepID, &x.ToolName, &x.Arguments,
			&x.Status, &x.Attempt, &x.MaxAttempts, &x.IdempotencyKey,
			&retryable, &x.LastError,
		); err != nil {
			return nil, err
		}
		x.Retryable = retryable == 1
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) MarkToolCallRetrying(id string) error {
	_, err := s.db.Exec(`
		UPDATE tool_calls
		SET status='RETRYING',
		    attempt=attempt+1,
		    updated_at=?
		WHERE id=?
	`, now(), id)
	return err
}

func (s *Store) MarkToolCallFailed(id, errMsg string) error {
	_, err := s.db.Exec(`
		UPDATE tool_calls
		SET status='FAILED',
		    error=?,
		    updated_at=?
		WHERE id=?
	`, errMsg, now(), id)
	return err
}

func (s *Store) GetToolCall(id string) (*RecoveryToolCall, error) {
	row := s.db.QueryRow(`
		SELECT id, run_id, step_id, tool_name, arguments, status,
		       attempt, max_attempts, idempotency_key, retryable, error
		FROM tool_calls
		WHERE id=?
	`, id)

	var x RecoveryToolCall
	var retryable int
	if err := row.Scan(
		&x.ID, &x.RunID, &x.StepID, &x.ToolName, &x.Arguments,
		&x.Status, &x.Attempt, &x.MaxAttempts, &x.IdempotencyKey,
		&retryable, &x.LastError,
	); err != nil {
		return nil, err
	}
	x.Retryable = retryable == 1
	return &x, nil
}

var _ *sql.DB
