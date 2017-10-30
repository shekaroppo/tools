package mflib

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
)

type MfDbNotSet bool
type McUrlError bool
type MfDbOpenError struct {
	mfdb string
}

type NowHelper func() time.Time
type NavHelper func([]MutualFund) (map[MutualFund]float64, error)

var NavHelperFunc NavHelper = moneyControlNavHelper
var NowHelperFunc NowHelper = time.Now

type navResult struct {
	mf  MutualFund
	nav float64
	err error
}

func moneyControlNavHelper(mfs []MutualFund) (map[MutualFund]float64, error) {
	navFunc := func(mf MutualFund, ch chan navResult) {
		resp, err := http.Get(mf.mcUrl)
		if err != nil {
			ch <- navResult{mf, 0, err}
			return
		}
		defer resp.Body.Close()
		htmlBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ch <- navResult{mf, 0, err}
			return
		}
		regex, _ := regexp.Compile(`class="bd30tp">([\d,]+\.\d+)`)
		matchingBytes := regex.FindSubmatch(htmlBytes)
		if matchingBytes == nil {
			ch <- navResult{mf, 0, McUrlError(true)}
			return
		}
		stringNav := strings.Replace(string(matchingBytes[1]), ",", "", -1)
		floatNav, err := strconv.ParseFloat(stringNav, 64)
		if err != nil {
			ch <- navResult{mf, 0, errors.New("Error parsing number " + stringNav)}
			return
		}
		ch <- navResult{mf, floatNav, nil}
	}
	ch := make(chan navResult)
	for _, mf := range mfs {
		go navFunc(mf, ch)
	}
	mfToNav := make(map[MutualFund]float64)
	var errorStr string
	for i := 0; i < len(mfs); i++ {
		result := <-ch
		if result.err != nil {
			errorStr = result.err.Error() + "\n"
		} else {
			fmt.Println("NAV for", result.mf.name, "is", result.nav)
			mfToNav[result.mf] = result.nav
		}
	}

	var err error
	if errorStr != "" {
		err = errors.New(errorStr)
	}

	return mfToNav, err
}

type MutualFund struct {
	mfid           int
	name           string
	mcUrl          string
	folio          string
	mftype         string
	amfiSchemeCode string
}

type MutualFundPurchase struct {
	MutualFund
	id     int
	amount float64
	nav    float64
	time   time.Time
}

type MutualFundSummary struct {
	MutualFund
	avgDays            int
	amount             float64
	units              float64
	currentValue       float64
	appreciation       float64
	projectedYearlyRet float64
}

type mutualFundSummarySorter struct {
	mfsums []MutualFundSummary
	by     func(mfsum1, mfsum2 MutualFundSummary) bool
}

func sortByAppreciation(mfsum1 MutualFundSummary, mfsum2 MutualFundSummary) bool {
	if mfsum1.mfid == 0 {
		return false
	}
	return mfsum1.appreciation < mfsum2.appreciation
}

func (mfss *mutualFundSummarySorter) Len() int {
	return len(mfss.mfsums)
}

func (mfss *mutualFundSummarySorter) Swap(i, j int) {
	mfss.mfsums[i], mfss.mfsums[j] = mfss.mfsums[j], mfss.mfsums[i]
}

func (mfss *mutualFundSummarySorter) Less(i, j int) bool {
	return mfss.by(mfss.mfsums[i], mfss.mfsums[j])
}

func (m MfDbNotSet) Error() string {
	return "'MFDB' environment variable is not set"
}

func (m McUrlError) Error() string {
	return "Unable to parse moneycontrol url"
}

func GetDb() (*sql.DB, error) {
	mfDb := os.Getenv("MFDB")
	if mfDb == "" {
		return nil, MfDbNotSet(true)
	}

	db, err := sql.Open("sqlite3", mfDb)
	if err != nil {
		return nil, err
	}

	return db, err
}

func InitDb() error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create mutual fund table
	sqlStmt := `
	create table mutual_fund (
		mfid integer PRIMARY KEY AUTOINCREMENT,
		name text,
		mc_url text,
		folio text,
		type text,
		amfi_scheme_code text )`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}

	sqlStmt = `
	create table mutual_fund_purchase (
		id integer PRIMARY KEY AUTOINCREMENT,
		mfid integer,
		amount integer,
		nav real,
		time timestamp,
		foreign key(mfid) references mutual_fund(mfid));`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}

	mfDb := os.Getenv("MFDB")
	log.Println("Initialized mutual fund database at", mfDb)
	return nil
}

func listMutualFundHelper(id int) ([]MutualFund, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sqlStmt := `
	select mfid, name, mc_url, folio, type, amfi_scheme_code
	from mutual_fund`
	if id >= 0 {
		sqlStmt += fmt.Sprintf(" where mfid=%d", id)
	}
	sqlStmt += " order by mfid"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mutualFunds []MutualFund
	for rows.Next() {
		var mf MutualFund
		err = rows.Scan(&mf.mfid, &mf.name, &mf.mcUrl, &mf.folio, &mf.mftype, &mf.amfiSchemeCode)
		if err != nil {
			return nil, err
		}
		mutualFunds = append(mutualFunds, mf)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return mutualFunds, nil
}

func ListMutualFunds() ([]MutualFund, error) {
	return listMutualFundHelper(-1)
}

func GetMutualFunds(id int) (MutualFund, error) {
	var mf MutualFund
	mfs, err := listMutualFundHelper(id)
	if err != nil {
		return mf, err
	}
	if len(mfs) != 1 {
		return mf, errors.New("Mutual fund with id=" + strconv.Itoa(id) + " not found")
	}
	return mfs[0], nil
}

func InsertMutualFund(mfs ...MutualFund) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, mf := range mfs {
		sqlStmt, err := tx.Prepare(`
		insert into
		mutual_fund(name, mc_url, folio, type, amfi_scheme_code)
		values(?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer sqlStmt.Close()

		_, err = sqlStmt.Exec(mf.name, mf.mcUrl, mf.folio, mf.mftype, mf.amfiSchemeCode)
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func InsertMutualFundHelper(
	name string, mcUrl string, folio string,
	mftype string, amfiSchemeCode string) error {
	mf := MutualFund{0, name, mcUrl, folio, mftype, amfiSchemeCode}
	return InsertMutualFund(mf)
}

func RemoveMutualFundHelper(mfid int) error {
	var mf MutualFund
	mf.mfid = mfid
	return RemoveMutualFund(mf)
}

func ListMutualFundHelper() error {
	mfs, err := ListMutualFunds()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"MfId", "Name", "Folio", "Type", "Schemecode"})
	for _, mf := range mfs {
		table.Append([]string{strconv.Itoa(mf.mfid), mf.name, mf.folio, mf.mftype, mf.amfiSchemeCode})
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func RemoveMutualFund(mf MutualFund) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	sqlStmt, err := tx.Prepare(`
	delete from
	mutual_fund
	where mfid=?`)
	if err != nil {
		return err
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.Exec(mf.mfid)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func ListMutualFundPurchases(mftype string) ([]MutualFundPurchase, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sqlStmt := `
	select mutual_fund.mfid,
			 mutual_fund.name,
			 mutual_fund.mc_url,
		    mutual_fund.folio,
		    mutual_fund.type,
		    mutual_fund.amfi_scheme_code,
		    mutual_fund_purchase.id,
		    mutual_fund_purchase.amount,
		    mutual_fund_purchase.nav,
		    date(mutual_fund_purchase.time)
	from mutual_fund join mutual_fund_purchase
	where mutual_fund_purchase.mfid=mutual_fund.mfid `

	if mftype != "" {
		sqlStmt += " and mutual_fund.type='" + mftype + "' "
	}

	sqlStmt += "order by mutual_fund_purchase.id"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mfps []MutualFundPurchase
	for rows.Next() {
		var mfp MutualFundPurchase
		var mfpTime string
		err = rows.Scan(
			&mfp.mfid, &mfp.name, &mfp.mcUrl,
			&mfp.folio, &mfp.mftype, &mfp.amfiSchemeCode,
			&mfp.id, &mfp.amount, &mfp.nav, &mfpTime)
		if err != nil {
			return nil, err
		}
		mfp.time, err = time.Parse("2006-01-02", mfpTime)
		if err != nil {
			return nil, err
		}
		mfps = append(mfps, mfp)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return mfps, nil
}

func InsertMutualFundPurchase(mfps ...MutualFundPurchase) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, mfp := range mfps {
		sqlStmt, err := tx.Prepare(`
		insert into
		mutual_fund_purchase(mfid, amount, nav, time)
		values(?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer sqlStmt.Close()

		_, err = sqlStmt.Exec(
			mfp.mfid, mfp.amount, mfp.nav, mfp.time.Format("2006-01-02"))
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func InsertMutualFundPurchaseHelper(
	mfid int, amount float64, nav float64, dateStr string) error {
	var mfp MutualFundPurchase
	mf, err := GetMutualFunds(mfid)
	if err != nil {
		return err
	}
	mfp.MutualFund = mf
	mfp.amount = amount
	mfp.nav = nav
	mfp.time, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New(
			"Error parsing date '" + dateStr + "'. " +
				"Please specify in format YYYY-MM-DD")
	}
	return InsertMutualFundPurchase(mfp)
}

func RemoveMutualFundPurchaseHelper(mfpid int) error {
	var mfp MutualFundPurchase
	mfp.id = mfpid
	return RemoveMutualFundPurchase(mfp)
}

func ListMutualFundPurchaseHelper() error {
	mfps, err := ListMutualFundPurchases("")
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"MfpId", "Name", "Type", "Amount", "NAV", "Units", "Date"})
	for _, mfp := range mfps {
		amount := fmt.Sprintf("%.3f", mfp.amount)
		nav := fmt.Sprintf("%.3f", mfp.nav)
		units := fmt.Sprintf("%.3f", mfp.amount/mfp.nav)
		date := mfp.time.Format("2006-01-02")
		table.Append([]string{strconv.Itoa(mfp.id), mfp.name, mfp.mftype, amount, nav, units, date})
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func RemoveMutualFundPurchase(mfp MutualFundPurchase) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	sqlStmt, err := tx.Prepare(`
	delete from
	mutual_fund_purchase
	where id=?`)
	if err != nil {
		return err
	}
	defer sqlStmt.Close()

	_, err = sqlStmt.Exec(mfp.id)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func GetMutualFundSummary(mftype string) ([]MutualFundSummary, error) {
	mfps, err := ListMutualFundPurchases(mftype)
	if err != nil {
		return nil, err
	}

	mutualFundsMap := make(map[MutualFund]bool)
	for _, mfp := range mfps {
		mutualFundsMap[mfp.MutualFund] = true

	}

	var mutualFunds []MutualFund
	for mf, _ := range mutualFundsMap {
		mutualFunds = append(mutualFunds, mf)
	}

	latestNav, err := NavHelperFunc(mutualFunds)
	if err != nil {
		return nil, err
	}

	mfsMap := make(map[MutualFund]MutualFundSummary)
	sumAmountxDays := make(map[MutualFund]float64)
	sumAmount := make(map[MutualFund]float64)
	for _, mfp := range mfps {
		mf := mfp.MutualFund
		mfs := mfsMap[mf]
		mfs.MutualFund = mfp.MutualFund
		days := NowHelperFunc().Sub(mfp.time).Hours() / 24
		sumAmountxDays[mf] += mfp.amount * days
		sumAmount[mf] += mfp.amount
		mfs.amount += mfp.amount
		mfs.units += mfp.amount / mfp.nav
		mfsMap[mf] = mfs
	}

	var mfss []MutualFundSummary
	var totalSumAmountxDays, totalSumAmount, totalCurrentValue float64
	var totalAppreciation, totalProjectedYearlyRet float64
	for mf, _ := range mutualFundsMap {
		mfs := mfsMap[mf]
		mfs.avgDays = int(sumAmountxDays[mf] / sumAmount[mf])
		mfs.currentValue = latestNav[mf] * mfs.units
		mfs.appreciation = ((mfs.currentValue - mfs.amount) / mfs.amount) * 100
		appr := mfs.currentValue / mfs.amount
		power := 365 / float64(mfs.avgDays)
		mfs.projectedYearlyRet = (math.Pow(appr, power) - 1) * 100
		mfss = append(mfss, mfs)

		totalSumAmountxDays += sumAmountxDays[mf]
		totalSumAmount += sumAmount[mf]
		totalCurrentValue += mfs.currentValue
	}

	totalAppreciation = ((totalCurrentValue - totalSumAmount) / totalSumAmount) * 100
	totalAvgDays := int(totalSumAmountxDays / totalSumAmount)
	appr := totalCurrentValue / totalSumAmount
	power := 365 / float64(totalAvgDays)
	totalProjectedYearlyRet = (math.Pow(appr, power) - 1) * 100

	totalMf := MutualFund{0, "Total", "", "", "", ""}
	totalMfSum := MutualFundSummary{
		totalMf, totalAvgDays, totalSumAmount, 0,
		totalCurrentValue, totalAppreciation, totalProjectedYearlyRet}
	mfss = append(mfss, totalMfSum)

	return mfss, nil
}

func MutualFundSummaryHelper(mftype string) error {
	mfsums, err := GetMutualFundSummary(mftype)

	var mfssorter mutualFundSummarySorter
	mfssorter.mfsums = mfsums
	mfssorter.by = sortByAppreciation
	sort.Sort(&mfssorter)

	if err != nil {
		return err
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"MfId", "Name", "Type", "AvgDays",
		"Amount", "Units", "CurrentVal", "Appr", "PrjRet"})
	for _, mfsum := range mfsums {
		amount := fmt.Sprintf("%d", int(mfsum.amount))
		var units string
		if mfsum.units != 0 {
			units = fmt.Sprintf("%.2f", mfsum.units)
		} else {
		}
		currentValue := fmt.Sprintf("%d", int(mfsum.currentValue))
		appreciation := fmt.Sprintf("%.2f", mfsum.appreciation)
		projectedYearlyRet := fmt.Sprintf("%.2f", mfsum.projectedYearlyRet)
		var mfIdStr string
		if mfsum.mfid != 0 {
			mfIdStr = strconv.Itoa(mfsum.mfid)
		}
		table.Append([]string{
			mfIdStr, mfsum.name, mfsum.mftype,
			strconv.Itoa(mfsum.avgDays), amount, units, currentValue, appreciation,
			projectedYearlyRet})
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func MutualFundDisHelper() error {
	mfps, err := ListMutualFundPurchases("")
	if err != nil {
		return err
	}
	typeToAmount := make(map[string]float64)
	typeToPercentage := make(map[string]float64)
	var total float64
	for _, mfp := range mfps {
		typeToAmount[mfp.mftype] += mfp.amount
		total += mfp.amount
	}
	for _, mfp := range mfps {
		typeToPercentage[mfp.mftype] = (typeToAmount[mfp.mftype] * 100 / total)
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Type", "Amount", "Percentage"})

	var types []string
	for mftype, _ := range typeToAmount {
		types = append(types, mftype)
	}
	sort.Strings(types)

	for _, mftype := range types {
		amount := fmt.Sprintf("%.3f", typeToAmount[mftype])
		percentage := fmt.Sprintf("%.3f", typeToPercentage[mftype])
		table.Append([]string{mftype, amount, percentage})
	}
	table.Render()
	log.Print(buf.String())
	return nil
}
