package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/category"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "amicus",
		Usage: "Choose a person to contact and email the name to the recipient",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "budget-id",
				EnvVars: []string{"YNAB_BUDGET_ID"},
			},
			&cli.StringFlag{
				Name:    "access-token",
				EnvVars: []string{"YNAB_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:      "output",
				Aliases:   []string{"o"},
				TakesFile: true,
			},
			&cli.IntFlag{
				Name:  "year",
				Value: time.Now().Year() - 1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			return pivotData(cCtx)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func pivotData(cCtx *cli.Context) error {
	budgetID := cCtx.String("budget-id")
	accessToken := cCtx.String("access-token")

	if budgetID == "" || accessToken == "" {
		log.Fatal("budget-id and access-token are required")
	}

	ynabClient := ynab.NewClient(accessToken)

	bs, err := ynabClient.Budget().GetBudget(budgetID, nil)
	if err != nil {
		return err
	}

	log.Printf(bs.Budget.Name)

	cgMap := make(map[string]*category.Group)
	for _, cg := range bs.Budget.CategoryGroups {
		cgMap[cg.ID] = cg
	}

	outWriter := os.Stdout
	if cCtx.IsSet("output") {
		outFile := cCtx.String("output")
		outWriter, err = os.Create(outFile)
		if err != nil {
			return err
		}
		defer outWriter.Close()
	}
	w := log.New(outWriter, "", 0)

	for _, m := range bs.Budget.Months {
		if m.Month.Year() != cCtx.Int("year") {
			continue
		}
		for _, c := range m.Categories {
			if c.Deleted || c.Hidden {
				continue
			}
			cg := cgMap[c.CategoryGroupID]
			actDollars := float64(c.Activity) / 1000
			actStr := strconv.FormatFloat(actDollars, 'f', 2, 64)
			w.Printf("%s\t%s\t%s\t%s\n", api.DateFormat(m.Month), cg.Name, c.Name, actStr)
		}
	}

	return nil
}
