package db

import (
	"database/sql"
	"time"
)

func now() int64 { return time.Now().UnixMilli() }

func (s *Store) CreateRun(id, input string) error {
	ts := now()
	_, err := s.db.Exec(`INSERT INTO runs(id,status,input,created_at,updated_at) VALUES(?,?,?,?,?)`, id, "RUNNING", input, ts, ts)
	return err
}
func (s *Store) CompleteRun(id, status, output, runErr string) error {
	_, err := s.db.Exec(`UPDATE runs SET status=?,output=?,error=?,updated_at=? WHERE id=?`, status, output, runErr, now(), id)
	return err
}
func (s *Store) CreateStep(id, runID string, stepNo int, kind, input string) error {
	ts := now()
	_, err := s.db.Exec(`INSERT INTO steps(id,run_id,step_no,kind,status,input,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`, id, runID, stepNo, kind, "RUNNING", input, ts, ts)
	return err
}
func (s *Store) CompleteStep(id, status, output, stepErr string) error {
	_, err := s.db.Exec(`UPDATE steps SET status=?,output=?,error=?,updated_at=? WHERE id=?`, status, output, stepErr, now(), id)
	return err
}
func (s *Store) CreateToolCall(id, runID, stepID, name, args, idempotencyKey string, retryable bool, maxAttempts int) error {
	ts := now()
	flag := 0
	if retryable {
		flag = 1
	}
	_, err := s.db.Exec(`INSERT INTO tool_calls(id,run_id,step_id,tool_name,arguments,status,attempt,max_attempts,idempotency_key,retryable,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, id, runID, stepID, name, args, "RUNNING", 1, maxAttempts, idempotencyKey, flag, ts, ts)
	return err
}
func (s *Store) CompleteToolCall(id, status, result, toolErr string) error {
	_, err := s.db.Exec(`UPDATE tool_calls SET status=?,result=?,error=?,updated_at=? WHERE id=?`, status, result, toolErr, now(), id)
	return err
}
func (s *Store) CreateCheckpoint(id, runID, stepID, state string) error {
	_, err := s.db.Exec(`INSERT INTO checkpoints(id,run_id,step_id,state,created_at) VALUES(?,?,?,?,?)`, id, runID, stepID, state, now())
	return err
}

type Run struct{ ID, Status, Input, Output, Error string }

func (s *Store) GetRun(id string) (*Run, error) {
	var r Run
	err := s.db.QueryRow(`SELECT id,status,input,output,error FROM runs WHERE id=?`, id).Scan(&r.ID, &r.Status, &r.Input, &r.Output, &r.Error)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func (s *Store) FindRunningRuns() ([]Run, error) {
	rows, err := s.db.Query(`SELECT id,status,input,output,error FROM runs WHERE status='RUNNING' ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Status, &r.Input, &r.Output, &r.Error); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type Checkpoint struct{ ID, RunID, StepID, State string }

func (s *Store) LatestCheckpoint(runID string) (*Checkpoint, error) {
	var c Checkpoint
	err := s.db.QueryRow(`SELECT id,run_id,step_id,state FROM checkpoints WHERE run_id=? ORDER BY created_at DESC LIMIT 1`, runID).Scan(&c.ID, &c.RunID, &c.StepID, &c.State)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type ToolCallRecord struct {
	ID, RunID, StepID, ToolName, Arguments, Status, Result, Error, IdempotencyKey string
	Attempt, MaxAttempts                                                          int
	Retryable                                                                     bool
}

func (s *Store) GetToolCall(id string) (*ToolCallRecord, error) {
	var x ToolCallRecord
	var flag int
	err := s.db.QueryRow(`SELECT id,run_id,step_id,tool_name,arguments,status,result,error,attempt,max_attempts,idempotency_key,retryable FROM tool_calls WHERE id=?`, id).Scan(&x.ID, &x.RunID, &x.StepID, &x.ToolName, &x.Arguments, &x.Status, &x.Result, &x.Error, &x.Attempt, &x.MaxAttempts, &x.IdempotencyKey, &flag)
	if err != nil {
		return nil, err
	}
	x.Retryable = flag == 1
	return &x, nil
}
func (s *Store) MarkToolCallRetrying(id string) error {
	_, err := s.db.Exec(`UPDATE tool_calls SET status='RETRYING',attempt=attempt+1,updated_at=? WHERE id=?`, now(), id)
	return err
}
func (s *Store) MarkToolCallRunning(id string) error {
	_, err := s.db.Exec(`UPDATE tool_calls SET status='RUNNING',updated_at=? WHERE id=?`, now(), id)
	return err
}

var _ *sql.DB
