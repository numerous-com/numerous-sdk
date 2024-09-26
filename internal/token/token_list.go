package token

import (
	"context"
	"time"
)

type TokenEntry struct {
	ID          string
	Name        string
	Description string
	ExpiresAt   *time.Time
}

type ListTokenOutput []TokenEntry

type personalAccessTokenListResponse struct {
	Me struct {
		PersonalAccessTokens []struct {
			ID          string
			Name        string
			Description string
			ExpiresAt   *time.Time
		}
	}
}

func (s *Service) List(ctx context.Context) (ListTokenOutput, error) {
	var resp personalAccessTokenListResponse

	if err := s.client.Query(ctx, &resp, map[string]interface{}{}); err != nil {
		return ListTokenOutput{}, ConvertErrors(err)
	} else {
		result := make(ListTokenOutput, len(resp.Me.PersonalAccessTokens))
		for i, entry := range resp.Me.PersonalAccessTokens {
			result[i] = TokenEntry{
				ID:          entry.ID,
				Name:        entry.Name,
				Description: entry.Description,
				ExpiresAt:   entry.ExpiresAt,
			}
		}

		return result, nil
	}
}
