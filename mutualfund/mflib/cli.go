package mflib

import (
	"errors"
	"log"
	"strconv"

	"github.com/urfave/cli"
)

type CommandArgumentError int

func (e CommandArgumentError) Error() string {
	return "This command exactly requires " + strconv.Itoa(int(e)) + " arguments"
}

func InitDbCliHelper(c *cli.Context) error {
	err := InitDb()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func InsertMfCliHelper(c *cli.Context) error {
	if c.NArg() != 5 {
		err := CommandArgumentError(5)
		log.Println(err)
		return err
	}
	err := InsertMutualFundHelper(
		c.Args().Get(0), c.Args().Get(1), c.Args().Get(2),
		c.Args().Get(3), c.Args().Get(4))
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("New mutual fund inserted successfully")
	return nil
}

func ListMfCliHelper(c *cli.Context) error {
	err := ListMutualFundHelper()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func RemoveMfCliHelper(c *cli.Context) error {
	if c.NArg() != 1 {
		err := CommandArgumentError(1)
		log.Println(err)
		return err
	}
	mfid, err := strconv.Atoi(c.Args().Get(0))
	if err != nil {
		msg := "Invalid argument '" + c.Args().Get(0) + "'"
		log.Println(msg)
		return err
	}
	err = RemoveMutualFundHelper(mfid)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Mutual fund removed successfully")
	return nil
}

func InsertMfpCliHelper(c *cli.Context) error {
	if c.NArg() != 4 {
		err := CommandArgumentError(5)
		log.Println(err)
		return err
	}
	mfidStr := c.Args().Get(0)
	mfid, err := strconv.Atoi(mfidStr)
	if err != nil {
		msg := "Invalid argument '" + mfidStr + "' for mutual fund id"
		log.Println(msg)
		return errors.New(msg)
	}
	amountStr := c.Args().Get(1)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		msg := "Invalid argument '" + amountStr + "' for amount"
		log.Println(msg)
		return errors.New(msg)
	}
	navStr := c.Args().Get(2)
	nav, err := strconv.ParseFloat(navStr, 64)
	if err != nil {
		msg := "Invalid argument '" + navStr + "' for NAV"
		log.Println(msg)
		return errors.New(msg)
	}
	dateStr := c.Args().Get(3)
	err = InsertMutualFundPurchaseHelper(mfid, amount, nav, dateStr)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("New mutual fund purchase inserted successfully")
	return nil
}

func ListMfpCliHelper(c *cli.Context) error {
	err := ListMutualFundPurchaseHelper()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func RemoveMfpCliHelper(c *cli.Context) error {
	if c.NArg() != 1 {
		err := CommandArgumentError(1)
		log.Println(err)
		return err
	}
	mfid, err := strconv.Atoi(c.Args().Get(0))
	if err != nil {
		msg := "Invalid argument '" + c.Args().Get(0) + "'"
		log.Println(msg)
		return err
	}
	err = RemoveMutualFundPurchaseHelper(mfid)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Mutual fund removed successfully")
	return nil
}

func DisCliHelper(c *cli.Context) error {
	err := MutualFundDisHelper()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SmfsCliHelper(c *cli.Context) error {
	err := MutualFundSummaryHelper(c.String("type"))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "mutualfund"
	app.Usage = "A tool to track mutual fund investments."
	app.Version = "1.0.0"
	app.UsageText =
		`
This is a tool to track mutual fund investments. The
information about the mutual funds and the investments made are
stored in an SQLite DB (a regular file).  The path to the
SQLite DB is to be mentioned by an environment variable 'MFDB'.

To begin, set MFDB to a path in your filesystem:
export MFDB='/home/user/mfdata.db'

Preferably, add this to your .bashrc.

Begin by initializing the database by:
mutualfund init

Insert/List/Remove mutual fund information by:
mutualfund (imf|lmf|rmf) <>

Insert/List/Remove mutual fund purchase information by:
mutualfund (imfp|lmfp|rmfp) <>

Summary of investments:
$ mutualfund (smfs|dis)
`

	insertMutualFundPurchaseUsage :=
		`
Insert a new mutual fund to database.

The arguments are
mutualfund imfp <mfid> <amount> <nav> <date>

mfid - Id of the mutual fund (use mutualfund lmf command to get this)
amount - Amount of investment made
nav - NAV at the time of investment
date - Date in the format YYYY-MM-DD
	`

	insertMutualFundUsage :=
		`
Insert a new mutual fund purchase to database.

The arguments are
mutualfund imf <name> <url> <folio> <type>

name - Name of the mutual fund to use while displaying it
url - Url to the moneycontrol page of mutual fund for grabbing NAV
amfi folio - Unused now (see later for usage)
type - Type of mutual fund (to filter while displaying)
	`

	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:   "init",
			Usage:  "Init the mutual fund DB",
			Action: InitDbCliHelper,
		},
		cli.Command{
			Name:      "insertMutualFund",
			Aliases:   []string{"imf"},
			Usage:     "Insert a new mutual fund",
			Action:    InsertMfCliHelper,
			UsageText: insertMutualFundUsage,
		},
		cli.Command{
			Name:    "listMutualFund",
			Aliases: []string{"lmf"},
			Usage:   "List all mutual funds",
			Action:  ListMfCliHelper,
		},
		cli.Command{
			Name:    "removeMutualFund",
			Aliases: []string{"rmf"},
			Usage:   "Remove an existing mutual funds",
			Action:  RemoveMfCliHelper,
		},
		cli.Command{
			Name:      "insertMutualFundPurchase",
			Aliases:   []string{"imfp"},
			Usage:     "Insert a new mutual fund purchase",
			Action:    InsertMfpCliHelper,
			UsageText: insertMutualFundPurchaseUsage,
		},
		cli.Command{
			Name:    "listMutualFundPurchase",
			Aliases: []string{"lmfp"},
			Usage:   "List all mutual fund purchase",
			Action:  ListMfpCliHelper,
		},
		cli.Command{
			Name:    "removeMutualFundPurchase",
			Aliases: []string{"rmfp"},
			Usage:   "Remove an existing mutual fund purchase",
			Action:  RemoveMfpCliHelper,
		},
		cli.Command{
			Name:    "showMutualFundSummary",
			Aliases: []string{"smfs"},
			Usage:   "Summary of all mutual fund investments",
			Action:  SmfsCliHelper,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Value: "",
					Usage: "Show summary for only this type",
				},
			},
		},
		cli.Command{
			Name:    "distribution",
			Aliases: []string{"dis"},
			Usage:   "Distribution of different types of mutual funds",
			Action:  DisCliHelper,
		},
	)
	return app
}
