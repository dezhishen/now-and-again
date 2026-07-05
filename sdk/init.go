package sdk

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

func (na *NA) Init(username, password string) error {
	ctx := context.Background()
	pair, err := na.User.Login(ctx, &types.LoginRequest{Username: username, Password: password})
	if err != nil {
		return fmt.Errorf("init: login: %w", err)
	}
	na.http.SetToken(pair.AccessToken)
	apiKey, err := na.ensureAPIKey(ctx)
	if err != nil {
		return fmt.Errorf("init: api key: %w", err)
	}
	na.SetToken(apiKey)
	families, err := na.Family.ListMyFamilies(ctx)
	if err != nil {
		return fmt.Errorf("init: list families: %w", err)
	}
	if len(families) > 0 {
		na.SetActiveFamily(families[0].ID, families[0].Name)
	}
	return na.Config().Save()
}

func (na *NA) InitWithKey(apiKey string) error {
	na.SetToken(apiKey)
	ctx := context.Background()
	families, err := na.Family.ListMyFamilies(ctx)
	if err != nil {
		return fmt.Errorf("init: list families: %w", err)
	}
	if len(families) > 0 {
		na.SetActiveFamily(families[0].ID, families[0].Name)
	}
	return na.Config().Save()
}

func (na *NA) IsInitialized() bool {
	return na.GetToken() != "" && na.ActiveFamilyID() != ""
}

func (na *NA) ensureAPIKey(ctx context.Context) (string, error) {
	keys, err := na.ApiKey.List(ctx)
	if err != nil {
		return "", err
	}
	for _, k := range keys {
		if k.RawKey != "" {
			return k.RawKey, nil
		}
	}
	resp, err := na.ApiKey.Create(ctx, &types.CreateApiKeyRequest{
		Name: "CLI SDK", Scopes: []string{"read", "write"},
	})
	if err != nil {
		return "", fmt.Errorf("create api key: %w", err)
	}
	if resp.ApiKey != nil && resp.ApiKey.RawKey != "" {
		return resp.ApiKey.RawKey, nil
	}
	return "", fmt.Errorf("api key created; use 'na init --key <key>' to configure manually")
}
