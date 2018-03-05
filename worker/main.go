package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FILEHOST File where we keep all hosts.
const FILEHOST string = "/Users/pawel/Private/Code/golang/SuperEasyMonitoring/worker/conf/hosts.json"

// LOGFILE path
const LOGFILE string = "/Users/pawel/Private/Code/golang/SuperEasyMonitoring/worker/logs/hosts.log"

// DB path
const dbpath = "/Users/pawel/Private/Code/golang/SuperEasyMonitoring/conf/foo.db"

// create slice of hosts
var hosts []string

// ping command path
var ping string

//
var database *sql.DB

//
var hostID int

// Check create interface for all checks
type Check interface {
	status() (bool, error)
	hostname() string
	checkname() string
}

// Hosts bla
type Hosts struct {
	Hosts []Host `json:"hosts"`
}

// Host bla
type Host struct {
	Hostname     string       `json:"hostname"`
	CheckICMP    checkICMP    `json:"icmp"`
	CheckHTTP    checkHTTP    `json:"http"`
	CheckTCPPort checkTCPPort `json:"tcp"`
}

// type struct for ICMP
type checkICMP struct {
	Hostname  string `json:"hostname"`
	Enabled   bool   `json:"enabled"`
	Timeout   int    `json:"timeout"`
	Checkname string `json:"checkname"`
}

// type struct for HTTP
type checkHTTP struct {
	Hostname  string `json:"hostname"`
	Enabled   bool   `json:"enabled"`
	Timeout   int    `json:"timeout"`
	Port      int    `json:"port"`
	Checkname string `json:"checkname"`
}

// type struct for TCP
type checkTCPPort struct {
	Hostname  string `json:"hostname"`
	Enabled   bool   `json:"enabled"`
	Timeout   int    `json:"timeout"`
	Port      int    `json:"port"`
	Checkname string `json:"checkname"`
}

// Status type struct for Status
type Status struct {
	ID               int
	Host             string
	Checkname        string
	Status           string
	InsertedDatetime string
}

// customErr is for returning error so we don't need to wring it every time
// when we have any err variable, we can simply call this one.
func customErr(text string, e error) {
	if e != nil {
		fmt.Println(text, e)
		os.Exit(1)
	}
}

// init function here is opening file name with the list of all hosts
// and adding each host to the slice so we can use it everywhere in the code.
func init() {
	file, err := os.Open(FILEHOST)
	if err != nil {
		customErr("ERROR: Cannot read file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		customErr("Cannot scan file: %v", err)
	}

	err = checkOS()
	customErr("Cannot checkOS: %v", err)
}

// InitDB initialize database
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	database = db
	return database
}

// CreateTable creating table
func CreateTable(db *sql.DB) {
	// create table if not exists
	sqlTable := `
	CREATE TABLE IF NOT EXISTS Monitoring(
		ID TEXT NOT NULL PRIMARY KEY,
		Hostname TEXT,
		Checkname TEXT,
		Status TEXT,
		InsertedDatetime DATETIME
	);
	`

	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
}

// StoreItem bla bla bla
func StoreItem(db *sql.DB, items Status) {
	sqlAdditem := `
	INSERT OR REPLACE INTO Monitoring(
		ID,
		Hostname,
		Checkname,
		Status,
		InsertedDatetime
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sqlAdditem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(items.ID, items.Host, items.Checkname, items.Status, items.InsertedDatetime)
	if err2 != nil {
		panic(err2)
	}
}

// checkOS is setting up global variable ping with the specific
// path to it per OS
func checkOS() error {
	if runtime.GOOS == "darwin" {
		ping = "/sbin/ping"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "386" {
		ping = "/bin/ping"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		ping = "/usr/bin/ping"
	} else {
		return errors.New("OS is not supported")
	}
	return nil
}

// pinger will ping host/ip using timeout to timeout
// and then will return True or False
func (c checkICMP) status() (bool, error) {
	out, err := exec.Command(ping, c.Hostname, "-c 1", "-t", intToString(c.Timeout)).Output()
	if strings.Contains(string(out), "64 bytes from") {
		return true, nil
	}
	return false, err
}

// checkTCPPort is checking if we can connect to the specific port
// such as SSH(22) or MYSQL(3306)
func (c checkTCPPort) status() (bool, error) {
	return true, nil
}

// checkHTTPRequest is checking if we can connect on the specific port
// in the specific ammount of time, otherwise it's returning
// true or false and error
func (c checkHTTP) status() (bool, error) {
	client := http.Client{
		Timeout: time.Second * time.Duration(c.Timeout),
	}
	resp, err := client.Get("http://" + c.Hostname)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

// intToString function will simply convert int to string
// this is for exec.Command()
func stringToInt(s int) string {
	return strconv.Itoa(s)
}

// intToString will simply convert time.Duration type
// to string for exec.Command()
func intToString(t int) string {
	return strconv.Itoa(t)
}

func (c checkICMP) hostname() string {
	return c.Hostname
}

func (c checkHTTP) hostname() string {
	return c.Hostname
}

func (c checkTCPPort) hostname() string {
	return c.Hostname
}

func (c checkICMP) checkname() string {
	return c.Checkname
}

func (c checkHTTP) checkname() string {
	return c.Checkname
}

func (c checkTCPPort) checkname() string {
	return c.Checkname
}

// getStatus is an interface type so we can call it from every
// struct which match status() (bool, error)
func getStatus(c Check) {
	// get time now and convert it to string
	timeNow := getCurrentTime()

	status, err := c.status()
	if err != nil {
		fmt.Println("Error:", err)
	}

	if status == true {
		log.Println(c.checkname(), c.hostname(), returnStatus(status))
		items := Status{hostID, c.hostname(), c.checkname(), returnStatus(status), timeNow}
		StoreItem(database, items)
		hostID++
	}
	if status == false {
		log.Println(c.checkname(), c.hostname(), returnStatus(status))
		items := Status{hostID, c.hostname(), c.checkname(), returnStatus(status), timeNow}
		StoreItem(database, items)
		hostID++
	}
}

// get currect time and convert it to string
func getCurrentTime() string {
	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

// returnStatus text
func returnStatus(status bool) string {
	if status != true {
		return "CRITICAL"
	}
	return "OK"
}

// readJSON file
func readJSON() Hosts {
	// Open our jsonFile
	jsonFile, err := os.Open(FILEHOST)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println("Successfully Opened hosts.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Hosts array
	var hosts Hosts

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'hosts' which we defined above
	json.Unmarshal(byteValue, &hosts)

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	return hosts
}

func runCheckICMP(host int, h Hosts) {
	if h.Hosts[host].CheckICMP.Enabled {
		checkStruct := checkICMP{
			Hostname:  h.Hosts[host].Hostname,
			Timeout:   h.Hosts[host].CheckICMP.Timeout,
			Checkname: "icmp",
		}
		getStatus(checkStruct)
	}
}

func runCheckHTTP(host int, h Hosts) {
	if h.Hosts[host].CheckHTTP.Enabled {
		checkStruct := checkHTTP{
			Hostname:  h.Hosts[host].Hostname,
			Timeout:   h.Hosts[host].CheckHTTP.Timeout,
			Port:      h.Hosts[host].CheckHTTP.Port,
			Checkname: "http",
		}
		getStatus(checkStruct)
	}
}

func runCheckTCPPort(host int, h Hosts) {
	if h.Hosts[host].CheckTCPPort.Enabled {
		checkStruct := checkTCPPort{
			Hostname:  h.Hosts[host].Hostname,
			Timeout:   h.Hosts[host].CheckHTTP.Timeout,
			Port:      h.Hosts[host].CheckHTTP.Port,
			Checkname: "tcp",
		}
		getStatus(checkStruct)
	}
}

func runChecks(h Hosts) error {
	for host := range h.Hosts {
		// ICMP
		runCheckICMP(host, h)

		// HTTP
		runCheckHTTP(host, h)

		// TCP
		runCheckTCPPort(host, h)
	}
	return nil
}

func main() {
	// logging
	f, err := os.OpenFile(LOGFILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(" Problem with opening os.OpenFile(): ", err)
		os.Exit(0)
	}
	defer f.Close()
	log.SetOutput(f)

	// database
	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	err = runChecks(readJSON())
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
