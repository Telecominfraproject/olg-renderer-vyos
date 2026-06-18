package normalize

import (
	"encoding/json"
	"strings"
	"testing"
)

/*
TC-INTERFACE-NORMALIZE-001
Type: Mixed
Title: VLAN ID range validation
Summary:
Verifies that downstream VLAN configurations are checked for valid range boundaries.
It ensures valid IDs (e.g. 1 and 4094) are accepted, and out-of-bounds IDs
(e.g. 0, 4095, negative values) are rejected with CodeNormalizeFailed.

Validates:
  - Boundaries 1 and 4094 are successfully accepted
  - Out of bounds values like 0 and 4095 fail normalization
  - Negative VLAN IDs fail normalization
*/
func TestNormalizeInterfacesVLANRange(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vlan ids (lower boundary)",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-1", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 1}}
				]
			}`,
			wantErr: false,
		},
		{
			name: "valid vlan ids (upper boundary)",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-4094", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 4094}}
				]
			}`,
			wantErr: false,
		},
		{
			name: "invalid vlan id (0)",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-0", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 0}}
				]
			}`,
			wantErr: true,
			errMsg:  "out of bounds (must be 1-4094)",
		},
		{
			name: "invalid vlan id (4095)",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-4095", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 4095}}
				]
			}`,
			wantErr: true,
			errMsg:  "out of bounds (must be 1-4094)",
		},
		{
			name: "invalid vlan id negative",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-neg", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": -10}}
				]
			}`,
			wantErr: true,
			errMsg:  "out of bounds (must be 1-4094)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var root map[string]json.RawMessage
			if err := json.Unmarshal([]byte(tc.payload), &root); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}
			_, err := normalizeInterfaces(root, map[string][]string{"LAN1": {"eth1"}})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				normErr, ok := err.(*Error)
				if !ok {
					t.Fatalf("expected error to be of type *Error, got %T", err)
				}
				if normErr.Code != CodeNormalizeFailed {
					t.Errorf("expected error code %q, got %q", CodeNormalizeFailed, normErr.Code)
				}
				if tc.errMsg != "" && !strings.Contains(normErr.Error(), tc.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tc.errMsg, normErr.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

/*
TC-INTERFACE-NORMALIZE-002
Type: Mixed
Title: Duplicate VLAN ID detection
Summary:
Verifies that duplicate VLAN IDs across downstream interfaces are detected and rejected.

Validates:
  - Non-overlapping downstream VLAN configurations are successfully accepted
  - Duplicate VLAN IDs on downstream interfaces fail normalization with CodeNormalizeFailed
*/
func TestNormalizeInterfacesDuplicateVLANID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr bool
		errMsg  string
	}{
		{
			name: "no duplicate vlan ids",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-10", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 10}},
					{"name": "VLAN-20", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.20.1/24"}, "vlan": {"id": 20}}
				]
			}`,
			wantErr: false,
		},
		{
			name: "duplicate vlan ids on downstream interfaces",
			payload: `{
				"interfaces": [
					{"name": "LAN", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.1.1/24"}},
					{"name": "VLAN-10-A", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.10.1/24"}, "vlan": {"id": 10}},
					{"name": "VLAN-10-B", "role": "downstream", "ethernet": [{"select-ports": ["LAN1"]}], "ipv4": {"addressing": "static", "subnet": "192.168.20.1/24"}, "vlan": {"id": 10}}
				]
			}`,
			wantErr: true,
			errMsg:  "duplicate vlan id 10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var root map[string]json.RawMessage
			if err := json.Unmarshal([]byte(tc.payload), &root); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}
			_, err := normalizeInterfaces(root, map[string][]string{"LAN1": {"eth1"}})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				normErr, ok := err.(*Error)
				if !ok {
					t.Fatalf("expected error to be of type *Error, got %T", err)
				}
				if normErr.Code != CodeNormalizeFailed {
					t.Errorf("expected error code %q, got %q", CodeNormalizeFailed, normErr.Code)
				}
				if tc.errMsg != "" && !strings.Contains(normErr.Error(), tc.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tc.errMsg, normErr.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
