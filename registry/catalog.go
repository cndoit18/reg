package registry

import (
	"context"
	"net/url"
	"strings"

	"github.com/peterhellberg/link"
)

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

// Catalog returns the repositories in a registry.
func (r *Registry) Catalog(ctx context.Context, u string) ([]string, error) {
	if u == "" {
		u = "/v2/_catalog"
	}
	uri := r.url(u)
	r.Logf("registry.catalog url=%s", uri)

	var response catalogResponse
	h, err := r.getJSON(ctx, uri, &response)
	if err != nil {
		return nil, err
	}

	o, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	last := o.Query().Get("last")
	for i := 0; i < len(response.Repositories); i++ {
		if !strings.HasPrefix(response.Repositories[i], last) {
			response.Repositories = append([]string{}, response.Repositories[:i]...)
			return response.Repositories, nil
		}
	}

	for _, l := range link.ParseHeader(h) {
		if l.Rel == "next" {
			unescaped, _ := url.QueryUnescape(l.URI)
			repos, err := r.Catalog(ctx, unescaped)
			if err != nil {
				return nil, err
			}
			response.Repositories = append(response.Repositories, repos...)
		}
	}

	return response.Repositories, nil
}
