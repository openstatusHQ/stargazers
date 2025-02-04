package api

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client *githubv4.Client
}

func NewClient(token string) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	return &Client{
		client: client,
	}
}

type Stargazer struct {
	Name    string
	Login   string
	Email   string
	Company string
}

func (c *Client) GetStargazers(owner string, name string) ([]Stargazer, error) {
	var query struct {
		Repository struct {
			Name string

			Stargazers struct {
				TotalCount int
				PageInfo   struct {
					EndCursor   githubv4.String
					HasNextPage githubv4.Boolean
				}
				Nodes []struct {
					Name    string
					Login   string
					Email   string
					Company string
				}
			} `graphql:"stargazers(first: $first, after: $after)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	var stargazers []Stargazer
	for {
		err := c.client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}

		for _, stargarer := range query.Repository.Stargazers.Nodes {
			stargazers = append(stargazers,
				Stargazer{
					Name:    stargarer.Name,
					Login:   stargarer.Login,
					Company: stargarer.Company,
					Email:   stargarer.Email,
				})
		}
		if !query.Repository.Stargazers.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.NewString(query.Repository.Stargazers.PageInfo.EndCursor)
	}
	return stargazers, nil
}
