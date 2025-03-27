package action

import (
	"context"

	"github.com/rodaine/table"
	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/internal/db"
)

func RepoView(ctx context.Context, cmd *cli.Command) error {
	output := cmd.String("output")

	database  := db.New(output)

	repos := []struct {
		Id    int    `db:"id"`
		Name  string `db:"name"`
		Owner string `db:"owner"`
	}{}

	err := database.Select(&repos, "SELECT id, name, owner FROM repository")
	if err != nil {
		return err
	}
	tbl := table.New("ID", "Owner", "Name")
	for _, repo := range repos {
		tbl.AddRow(repo.Id, repo.Owner, repo.Name)
	}
	tbl.Print()
	return nil
}
