// Package client provides API client implementations of shared contracts.
//
// Each domain client (UserClient, TaskClient, etc.) implements the
// corresponding contract interface from shared/contracts.
//
// Compile-time assertions ensure CLI clients always satisfy the shared contracts.
// Adding a method to any contract will break compilation here until implemented.
package client

import (
	"github.com/dezhishen/now-and-again/shared/contracts"
)

// ─── Compile-time interface compliance ────────────────────────────

var (
	_ contracts.UserContract         = (*UserClient)(nil)
	_ contracts.FamilyContract       = (*FamilyClient)(nil)
	_ contracts.SubGroupContract     = (*SubGroupClient)(nil)
	_ contracts.TaskContract         = (*TaskClient)(nil)
	_ contracts.ChainContract        = (*ChainClient)(nil)
	_ contracts.InspectionContract   = (*InspectionClient)(nil)
	_ contracts.LogContract          = (*LogClient)(nil)
	_ contracts.NotificationContract = (*NotificationClient)(nil)
	_ contracts.ApiKeyContract       = (*ApiKeyClient)(nil)
)

// AllClients bundles all domain clients for injection into CLI commands.
type AllClients struct {
	User         *UserClient
	Family       *FamilyClient
	SubGroup     *SubGroupClient
	Task         *TaskClient
	Chain        *ChainClient
	Inspection   *InspectionClient
	Log          *LogClient
	Notification *NotificationClient
	ApiKey       *ApiKeyClient
}

// NewAllClients creates all domain clients sharing a single HTTP transport.
func NewAllClients(baseURL, token string) *AllClients {
	httpClient := New(baseURL, token)
	return &AllClients{
		User:         &UserClient{http: httpClient},
		Family:       &FamilyClient{http: httpClient},
		SubGroup:     &SubGroupClient{http: httpClient},
		Task:         &TaskClient{http: httpClient},
		Chain:        &ChainClient{http: httpClient},
		Inspection:   &InspectionClient{http: httpClient},
		Log:          &LogClient{http: httpClient},
		Notification: &NotificationClient{http: httpClient},
		ApiKey:       &ApiKeyClient{http: httpClient},
	}
}

// ─── Domain client structs ────────────────────────────────────────

type UserClient struct{ http *Client }
type FamilyClient struct{ http *Client }
type SubGroupClient struct{ http *Client }
type TaskClient struct{ http *Client }
type ChainClient struct{ http *Client }
type InspectionClient struct{ http *Client }
type LogClient struct{ http *Client }
type NotificationClient struct{ http *Client }
type ApiKeyClient struct{ http *Client }
