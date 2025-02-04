package stargazers

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

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

	err = WriteToCsv(stargazers, cmd.String("file"))
	if err != nil {
		return err
	}

	fmt.Printf("Stargazers saved to %s", cmd.String("file"))
	return nil
}

func WriteToCsv(records []api.Stargazer, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Name", "Login", "Email", "Company"})
	for _, record := range records {
		err := writer.Write([]string{record.Name, record.Login, record.Email, record.Company})
		if err != nil {
			return err
		}
	}
	return nil
}
