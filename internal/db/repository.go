package db

import "time"

func now() int64 { return time.Now().UnixMilli() }

func (s *Store) CreateRun(id, input string) error {
	ts := now()
	_, err := s.db.Exec(`
		INSERT INTO runs(id,status,input,created_at,updated_at)
		VALUES(?,?,?,?,?)
	`, id, "RUNNING", input, ts, ts)
	return err
}

func (s *Store) CompleteRun(id, status, output, runErr string) error {
	_, err := s.db.Exec(`
		UPDATE runs
		SET status=?,output=?,error=?,updated_at=?
		WHERE id=?
	`, status, output, runErr, now(), id)
	return err
}

func (s *Store) CreateStep(id, runID string, stepNo int, kind, input string) error {
	ts := now()
	_, err := s.db.Exec(`
		INSERT INTO steps(
			id,run_id,step_no,kind,status,input,created_at,updated_at
		)
		VALUES(?,?,?,?,?,?,?,?)
	`, id, runID, stepNo, kind, "RUNNING", input, ts, ts)
	return err
}

func (s *Store) CompleteStep(id, status, output, stepErr string) error {
	_, err := s.db.Exec(`
		UPDATE steps
		SET status=?,output=?,error=?,updated_at=?
		WHERE id=?
	`, status, output, stepErr, now(), id)
	return err
}

func (s *Store) CreateToolCall(
	id, runID, stepID, name, args, idempotencyKey string,
	retryable bool, maxAttempts int,
) error {
	ts := now()
	flag := 0
	if retryable {
		flag = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO tool_calls(
			id,run_id,step_id,tool_name,arguments,status,
			attempt,max_attempts,idempotency_key,retryable,
			created_at,updated_at
		)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
	`, id, runID, stepID, name, args, "RUNNING",
		1, maxAttempts, idempotencyKey, flag, ts, ts)
	return err
}

func (s *Store) CompleteToolCall(id, status, result, toolErr string) error {
	_, err := s.db.Exec(`
		UPDATE tool_calls
		SET status=?,result=?,error=?,updated_at=?
		WHERE id=?
	`, status, result, toolErr, now(), id)
	return err
}

func (s *Store) CreateCheckpoint(id, runID, stepID, state string) error {
	_, err := s.db.Exec(`
		INSERT INTO checkpoints(id,run_id,step_id,state,created_at)
		VALUES(?,?,?,?,?)
	`, id, runID, stepID, state, now())
	return err
}
