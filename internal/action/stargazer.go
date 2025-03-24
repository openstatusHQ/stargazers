package action

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/api"
)

func Stargazers(ctx context.Context, cmd *cli.Command) error {
	owner := cmd.String("owner")
	name := cmd.String("name")

	fmt.Printf("Fetching Stargazers for %s/%s\n", owner, name)

	client := api.NewClient(cmd.String("github-token"))
	stargazers, err := client.GetStargazers(owner, name)
	if err != nil {
		return err
	}

	err = WriteToCsv(stargazers, cmd.String("output"))
	if err != nil {
		return err
	}

	fmt.Printf("Stargazers saved to %s", cmd.String("output"))
	return nil
}

func WriteToCsv(records []api.Stargazer, filename string) error {
	var newFilename string
	if filename == "" {
		newFilename = "stargazers.csv"
	} else {
		newFilename = filename
	}
	fmt.Printf("Writing to %s\n", newFilename)
	file, err := os.Create(newFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{"AvatarUrl", "Bio", "Company", "Email", "Followers", "Following", "Login", "Name"})
	for _, record := range records {

		err := writer.Write([]string{
			record.AvatarUrl,
			record.Bio,
			strings.Trim(record.Company, " "),
			record.Email,
			fmt.Sprintf("%d", record.Following),
			fmt.Sprintf("%d", record.Followers),
			record.Login, record.Name,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
