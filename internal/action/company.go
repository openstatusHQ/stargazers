package action

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/api"
)

func Company(ctx context.Context, cmd *cli.Command) error {
	filename := cmd.String("input")
	fmt.Printf("Fetching Company for %s\n", filename)
	stargazers, err := ReadFromCsv(filename)
	if err != nil {
		return err
	}
	client := api.NewClient(cmd.String("github-token"))

	output := cmd.String("output")
	if output == "" {
		output = "companies.csv"
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"User_AvatarUrl",
		"User_Bio",
		"User_Company",
		"User_Email",
		"User_Followers",
		"User_Following",
		"User_Login",
		"User_Name",
		"Org_AvatarUrl",
		"Org_Description",
		"Org_Email",
		"Org_Location",
		"Org_Login",
		"Org_Name",
		"Org_Members",
		"Org_Repositories",
		"Org_WebsiteUrl"})

	companies := make(map[string]*api.Company)
	for _, stargazer := range stargazers {
		if stargazer.Company != "" {
			if _, ok := companies[stargazer.Company]; !ok {
				login := strings.TrimPrefix(stargazer.Company, "@")
				login = strings.Trim(login, " ")
				company, err := client.GetCompany(login)
				if err != nil {
					fmt.Println("Company does not exist")
				} else {

					companies[stargazer.Company] = company
				}
			}
		}
	}

	// Header
	for _, record := range stargazers {
		if record.Company != "" {
			if c, ok := companies[record.Company]; ok {
				err = writer.Write([]string{
					record.AvatarUrl,
					record.Bio,
					record.Company,
					record.Email,
					fmt.Sprintf("%d", record.Following),
					fmt.Sprintf("%d", record.Followers),
					record.Login, record.Name,
					c.AvatarUrl,
					c.Description,
					c.Email,
					c.Location,
					c.Login,
					c.Name,
					fmt.Sprintf("%d", c.Members),
					fmt.Sprintf("%d", c.Repositories),
					c.WebsiteUrl,
				})
			} else {
				err = writer.Write([]string{
					record.AvatarUrl,
					record.Bio,
					record.Company,
					record.Email,
					fmt.Sprintf("%d", record.Following),
					fmt.Sprintf("%d", record.Followers),
					record.Login, record.Name,
				})
				if err != nil {
					return err
				}
			}
		} else {

			err = writer.Write([]string{
				record.AvatarUrl,
				record.Bio,
				record.Company,
				record.Email,
				fmt.Sprintf("%d", record.Following),
				fmt.Sprintf("%d", record.Followers),
				record.Login, record.Name,
			})
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func ReadFromCsv(filename string) ([]api.Stargazer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := csv.NewReader(file)

	if _, err := r.Read(); err != nil {
		return nil, err
	}
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var stargazers []api.Stargazer
	for _, record := range records {
		following, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, err
		}
		followers, err := strconv.Atoi(record[5])
		if err != nil {
			return nil, err
		}
		stargazers = append(stargazers, api.Stargazer{
			AvatarUrl: record[0],
			Bio:       record[1],
			Company:   record[2],
			Email:     record[3],
			Following: following,
			Followers: followers,
			Login:     record[6],
			Name:      record[7],
		})
	}
	return stargazers, nil
}
