package udemyapi

import "context"

type page[T any] struct {
	Next    string `json:"next"`
	Results []T    `json:"results"`
}

func firstPage[T any](ctx context.Context, client *Client, target string) ([]T, error) {
	var current page[T]
	if err := client.getJSON(ctx, target, &current); err != nil {
		return nil, err
	}
	return current.Results, nil
}

func allPages[T any](ctx context.Context, client *Client, target string) ([]T, error) {
	var results []T
	for next := target; next != ""; {
		var current page[T]
		if err := client.getJSON(ctx, next, &current); err != nil {
			return nil, err
		}
		results = append(results, current.Results...)
		next = current.Next
	}
	return results, nil
}
