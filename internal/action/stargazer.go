package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/api"
)

func Stargazers(ctx context.Context, cmd *cli.Command) error {
	owner := cmd.String("owner")
	name := cmd.String("name")

	client := api.NewClient(cmd.String("github-token"))
	stargazers, err := client.GetStargazers(owner, name)
	if err != nil {
		return err
	}

	fmt.Println("AvatarUrl,Bio,Company,Email,Followers,Following,Login,Name")
	for _, record := range stargazers {
		fmt.Println(record.AvatarUrl,
			record.Bio,
			strings.Trim(record.Company, " "),
			record.Email,
			fmt.Sprintf("%d", record.Following),
			fmt.Sprintf("%d", record.Followers),
			record.Login, record.Name)
	}

	return nil
}
