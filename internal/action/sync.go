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

	var database = db.New(output)

	client := api.NewClient(cmd.String("github-token"))

	repos := []struct {
		Id    int    `db:"id"`
		Name  string `db:"name"`
		Owner string `db:"owner"`
	}{}

	err := database.Select(&repos, "SELECT id, name, owner FROM repository")
	if err != nil {
		return err
	}
	for _, repo := range repos {
		fmt.Fprintf(os.Stderr, "\nStart fetching stargazers for %s/%s\n",repo.Owner, repo.Name)
		stargazers, err := client.GetStargazers(repo.Owner, repo.Name)
		if err != nil {
			return err
		}

		for _, record := range stargazers {
			var res sql.Result
			if record.Company != "" {
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

				res = database.MustExec(`INSERT INTO user(
							avatar_url,
				 			bio,
							email,
							followers_ct,
							following_ct,
							fullname,
							is_stargazer,
							login,
							company_id
							) VALUES (
							$1, $2, $3, $4, $5, $6, $7, $8, $9
							) ON CONFLICT(login) DO NOTHING
					`, record.AvatarUrl, record.Bio, record.Email, record.Followers, record.Following,record.Name, time.Now().Unix(),  record.Login, companyId)

			} else {
				res = database.MustExec(`INSERT INTO user(
							avatar_url,
				 			bio,
							email,
							followers_ct,
							following_ct,
							fullname,
							is_stargazer,
							login
							) VALUES (
							$1, $2, $3, $4, $5, $6, $7, $8
							) ON CONFLICT(login) DO NOTHING
					`, record.AvatarUrl, record.Bio, record.Email, record.Followers, record.Following,  record.Name, time.Now().Unix(), record.Login)
			}
			id, _ := res.LastInsertId()
			database.MustExec("INSERT INTO users_to_repositories(user_id, repository_id) VALUES ($1, $2) ", id, repo.Id)
		}
	}
	companies := []struct {
		Login string `db:"login"`
	}{}
	err = database.Select(&companies, "SELECT login FROM company")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "\nStart fetching company data" )
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
		// Silent error

	}
	bar.Close()
	fmt.Fprintln(os.Stderr, "\nCompany data fetched" )
	return nil
}
