package utils

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
)

func Open(dbpath string) *sql.DB {

	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open sqlite3 error!", err)
		os.Exit(-1)
	}
	return db

}

func InitSqliteDB(db string) {
	con := Open(db)
	sql := "create table tasks ( id  INTEGER PRIMARY KEY AUTOINCREMENT, startime varchar, endtime varchar, target varchar, status varchar,spendtime varchar);"

	_, err := con.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("init sqlite3 db success...")

}
func CheckTaskIsRunning(target string, dbpath string) (bool, int) {
	var db *sql.DB = Open(dbpath)
	defer db.Close()

	sql := fmt.Sprintf("select status from tasks where target='%s' and status='submit'", target)
	tx, _ := db.Begin()
	defer tx.Commit()
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "check tasks status error:%s ", err.Error())
		return true, 0
	}
	is_submit := false
	i := 0
	for rows.Next() {
		var s string
		i += 1
		rows.Scan(&s)
		if s == "submit" {
			is_submit = true
		}
	}
	return is_submit, i
}
func InsertTask(target string, dbpath string) error {

	var db *sql.DB = Open(dbpath)
	defer db.Close()
	tx, _ := db.Begin()
	defer tx.Commit()
	ex, err := db.Prepare("insert into tasks (id,startime,endtime,target,status,spendtime) values(?,?,?,?,?,?)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert tasks error:%s ", err.Error())
		return err
	}
	t := time.Now().Format("2006-01-02 15:04:05")
	r, err := ex.Exec(nil, t, "", target, "submit", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert tasks error:%s ", err.Error())
		return err
	}
	_, err = r.RowsAffected()
	return nil

}
func SearchStartTime(target string, dbpath string) string {
	var db *sql.DB = Open("./.quecmd.db")
	defer db.Close()
	tx, _ := db.Begin()
	defer tx.Commit()
	e, err := db.Query(fmt.Sprintf("select startime from tasks where target='%s' and status='submit' limit 1", target))
	defer e.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "query tasks error:%s ", err.Error())
		return ""
	}
	var startime string
	for e.Next() {
		e.Scan(&startime)
	}
	return startime
}

func UpdateTaskStatus(target string, status string, dbpath string) {
	var spendTime string
	var db *sql.DB = Open(dbpath)
	defer db.Close()
	tx, _ := db.Begin()
	defer tx.Commit()
	loc, _ := time.LoadLocation("Local")
	t := time.Now().Format("2006-01-02 15:04:05")
	st := SearchStartTime(target, dbpath)
	s, _ := time.ParseInLocation("2006-01-02 15:04:05", st, loc)
	spendTime = time.Now().Sub(s).String()

	ex, err := db.Prepare("update tasks set status=?,endtime=?,spendtime=? where target=? and status=?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert tasks error: %s", err.Error())
		return
	}

	r, err := ex.Exec(status, t, spendTime, target, "submit")
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert tasks error: %s ", err.Error())
		return
	}

	_, err = r.RowsAffected()
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert tasks error:%s ", err.Error())
		return
	}

}
func QueryAllTasks(t string, dbpath string) {

	var db *sql.DB = Open(dbpath)
	defer db.Close()

	var data [][]string
	var q string = "select id,startime,endtime,target,status,spendtime from tasks"
	tx, _ := db.Begin()
	defer tx.Commit()
	if len(t) != 0 {
		q = fmt.Sprintf("select id,startime,endtime,target,status,spendtime from tasks where target='%s'", t)
	}
	rows, err := db.Query(q)
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "query tasks error: %s ", err.Error())
		os.Exit(-1)
	}
	for rows.Next() {
		var id int
		var startTime string
		var endTime string
		var t string
		var status string
		var spendtime string
		err = rows.Scan(&id, &startTime, &endTime, &t, &status, &spendtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "query tasks error: %s", err)
			os.Exit(-1)
		}
		var s []string
		s = append(s, strconv.Itoa(id), startTime, endTime, t, status, spendtime)
		data = append(data, s)
	}

	var table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "startime", "endtime", "target", "status", "spendtime"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()

}
