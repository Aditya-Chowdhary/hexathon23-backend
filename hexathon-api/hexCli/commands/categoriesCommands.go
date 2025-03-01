package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/GDGVIT/hexathon23-backend/hexathon-api/internal/models"
	"github.com/urfave/cli/v2"
)

// categoriesCommands is the list of commands related to categories
var categoriesCommands = []*cli.Command{
	{
		Name:    "load-categories",
		Aliases: []string{"lc"},
		Usage:   "Loads categories from the csv",
		Action:  loadCategories,
	},
}

// Index in csv file for each column; iota auto increments
const (
	category_name int = iota
	category_photo_url
	category_description
	category_maxitems
)

func loadCategories(c *cli.Context) error {
	path := c.Args().Get(0)

	if path == "" {
		fmt.Println("Please provide a path")
		fmt.Scanln(&path)
	}

	// Read the file to load data from
	fd, err := os.Open(path)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fileReader := csv.NewReader(fd)
	records, err := fileReader.ReadAll()

	// remove header
	fmt.Println("Discarding header: ", records[0])
	records = records[1:]

	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, record := range records {
		// // Check if participant already exists
		// if models.CheckParticipantExists(record[user_reg_no]) {
		// 	fmt.Println("Participant already exists")
		// 	continue
		// }

		maxItems, err := strconv.Atoi(record[category_maxitems])
		if err != nil {
			fmt.Printf("Invalid max items for %s: %s", record[category_name], record[category_maxitems])
		}

		// Each record is made into a category object
		category := &models.Category{
			Name:        record[category_name],
			PhotoURL:    record[category_photo_url],
			Description: record[category_description],
			MaxItems:    maxItems,
		}
		// Writing the object to database
		err = category.CreateCategory()

		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Load successfull")
	return nil
}
