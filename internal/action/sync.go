package action

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/api"
	"thibaultleouay.dev/stargazers/internal/db"
)

func Sync(ctx context.Context, cmd *cli.Command) error {
	output := cmd.String("output")

	stargazersOnly := cmd.Bool("stargazers-only")
	fullSync := cmd.Bool("full")

	var database = db.New(output)

	client := api.NewClient(cmd.String("github-token"))

	repos := []struct {
		Id         int            `db:"id"`
		Name       string         `db:"name"`
		Owner      string         `db:"owner"`
		LastCursor sql.NullString `db:"last_cursor"`
	}{}

	err := database.Select(&repos, "SELECT id, name, owner, last_cursor FROM repository")
	if err != nil {
		return err
	}
	for _, repo := range repos {
		fmt.Fprintf(os.Stderr, "\nStart fetching stargazers for %s/%s\n", repo.Owner, repo.Name)

		cursor := ""
		if repo.LastCursor.Valid && !fullSync {
			cursor = repo.LastCursor.String
		}

		stargazers, endCursor, err := client.GetStargazers(repo.Owner, repo.Name, cursor)
		if err != nil {
			return err
		}

		for _, record := range stargazers {
			if record.Company != "" && !stargazersOnly {
				var companyId int64
				company := struct {
					Login string `db:"login"`
					Id    int64  `db:"id"`
				}{}

				err := database.Get(&company, "SELECT login, id FROM company WHERE login = $1", record.Company)

				if err != nil {
					r := database.MustExec("INSERT INTO company (login) VALUES ($1) ON CONFLICT(login) DO NOTHING", record.Company)
					companyId, _ = r.LastInsertId()
				} else {
					companyId = company.Id
				}

				database.MustExec(`INSERT INTO user(
						avatar_url, bio, email, followers_ct, following_ct,
						fullname, is_stargazer, login, company_id, linkedin_url, bsky_url
						) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
						) ON CONFLICT(login) DO UPDATE SET
						avatar_url = $1, bio = $2, email = $3, followers_ct = $4, following_ct = $5,
						fullname = $6, is_stargazer = $7, company_id = $9, linkedin_url = $10, bsky_url = $11
				`, record.AvatarUrl, record.Bio, record.Email, record.Followers, record.Following,
					record.Name, time.Now().Unix(), record.Login, companyId, record.LinkedinUrl, record.BskyUrl)

			} else {
				database.MustExec(`INSERT INTO user(
						avatar_url, bio, email, followers_ct, following_ct,
						fullname, is_stargazer, login, linkedin_url, bsky_url
						) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
						) ON CONFLICT(login) DO UPDATE SET
						avatar_url = $1, bio = $2, email = $3, followers_ct = $4, following_ct = $5,
						fullname = $6, is_stargazer = $7, linkedin_url = $9, bsky_url = $10
				`, record.AvatarUrl, record.Bio, record.Email, record.Followers, record.Following,
					record.Name, time.Now().Unix(), record.Login, record.LinkedinUrl, record.BskyUrl)
			}

			var userId int64
			err := database.Get(&userId, "SELECT id FROM user WHERE login = $1", record.Login)
			if err != nil {
				continue
			}
			database.MustExec("INSERT INTO users_to_repositories(user_id, repository_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userId, repo.Id)
		}

		database.MustExec("UPDATE repository SET last_cursor = $1, last_synced_at = $2 WHERE id = $3", endCursor, time.Now().Unix(), repo.Id)
	}
	if stargazersOnly {
		return nil
	}
	companies := []struct {
		Login string `db:"login"`
	}{}
	err = database.Select(&companies, "SELECT login FROM company")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "\nStart fetching company data")
	bar := progressbar.NewOptions(
		len(companies),
		progressbar.OptionSetDescription("Fetching Companies"),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(os.Stderr),
	)
	for _, company := range companies {
		bar.Add(1)
		login := strings.Trim(company.Login, " ")
		if strings.HasPrefix(login, "@") {
			login = strings.Trim(login, "@")
			apiCompany, err := client.GetCompany(login)
			if err == nil {
				database.MustExec(`UPDATE company SET
				avatar_url = $1,
				description = $2,
				email = $3,
				location = $4,
				name = $5,
				members_ct = $6,
				repositories_ct = $7,
				website_url = $8
				WHERE login = $9`,
					apiCompany.AvatarUrl,
					apiCompany.Description,
					apiCompany.Email,
					apiCompany.Location,
					apiCompany.Name,
					apiCompany.Members,
					apiCompany.Repositories,
					apiCompany.WebsiteUrl,
					company.Login,
				)
			}
		}
	}
	bar.Close()
	fmt.Fprintln(os.Stderr, "\nCompany data fetched")
	return nil
}
