package action

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/api"
	"thibaultleouay.dev/stargazers/internal/db"
)

func Sync(ctx context.Context, cmd *cli.Command) error {
	var database = db.New()

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
		stargazers, err := client.GetStargazers(repo.Owner, repo.Name)
		if err != nil {
			return err
		}

		tx := database.MustBegin()
		for _, record := range stargazers {
			var res sql.Result
			if record.Company != "" {
				r := tx.MustExec("INSERT INTO company (login) VALUES ($1) ON CONFLICT(login) DO NOTHING", record.Company)
				id, _ := r.LastInsertId()
				res = tx.MustExec(`INSERT INTO user(
							avatar_url,
				 			bio,
							email,
							followers_ct,
							following_ct,
							fullname,
							login,
							company_id) VALUES (
							$1, $2, $3, $4, $5, $6, $7, $8
							) ON CONFLICT(login) DO NOTHING
					`, record.AvatarUrl, record.Bio, record.Email, record.Followers, record.Following, record.Name, record.Login, id)

			} else {
				res, _ = tx.NamedExec(`INSERT INTO user
				(avatar_url, bio, email, followers_ct, following_ct, fullname, login) VALUES
			 	(:avatar_url, :bio, :email, :followers_ct, :following_ct, :fullname, :login)`, record)
			}

			id, _ := res.LastInsertId()
			tx.MustExec("INSERT INTO users_to_repositories(user_id, repository_id) VALUES ($1, $2) ", id, repo.Id)
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		tx = database.MustBegin()
		companies := []struct {
			Login string `db:"login"`
		}{}
		err = database.Select(&companies, "SELECT login FROM company")
		if err != nil {
			return err
		}
		for _, company := range companies {
			fmt.Println(company.Login)
			login := strings.Trim(company.Login, " ")
			if strings.HasPrefix(login, "@") {
				login = strings.Trim(login, "@")
				apiCompany, err := client.GetCompany(login)
				fmt.Println(apiCompany)
				if err == nil {
					tx.MustExec(`UPDATE company SET
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
		tx.Commit()
	}
	return nil
}
