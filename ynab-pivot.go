package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/category"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "amicus",
		Usage: "Choose a person to contact and email the name to the recipient",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "budget-id",
				Sources: cli.EnvVars("YNAB_BUDGET_ID"),
			},
			&cli.StringFlag{
				Name:    "access-token",
				Sources: cli.EnvVars("YNAB_ACCESS_TOKEN"),
			},
			&cli.StringFlag{
				Name:      "output",
				Aliases:   []string{"o"},
				TakesFile: true,
			},
			&cli.IntFlag{
				Name:  "year",
				Value: int64(time.Now().Year() - 1),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return pivotData(ctx, cmd)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func pivotData(_ context.Context, cmd *cli.Command) error {
	budgetID := cmd.String("budget-id")
	accessToken := cmd.String("access-token")

	if budgetID == "" || accessToken == "" {
		log.Fatal("--budget-id and --access-token are required")
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
	if cmd.IsSet("output") {
		outFile := cmd.String("output")
		outWriter, err = os.Create(outFile)
		if err != nil {
			return err
		}
		defer outWriter.Close()
	}
	w := log.New(outWriter, "", 0)

	for _, m := range bs.Budget.Months {
		if int64(m.Month.Year()) != cmd.Int("year") {
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
