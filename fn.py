#!/usr/bin/env python

import argparse
import os
import sqlite3
import sys
import parser
import requests
import re
import datetime
import dateutil
import threading
import collections
from prettytable import PrettyTable

assert "MFDB" in os.environ, "Do export MFDB=<path-to-mf-database>"
mfDb = os.getenv("MFDB")

CONN = sqlite3.connect(mfDb)
CURSOR = CONN.cursor()

TOOL_VERSION = 2

AMFI_MF_IDS = {
    'ABN  AMRO': '39',
    'AEGON': '50',
    'Alliance Capital': '1',
    'Axis': '53',
    'Baroda Pioneer': '4',
    'Benchmark': '36',
    'Birla Sun Life': '3',
    'BNP Paribas': '59',
    'BOI AXA': '46',
    'Canara Robeco': '32',
    'Daiwa': '60',
    'DBS Chola': '31',
    'Deutsche': '38',
    'DHFL Pramerica': '58',
    'DSP BlackRock': '6',
    'Edelweiss': '47',
    'Escorts': '13',
    'Fidelity': '40',
    'Fortis': '51',
    'Franklin': '27',
    'GIC': '8',
    'Goldman Sachs': '49',
    'HDFC': '9',
    'HSBC': '37',
    'ICICI Prudential': '20',
    'IDBI': '57',
    'IDFC': '48',
    'IIFL': '62',
    'Indiabulls': '63',
    'ING': '14',
    'Invesco': '42',
    'JM Financial': '16',
    'JPMorgan': '43',
    'Kotak': '17',
    'L&T': '56',
    'LIC': '18',
    'Mahindra': '69',
    'Mirae Asset': '45',
    'Morgan Stanley': '19',
    'Motilal Oswal': '55',
    'Peerless': '54',
    'PineBridge': '44',
    'PNB': '34',
    'PPFAS': '64',
    'PRINCIPAL': '10',
    'Quantum': '41',
    'Reliance': '21',
    'Sahara': '35',
    'SBI': '22',
    'Shinsei': '52',
    'Shriram': '67',
    'Standard Chartered': '2',
    'Sundaram': '33',
    'Tata': '25',
    'Taurus': '26',
    'Union': '61',
    'UTI': '28',
    'Zurich India': '29',
}


def get_table(*fields):
    pretty_table = PrettyTable(list(fields))
    empty_row = [""] * len(fields)
    return pretty_table, empty_row


def show_mutual_fund(parsed_args):
    fields = ["id", "folio", "name", "type", "nav", "scheme_code"]
    if parsed_args.url:
        fields.append("url")
    table, empty_row = get_table(*fields)
    for row in CURSOR.execute("select %s from mutual_fund" % ",".join(fields)):
        table.add_row(list(row))
    print table


def query_filter_list(query, field_name, member, args):
    if args:
        newArgs = ['"' + x + '"' for x in args]
        check = "in" if member else "not in"
        query += " and %s %s (%s)" % (field_name, check, ','.join(newArgs))
    return query


def query_filter_value(query, field_name, oper, arg):
    if arg:
        query += " and %s%s\"%s\"" % (field_name, oper, arg)
    return query


def show_mutual_fund_purchase(parsed_args):
    table, empty_row = get_table(
        "Index", "Name", "Type", "Amount", "Time", "Units", "Purchase NAV")
    query = "select mutual_fund.name, mutual_fund.type, amount, time, units, mutual_fund_purchase.nav " \
            "from mutual_fund_purchase join mutual_fund ""where mutual_fund_purchase.id=mutual_fund.id"
    query = query_filter_list(
        query, "mutual_fund.type", True, parsed_args.type)
    query = query_filter_list(
        query, "mutual_fund.type", False, parsed_args.exclude_type)
    query = query_filter_list(query, "mutual_fund.id", True, parsed_args.id)
    query = query_filter_list(query, "mutual_fund.id",
                              False, parsed_args.exclude_id)
    query = query_filter_value(query, "time", "<=", parsed_args.date)
    query += " order by time"
    for idx, row in enumerate(CURSOR.execute(query)):
        table.add_row([idx + 1] + list(row))
    print table


def get_mutual_fund_status_rows( date=None, type=None, exclude_type=None,
                                 id=None, exclude_id=None,
                                 from_money_control=True, quiet=False ):
    rows = [ ]
    if not date:
        date = str(datetime.date.today())
        date_obj = datetime.date.today()
    else:
        date = date
        date_obj = datetime.datetime.strptime(date, "%Y-%m-%d").date()

    total_amount_invested = 0
    total_weighted_sum_x_average = 0
    total_current_value = 0
    purchases = collections.defaultdict(list)
    query = ( "select mutual_fund.name, mutual_fund.type, mutual_fund.id, "
                    " mutual_fund_purchase.amount, "
                    " mutual_fund_purchase.units, "
                    " mutual_fund_purchase.time from "
              "mutual_fund_purchase join mutual_fund "
              "where mutual_fund_purchase.id=mutual_fund.id " )

    query = query_filter_list(
        query, "mutual_fund.type", True, type)
    query = query_filter_list(
        query, "mutual_fund.type", False, exclude_type)
    query = query_filter_list(query, "mutual_fund.id", True, id)
    query = query_filter_list(query, "mutual_fund.id",
                              False, exclude_id)

    latest_nav_dict = get_nav(date, from_money_control, quiet=quiet)

    for row in CURSOR.execute(query):
        date_delta = (
            date_obj -
            datetime.datetime.strptime(
                row[5],
                "%Y-%m-%d").date()).days
        nav = latest_nav_dict[row[2]]
        purchases[row[0]].append({
            'type': row[1],
            'nav': nav,
            'amount': row[3],
            'units': row[4],
            'delta': date_delta
        })

    for name, info in purchases.items():
        weighted_sum_x_average = sum([x['amount'] * x['delta'] for x in info])
        amount_invested = sum([x['amount'] for x in info])
        total_units = sum([x['units'] for x in info])
        weighted_average = int(float(
            weighted_sum_x_average) / float(amount_invested))
        total_amount_invested += amount_invested
        total_weighted_sum_x_average += weighted_sum_x_average
        current_value = int(total_units * info[0]['nav'])
        total_current_value += current_value
        appreciation = round(
            ((current_value - amount_invested) * 100.0) / amount_invested, 2)
        if weighted_average:
            # When the mutual was started on that day.
            yearly_return = (round((1.0 + (appreciation / 100.0))
                                ** (365.0 / weighted_average), 4) - 1.0) * 100
        else:
            yearly_return = "NA"
        rows.append([
            name,
            info[0]['type'],
            info[0]['nav'],
            weighted_average,
            amount_invested,
            total_units,
            current_value,
            appreciation,
            yearly_return
        ])

    total_info = {
            'total_weighted_sum_x_average': total_weighted_sum_x_average,
            'total_amount_invested': total_amount_invested,
            'total_current_value': total_current_value,
    }

    return rows, total_info

def show_mutual_fund_status(parsed_args):
    table, empty_row = get_table("Index", "Name", "Type", "Ltst NAV",
                                 "Avg Days", "Amount", "Units",
                                 "Crnt val", "Appr", "Pj Yrly Ret", )
    rows = []
    sort_field = {"appr": 7, "amt": 6, "name": 0, "type": 1, "yappr": 8}
    field = parsed_args.sort

    rows, total_info = get_mutual_fund_status_rows(
            date=parsed_args.date, type=parsed_args.type,
            exclude_type=parsed_args.exclude_type,
            id=parsed_args.id, exclude_id=parsed_args.exclude_id,
            from_money_control=parsed_args.moneycontrol )

    total_weighted_sum_x_average = total_info[ 'total_weighted_sum_x_average' ]
    total_amount_invested = total_info[ 'total_amount_invested' ]
    total_current_value = total_info[ 'total_current_value' ]

    for idx, row in enumerate(
            sorted(rows, key=lambda x: x[sort_field[field]])):
        table.add_row([idx + 1] + row)
    table.add_row(empty_row)
    total_weighted_average_days = total_weighted_sum_x_average / total_amount_invested
    total_appreciation = round(
        ((total_current_value -
          total_amount_invested) *
         100.0) /
        total_amount_invested,
        2)
    yearly_return = (round((1.0 + (total_appreciation / 100.0))
                           ** (365.0 / total_weighted_average_days), 4) - 1.0) * 100
    table.add_row([
        "",
        "Total",
        "",
        "",
        total_weighted_average_days,
        total_amount_invested,
        "",
        total_current_value,
        total_appreciation,
        yearly_return
    ])
    print table


def update_latest_nav(parsed_args):
    queries = []
    date_str = str(datetime.date.today())
    queries.append((
        "Deleting current NAV for date %s" % date_str,
        ["delete from nav where date(date)='%s'" % date_str]
    ))

    def fetch_url(row):
        rsp = requests.get(row[2])
        m = re.search(r'class="bd30tp">(\d+\.\d+)</span>', rsp.text)
        assert m
        latest_nav = m.group(1)
        helper_str = "Setting NAV=%s for %s" % (latest_nav, row[1])
        query1 = "update mutual_fund set nav=%s where id=%s" % (
            latest_nav, row[0])
        query2 = ("insert into nav(mutual_fund_id, date, nav) " "values(%s, CURRENT_TIMESTAMP, %s)" % (
            row[0], latest_nav))
        queries.append((helper_str, [query1, query2]))

    threads = [
        threading.Thread(
            target=fetch_url,
            args=(
                row,
            )) for row in CURSOR.execute("select id, name, url from mutual_fund")]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()

    for query in queries:
        print query[0]
        for q in query[1]:
            CURSOR.execute(q)
    CONN.commit()


def fetch_data_from_amfi(date):
    latest_nav = { }
    urls_to_fetch = set()
    for row in CURSOR.execute("select name from mutual_fund"):
        name = row[0]
        matching_mfs = [ x for x in AMFI_MF_IDS if x.lower() in name.lower() ]
        urls_to_fetch |= set([get_amf_url(AMFI_MF_IDS[x],to_date=date) for x in matching_mfs])

    url_data = [ ]

    def fetch_url(url):
        rsp = requests.get(url)
        url_data.append(rsp.text)

    threads = [
        threading.Thread(
            target=fetch_url,
            args=(
                url,
            )) for url in urls_to_fetch ]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()

    all_data = '\n'.join(url_data)
    return all_data


def dump_amfi_data_for_debug(parsed_args):
    print fetch_data_from_amfi()


def get_nav(date, from_money_control=False, quiet=False):
    if from_money_control:
        return get_nav_from_moneycontrol(quiet=quiet)
    else:
        return get_nav_from_amfi(date, quiet=quiet)


def get_nav_from_amfi(date, quiet=False):
    all_data = fetch_data_from_amfi(date)
    lines = [x for x in all_data.split("\n") if x]
    scheme_code_to_nav = { }
    scheme_code_to_last_date = { }
    for line in lines:
        fields = line.split(";")
        # 123651;ICICI Prudential Global Stable Equity Fund - Growth;13.35;12.95;13.35;14-Aug-2017
        if len(fields) != 6:
            continue
        # We assume that data will be sorted, so the last one updated will be
        # the last dates value
        scheme_code_to_nav[fields[0]] = fields[2]
        scheme_code_to_last_date[fields[0]] = fields[5].strip()

    mf_id_to_nav = { }
    for row in CURSOR.execute("select id, scheme_code, name from mutual_fund"):
        mf_id = row[0]
        scheme_code = row[1]
        name = row[2]
        if scheme_code not in scheme_code_to_nav:
            raise Exception("Scheme code not available for %s" % name)
        if not quiet:
            print "Using NAV of date %s for %s" % (
                    scheme_code_to_last_date[scheme_code], name)
        mf_id_to_nav[mf_id] = float(scheme_code_to_nav[scheme_code])
    return mf_id_to_nav


def get_nav_from_moneycontrol(quiet=False):
    mf_id_to_nav = { }

    def fetch_url(mf_id, url, name):
        rsp = requests.get(url)
        m = re.search(r'class="bd30tp">(\d+\.\d+)</span>', rsp.text)
        assert m
        latest_nav = m.group(1)
        m = re.search(r'NAV as on (\d+ \w+, \d+)', rsp.text)
        assert m
        date = m.group(1)
        if not quiet:
            print "Using NAV of date %s for %s" % ( date, name)
        mf_id_to_nav[mf_id] = float(latest_nav)

    threads = [
        threading.Thread(
            target=fetch_url,
            args=(
                row[0], row[1], row[2]
            )) for row in CURSOR.execute("select id, url, name from mutual_fund")]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()

    return mf_id_to_nav


def insert_mf_purchase(parsed_args):
    units = round(parsed_args.amount / parsed_args.nav, 3)
    query = ("insert into mutual_fund_purchase(id, units, nav, amount, time) "
             "values(%s, %s, %s, %s, \"%s\")" % (
                 parsed_args.id,
                 units,
                 parsed_args.nav,
                 parsed_args.amount,
                 parsed_args.date
             ))
    CURSOR.execute(query)
    CONN.commit()


def insert_fixed_deposit(parsed_args):
    query = (
        "insert into fixed_deposit(name, amount, rate, tenure, deposit_date, maturity_date, maturity_amount) "
        "values(\"%s\", %s, %s, %s, \"%s\", \"%s\", %s)" %
        (parsed_args.name,
         parsed_args.amount,
         parsed_args.rate,
         parsed_args.tenure,
         parsed_args.deposit_date,
         parsed_args.maturity_date,
         parsed_args.maturity_amount))
    CURSOR.execute(query)
    CONN.commit()


def show_fixed_deposit_sort_fields():
    sort_field = {"amt": 3, "name": 2, "rate": 4, "date": 7}
    return sort_field


def show_fixed_deposit(parsed_args):
    table, empty_row = get_table("Id", "Internal Id", "Name", "Amount",
                                 "Rate", "Tenure", "Deposit Date", "Maturity Date", "Maturity Amount")
    sort_field = show_fixed_deposit_sort_fields()
    field = parsed_args.sort
    fds = []
    for row in CURSOR.execute("select * from fixed_deposit"):
        fd_info = list(row)
        fds.append(fd_info)

    for idx, fd in enumerate(sorted(fds, key=lambda x: x[sort_field[field]])):
        table.add_row([idx + 1] + fd)
    print table


def show_nav(parsed_args):
    table, empty_row = get_table("Index", "Date")
    for idx, row in enumerate(CURSOR.execute(
            "select date(date) from nav group by date(date);")):
        table.add_row([idx + 1] + list(row))
    print table


def datetime_to_amf_date(date):
    day = date.strftime("%d")
    month = date.strftime("%B")[:3]
    year = date.strftime("%Y")
    return "%s-%s-%s" % (day, month, year)


def amf_date_to_datetime(date_str):
    return datetime.strptime(date_str, '%d-%b-%Y')


def get_amf_url(mf_id, from_date=None, to_date=None):
    url_template = ( "http://portal.amfiindia.com/"
                     "DownloadNAVHistoryReport_Po.aspx?"
                     "mf=%s&tp=1&frmdt=%s&todt=%s" )
    if not to_date:
        to_date = datetime_to_amf_date(datetime.date.today())
        to_date_obj = datetime.date.today()
    else:
        to_date_obj = datetime.datetime.strptime(to_date, "%Y-%m-%d").date()
    to_date = datetime_to_amf_date(to_date_obj)

    if not from_date:
        date1 = to_date_obj - datetime.timedelta(days=7)
        from_date = datetime_to_amf_date(date1)
    return url_template % (mf_id, from_date, to_date)


def print_amf_url(parsed_args):
    for mf_name, mf_id in AMFI_MF_IDS.items():
        print mf_name
        print get_amf_url(mf_id, parsed_args.from_date, parsed_args.to_date)
        print ""


FUNCTION_MAP = {
    'smf': show_mutual_fund,
    'smfp': show_mutual_fund_purchase,
    'smfs': show_mutual_fund_status,
    'uln': update_latest_nav,
    'imfp': insert_mf_purchase,
    'ifd': insert_fixed_deposit,
    'sfd': show_fixed_deposit,
    'snav': show_nav,
    'pamfurl': print_amf_url,
    'dumpmfdata': dump_amfi_data_for_debug,
}


def migrate_db_to_next_version(current_version):
    if current_version == 1:
        # 1) nav table was dropped. instead query from http://portal.amfiindia.com
        #    for history.
        # 2) add a new column scheme code in mutual_fund table
        CURSOR.execute("drop table if exists nav")
        try:
            CURSOR.execute(
                "alter table mutual_fund add column scheme_code TEXT")
        except sqlite3.OperationalError as e:
            if 'duplicate column name' not in str(e):
                raise
        CURSOR.execute("drop table if exists tool_info")
        CURSOR.execute("create table tool_info(key TEXT, value TEXT)")
        CURSOR.execute("insert into tool_info values( 'db_version', 1 )")
        url = "https://www.amfiindia.com/spages/NAVOpen.txt"
        print "This version of DB is adding scheme code to mutual funds."
        print "Open %s for getting scheme codes and fill in DB." % url
    else:
        msg = "No known migration handler to version=%s" % (
            current_version + 1)
        assert False, msg
    print "DB migrated to version %s" % (current_version + 1)
    CURSOR.execute("update tool_info set value=%s where key='db_version'" %
                   (current_version + 1))
    CONN.commit()


def check_db_version():
    try:
        row = CURSOR.execute(
            "select value from tool_info where key='db_version'")
        db_version = int(list(row)[0][0])
    except (sqlite3.OperationalError, IndexError) as e:
        # db_version was introduced in tool_version 2. So if table
        # doesn't exist, then assume it is version 1.
        db_version = 1

    for version in range(db_version, TOOL_VERSION):
        migrate_db_to_next_version(version)


def main():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(title='subcommands')

    check_db_version()

    smf = subparsers.add_parser('smf')
    smf.set_defaults(subcommand='smf')
    smf.add_argument("--url", action='store_true')

    smfp = subparsers.add_parser('smfp')
    smfp.set_defaults(subcommand='smfp')
    smfpFilter = smfp.add_mutually_exclusive_group(required=False)
    smfpFilter.add_argument('-t', '--types', nargs="+",
                            type=str, action="store", dest='type', help="fund type")
    smfpFilter.add_argument('-e', '--exclude-types', nargs="+", type=str,
                            action="store", dest='exclude_type', help="exclude fund types")
    smfpFilter.add_argument('-f', '--id', nargs="+",
                            type=str, action="store", dest='id', help="fund ids")
    smfpFilter.add_argument('-n', '--exclude-id', nargs="+", type=str,
                            action="store", dest='exclude_id', help="exclude fund ids")
    smfp.add_argument('-d', '--date', type=str, action="store",
                      dest='date', help="as on date")

    smfs = subparsers.add_parser('smfs')
    smfs.set_defaults(subcommand='smfs')
    smfsFilter = smfs.add_mutually_exclusive_group(required=False)
    smfsFilter.add_argument('-t', '--types', nargs="+",
                            type=str, action="store", dest='type', help="fund type")
    smfsFilter.add_argument('-e', '--exclude-types', nargs="+", type=str,
                            action="store", dest='exclude_type', help="exclude fund types")
    smfsFilter.add_argument('-f', '--id', nargs="+",
                            type=str, action="store", dest='id', help="fund ids")
    smfsFilter.add_argument('-n', '--exclude-id', nargs="+", type=str,
                            action="store", dest='exclude_id', help="exclude fund ids")
    smfs.add_argument('-d', '--date', type=str, action="store",
                      dest='date', help="as on date")
    smfs.add_argument(
        '--sort', choices=["appr", "amt", "name", "type", "yappr"], default="appr")
    smfs.add_argument('--moneycontrol', action="store_true")

    imfp = subparsers.add_parser('imfp')
    imfp.set_defaults(subcommand='imfp')
    imfp.add_argument('id', type=int)
    imfp.add_argument('nav', type=float)
    imfp.add_argument('amount', type=int)
    imfp.add_argument('date', type=str)

    dumpmfdata = subparsers.add_parser('dumpmfdata')
    dumpmfdata.set_defaults(subcommand='dumpmfdata')

    uln = subparsers.add_parser('uln')
    uln.set_defaults(subcommand='uln')

    ifd = subparsers.add_parser('ifd')
    ifd.set_defaults(subcommand='ifd')
    ifd.add_argument('name', type=str)
    ifd.add_argument('amount', type=int)
    ifd.add_argument('rate', type=float)
    ifd.add_argument('tenure', type=int)
    ifd.add_argument('deposit_date', type=str)
    ifd.add_argument('maturity_date', type=str)
    ifd.add_argument('maturity_amount', type=str)

    snav = subparsers.add_parser('snav')
    snav.set_defaults(subcommand='snav')

    pamfurl = subparsers.add_parser('pamfurl')
    pamfurl.set_defaults(subcommand='pamfurl')
    pamfurl.add_argument('--from_date')
    pamfurl.add_argument('--to_date')

    sfd = subparsers.add_parser('sfd')
    sfd.set_defaults(subcommand='sfd')
    sfd.add_argument(
        '--sort', choices=show_fixed_deposit_sort_fields().keys(), default="amt")

    parsed_args = parser.parse_args()
    FUNCTION_MAP[parsed_args.subcommand](parsed_args)
    CONN.close()

if __name__ == "__main__":
    main()
