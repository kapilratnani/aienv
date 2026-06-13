package audit

import "time"

type SessionMeta struct {
	ID           string    `json:"session_id"`
	EnvName      string    `json:"env_name"`
	AgentCommand string    `json:"agent_command"`
	StartedAt    time.Time `json:"started_at"`
}

type NetworkEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Method    string    `json:"method"`
	Allowed   bool      `json:"allowed"`
}
