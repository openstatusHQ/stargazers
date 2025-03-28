package api

import (
	"context"
	"os"
	"time"

	// "os"

	// "github.com/schollz/progressbar/v3"
	"github.com/schollz/progressbar/v3"
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

type User struct {
	AvatarUrl string `db:"avatar_url"`
	Bio       string `db:"bio"`
	Company   string
	Email     string `db:"email"`
	Followers int    `db:"followers_ct"`
	Following int    `db:"following_ct"`
	Login     string `db:"login"`
	Name      string `db:"fullname"`
}

type Company struct {
	AvatarUrl    string `db:"avatar_url"`
	Description  string `db:"description"`
	Email        string `db:"email"`
	Location     string `db:"location"`
	Login        string `db:"login"`
	Name         string `db:"name"`
	Members      int    `db:"members_ct"`
	Repositories int    `db:"repositories_ct"`
	WebsiteUrl   string `db:"website_url"`
}

func (c *Client) GetStargazers(owner string, name string) ([]User, error) {
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

	var stargazers []User
	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetDescription("Fetching Stargazers"),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
	)
	for {
		err := c.client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}

		for _, stargarer := range query.Repository.Stargazers.Nodes {
			stargazers = append(stargazers,
				User{
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
		// Let's display the progress bar
		bar.ChangeMax(query.Repository.Stargazers.TotalCount)
		bar.Add(len(query.Repository.Stargazers.Nodes))

		if !query.Repository.Stargazers.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Repository.Stargazers.PageInfo.EndCursor)
	}
	bar.Close()

	return stargazers, nil
}

func (c *Client) GetCompany(login string) (*Company, error) {
	var query struct {
		Organization struct {
			Name            string
			Login           string
			AvatarUrl       string
			Description     string
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
		Description:  query.Organization.Description,
		Email:        query.Organization.Email,
		Location:     query.Organization.Location,
		Login:        query.Organization.Login,
		Name:         query.Organization.Name,
		Members:      query.Organization.MembersWithRole.TotalCount,
		Repositories: query.Organization.Repositories.TotalCount,
		WebsiteUrl:   query.Organization.WebsiteUrl,
	}, nil
}
