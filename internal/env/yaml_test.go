package env

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     Env
		wantErr bool
	}{
		{"empty name", Env{Agent: "opencode"}, true},
		{"empty agent", Env{Name: "test"}, true},
		{"unsupported agent", Env{Name: "test", Agent: "cursor"}, true},
		{"valid opencode", Env{Name: "test", Agent: "opencode"}, false},
		{"valid claude-code", Env{Name: "test", Agent: "claude-code"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
