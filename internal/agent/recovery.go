package agent

import (
	"context"
	"fmt"
	"local-llm-lab/internal/db"
	"local-llm-lab/internal/tools"
)

type RecoveryManager struct {
	store *db.Store
	tools *tools.Registry
	agent *Agent
}

func NewRecoveryManager(store *db.Store, registry *tools.Registry, agent ...*Agent) *RecoveryManager {
	r := &RecoveryManager{store: store, tools: registry}
	if len(agent) > 0 {
		r.agent = agent[0]
	}
	return r
}

func (r *RecoveryManager) Recover(ctx context.Context) error {
	runs, err := r.store.FindRunningRuns()
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		return nil
	}
	if r.agent == nil {
		return fmt.Errorf("recovery agent is required for RUNNING runs")
	}
	for _, run := range runs {
		if _, err := r.agent.Resume(ctx, run.ID); err != nil {
			fmt.Printf("[recovery] run=%s error=%v\n", run.ID, err)
		}
	}
	return nil
}
