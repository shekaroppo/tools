package mflib

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
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

type navResult struct {
	nav float64
	err error
}

/*
func moneyControlNavHelper(mfs []MutualFund) (map[MutualFund]float64, error) {
	navFunc := func(mf MutualFund) navResult {
		resp, err := http.Get(mf.mcUrl)
		if err != nil {
			return navResult{0, err}
		}
		defer resp.Body.Close()
		htmlBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return navResult{0, err}
		}
			regex = regexp.Compile(`class="bd30tp">(\d+\\.\d+)</span>`)
			matchingBytes := re.Find(htmlBytes)
			if matchingBytes == nil {
				return McUrlError(true)
			}

	}
	for _, mf := range mfs {
	}
}
*/

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

func sortByMfId(mfsum1 MutualFundSummary, mfsum2 MutualFundSummary) bool {
	return mfsum1.mfid < mfsum2.mfid
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
		foreign key(mfid) references mutual_fund(id));`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
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

func ListMutualFundHelper() (string, error) {
	mfs, err := ListMutualFunds()
	if err != nil {
		return "", nil
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"MfId", "Name", "Folio", "Type"})
	for _, mf := range mfs {
		table.Append([]string{strconv.Itoa(mf.mfid), mf.name, mf.folio, mf.amfiSchemeCode})
	}
	table.Render()
	return buf.String(), nil
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

func ListMutualFundPurchases() ([]MutualFundPurchase, error) {
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
	where mutual_fund_purchase.mfid=mutual_fund.mfid
	order by mutual_fund_purchase.id`
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
		return err
	}
	return InsertMutualFundPurchase(mfp)
}

func RemoveMutualFundPurchaseHelper(mfpid int) error {
	var mfp MutualFundPurchase
	mfp.id = mfpid
	return RemoveMutualFundPurchase(mfp)
}

func ListMutualFundPurchaseHelper() (string, error) {
	mfps, err := ListMutualFundPurchases()
	if err != nil {
		return "", err
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
	return buf.String(), nil
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

func GetMutualFundSummary(
	navHelper NavHelper, nowHelper NowHelper) ([]MutualFundSummary, error) {
	mfps, err := ListMutualFundPurchases()
	if err != nil {
		return nil, err
	}

	if nowHelper == nil {
		nowHelper = time.Now
	}

	mutualFundsMap := make(map[MutualFund]bool)
	for _, mfp := range mfps {
		mutualFundsMap[mfp.MutualFund] = true

	}

	var mutualFunds []MutualFund
	for mf, _ := range mutualFundsMap {
		mutualFunds = append(mutualFunds, mf)
	}

	latestNav, err := navHelper(mutualFunds)
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
		days := nowHelper().Sub(mfp.time).Hours() / 24
		sumAmountxDays[mf] += mfp.amount * days
		sumAmount[mf] += mfp.amount
		mfs.amount += mfp.amount
		mfs.units += mfp.amount / mfp.nav
		mfsMap[mf] = mfs
	}

	var mfss []MutualFundSummary
	for mf, _ := range mutualFundsMap {
		mfs := mfsMap[mf]
		mfs.avgDays = int(sumAmountxDays[mf] / sumAmount[mf])
		mfs.currentValue = latestNav[mf] * mfs.units
		mfs.appreciation = ((mfs.currentValue - mfs.amount) / mfs.amount) * 100
		appr := mfs.currentValue / mfs.amount
		power := 365 / float64(mfs.avgDays)
		mfs.projectedYearlyRet = (math.Pow(appr, power) - 1) * 100
		mfss = append(mfss, mfs)
	}

	return mfss, nil
}

func MutualFundSummaryHelper(navHelper NavHelper, nowHelper NowHelper) (string, error) {
	mfsums, err := GetMutualFundSummary(navHelper, nowHelper)

	var mfssorter mutualFundSummarySorter
	mfssorter.mfsums = mfsums
	mfssorter.by = sortByMfId
	sort.Sort(&mfssorter)

	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"MfId", "Name", "Type", "AvgDays",
		"Amount", "Units", "CurrentVal", "Appr", "PrjRet"})
	for _, mfsum := range mfsums {
		amount := fmt.Sprintf("%.3f", mfsum.amount)
		units := fmt.Sprintf("%.3f", mfsum.units)
		currentValue := fmt.Sprintf("%.3f", mfsum.currentValue)
		appreciation := fmt.Sprintf("%.3f", mfsum.appreciation)
		projectedYearlyRet := fmt.Sprintf("%.3f", mfsum.projectedYearlyRet)
		table.Append([]string{strconv.Itoa(mfsum.mfid), mfsum.name, mfsum.mftype,
			strconv.Itoa(mfsum.avgDays), amount, units, currentValue, appreciation,
			projectedYearlyRet})
	}
	table.Render()
	return buf.String(), nil
}
