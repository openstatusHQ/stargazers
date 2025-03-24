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
	AvatarUrl string
	Bio       string
	Company   string
	Email     string
	Followers int
	Following int
	Login     string
	Name      string
}

type Company struct {
	AvatarUrl    string
	Description   string
	Email        string
	Location     string
	Login        string
	Name         string
	Members      int
	Repositories int
	WebsiteUrl   string
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
					AvatarUrl string
					Bio       string
					Company   string
					Email     string
					Followers struct {
						TotalCount int
					}
					Following struct {
						TotalCount int
					}
					Login string
					Name  string
				}
			} `graphql:"stargazers(first: $first, after: $after)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
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
					AvatarUrl: stargarer.AvatarUrl,
					Bio:       stargarer.Bio,
					Company:   stargarer.Company,
					Email:     stargarer.Email,
					Following: stargarer.Following.TotalCount,
					Followers: stargarer.Followers.TotalCount,
					Login:     stargarer.Login,
					Name:      stargarer.Name,
				})
		}
		if !query.Repository.Stargazers.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.NewString(query.Repository.Stargazers.PageInfo.EndCursor)
	}
	return stargazers, nil
}

func (c *Client) GetCompany(login string) (*Company, error) {
	var query struct {
		Organization struct {
			Name            string
			Login           string
			AvatarUrl       string
			Description      string
			Email           string
			Location        string
			MembersWithRole struct {
				TotalCount int
			}
			Repositories struct {
				TotalCount int
			}
			WebsiteUrl string
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]any{
		"login": githubv4.String(login),
	}

	err := c.client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return &Company{
		AvatarUrl:    query.Organization.AvatarUrl,
		Description:   query.Organization.Description,
		Email:        query.Organization.Email,
		Location:     query.Organization.Location,
		Login:        query.Organization.Login,
		Name:         query.Organization.Name,
		Members:      query.Organization.MembersWithRole.TotalCount,
		Repositories: query.Organization.Repositories.TotalCount,
		WebsiteUrl:   query.Organization.WebsiteUrl,
	}, nil
}
