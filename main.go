package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"github.com/tvitcom/go-workreport/mstime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	clientIp           string
	isPrivate          bool
	dimLen             int
	err                error
	currYearNeededDays [12]int = [12]int{22, 20, 21, 22, 22, 20, 22, 22, 20, 23, 20, 21}
)

type (
	Env struct {
		// db *sql.DB
	}
	MonthesReport struct {
		MonthId      int
		Month        string
		Needed       int
		ActualNeeded int
		WorkedOut    int     // inMinutes
		WorkedDebt   float64 // inMinutes
	}
	YearSummaryReport [12]MonthesReport
	SummaryWorktimes  struct {
		Needed         int     // inMinutes
		WorkedOut      int     // inMinutes
		WorkedDebt     float64 // inMinutes
		WorkedDebtDays float64 // inMinutes
	}
	ReportsMeta struct {
		Year     int
		Reporter string
		Manager  string
	}
	ViewSummary struct {
		Meta      ReportsMeta
		Monthes   YearSummaryReport
		Vacations int // inMinutes
		Summary   SummaryWorktimes
	}
	Workreport struct {
		Id_workreport int
		Id_attempt    int
		Id_project    int
		Id_user       int
		Year          int
		Month_num     int
		Date_ms       int
		Duration_m    string
		Task_name     string
	}
)

const (
	PRODMODE                 = false
	DRIVER                   = "mysql"
	DSN                      = "workreport:pass_to_workreport@/workreport"
	DBNAME                   = "searchforms"
	DIRSEP                   = "/"
	holyDays             int = 10 // Vacation
	sickDays             int = 4  // Absent by health
	workdayDuration      int = 8  //Hours
	reportYearRow        int = 0
	reportYearCell       int = 2
	reportReporterRow    int = 1
	reportReporterCell   int = 2
	reportManagerRow     int = 2
	reportManagerCell    int = 2
	reportStartRecords   int = 6
	worktimeDateCell     int = 0
	worktimeStartCell    int = 2
	worktimeFinishCell   int = 3
	worktimeDurationCell int = 4
	worktimeProjectCell  int = 5
	worktimeTasknameCell int = 6
)

func main() {

	// db, err := models.InitDB(DRIVER, DSN)
	// if err != nil {
	// 	log.Panic(err)
	// }
	env := &Env{
		// db: db
	}

	router := gin.New()

	if PRODMODE == true {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DisableConsoleColor()
	f, _ := os.Create("./logs/server.log")
	gin.DefaultWriter = io.MultiWriter(f)

	router.Delims("{*", "*}")
	router.LoadHTMLGlob("templates/pwa/*")

	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	s := &http.Server{
		Addr:           ":3000",
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	router.StaticFile("/favicon.ico", "./favicon.ico")

	router.Static("/images", "./static/images")
	router.Static("/scripts", "./static/scripts")
	router.Static("/styles", "./static/styles")

	//Root handler GET:
	router.GET("/", func(c *gin.Context) {
		c.Header("server", "REPORT Server")
		c.HTML(http.StatusOK, "form.html", gin.H{"reporter": "You!"})
	})
	router.GET("/fileupload", func(c *gin.Context) {
		c.Header("Location", "/")
		c.HTML(http.StatusUnsupportedMediaType, "error.html", gin.H{"status": "404", "cause": "It is not a http page"})
	})

	router.POST("/fileupload", func(c *gin.Context) {

		file, _ := c.FormFile("reportfile")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		isPrivMode := c.PostForm("privacy_mode")

		if isPrivMode != "true" {
			isPrivate = false
			clientIp = c.ClientIP()
			log.Println("IP:", clientIp)
		} else {
			isPrivate = true
			timeNow := time.Now()
			timeString := timeNow.String()
			clientIp = GetMD5Hash(timeString)
			log.Println("IP (as in the Priv mode):", clientIp)
		}

		dst := "./uploaded/" + clientIp + string(filepath.Ext(file.Filename))
		c.SaveUploadedFile(file, dst)

		mime := file.Header["Content-Type"][0]
		if err := xlsxValIdation(mime); err != nil {
			c.HTML(http.StatusUnsupportedMediaType, "error.html", gin.H{"status": "415", "cause": file.Filename + " is unsupported media type"})
			log.Println("Weird file:", dst)
			return
		}

		data, err := env.ParseReport(dst, isPrivate)
		if !isPrivate {
			log.Printf("%T", data)
			log.Println("Parsing result:", data)
		} else {
			log.Println("Parsing result: none display in the Private mode")
			// Will delete file as in the privacy mode
			var err = os.Remove(dst)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		if err != nil {
			c.HTML(http.StatusInsufficientStorage, "error.html", gin.H{"status": "507", "cause": file.Filename + " is not calcuated. Sorry."})
			log.Println("Error with parcing file:", dst)
			return
		}
		c.HTML(http.StatusOK, "summarynojs.html", gin.H{
			"meta":      data.Meta,
			"bymonth":   data.Monthes,
			"vacations": data.Vacations,
			"summary":   data.Summary,
		})
	})

	//Run web-server:
	s.ListenAndServe()
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func xlsxValIdation(mimetype string) (err error) {
	xlsxMime := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if xlsxMime != mimetype {
		return errors.New("Weird file")
	}
	return
}

/*
 * Parsing and get durtions of workedout time per month:
 */
func (env *Env) ParseReport(filepath string, isPrivate bool) (ViewSummary, error) {

	var Report YearSummaryReport //Year summary yeld durations (monthes durations summary)
	var viewSummary ViewSummary
	var Year, Name, Manager string

	// Calc current times parameters
	currDate := time.Now()
	currMonthId := int(currDate.Month())

	// Handle file
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		fmt.Println(err)
		return ViewSummary{}, err
	}

	// Walk by sheets (MONTHES:)
	sheetMonthes := f.GetSheetMap()

	for iSheet, sheetName := range sheetMonthes {
		var monthYeld int
		mData := MonthesReport{}

		// Walk by rows
		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Fatal(err)
		}

		// Get meta parameters:
		var researchedYear string
		if len(rows[reportYearRow]) > reportYearRow && rows[reportYearRow][reportYearCell] != "" {
			researchedYear = rows[reportYearRow][reportYearCell]
		}
		if researchedYear != "" {
			Year = researchedYear
		}
		var researchedName string
		if len(rows[reportReporterRow]) > reportReporterRow && rows[reportReporterRow][reportReporterCell] != "" {
			researchedName = rows[reportReporterRow][reportReporterCell]
		}

		if !isPrivate && researchedName != "" {
			Name = researchedName
		} else if isPrivate {
			Name = "Private Person"
		}

		// Walk by cells (TASKS:)
		maxRowAddr := len(rows)
		// var DayDurationItem int // In minutes
		// var DayDurationSumm []int // Slice of items
		for i := reportStartRecords; i < maxRowAddr; i++ {

			/* test PROTOTYPING:
			CELL: 0 43493
			CELL: 1 Mon
			CELL: 2 0.520833333333333
			CELL: 3 0.604166666666667
			CELL: 4 0.0833333333333339
			CELL: 5 DL
			CELL: 6 DDL-2107 Security improvement - exclude tilda files
			*/

			days, _ := mstime.GetStringTime(rows[i][4], true, false)
			inMinutes, _ := mstime.GetDurateInMinutes("00:00", days)
			if rows[i][0] == "" && rows[i][3] == "TOTAL" {
				// fmt.Println("TOTAL for day:", inMinutes)
				monthYeld = monthYeld + inMinutes
			}

			/*
				Workreport struct {
				  Id_workreport int
				  Id_attempt int
				  Id_project int
				  Id_user int
				  Year int
				  Month_num int
				  Date_ms int
				  Duration_m string
				  Task_name  string
				}
			*/
            //TODO: Save it is record into database
		}

		mData.MonthId = iSheet
		mData.Month = sheetName
		mData.WorkedOut = monthYeld

		if -1 < iSheet && iSheet < 13 {
			mData.Needed = currYearNeededDays[iSheet-1] * workdayDuration * 60 // in hours
			if mData.MonthId < int(currMonthId) {
				mData.ActualNeeded = mData.Needed
				mData.WorkedDebt = float64((mData.ActualNeeded - mData.WorkedOut) / 60) //in hours
			}

			Report[iSheet-1] = mData
		}
	}
	var yearNeededDays int
	var yearWorkedOut int
	lenYearNeededDays := len(currYearNeededDays)
	for k := 0; k < lenYearNeededDays; k++ {
		yearNeededDays += currYearNeededDays[k]
		yearWorkedOut += currYearNeededDays[k]
	}
	msDays, _ := strconv.Atoi(Year)
	year, _, _ := mstime.GetDateByDayNumber(msDays)

	//Minor calculations of SummaryWorktimes struct
	neededDaysSumm := yearNeededDays * workdayDuration * 60   // in hours
	workedoutDaysSumm := yearWorkedOut * workdayDuration * 60 //in hours

	//Minor calculation: actualsummaryDebts
	var actualSummaryDebts float64
	for n := 0; n < len(Report); n++ {

		// Only calc monthes before current
		if re := Report[n]; re.WorkedOut != 0 && re.MonthId < int(currMonthId) {
			actualSummaryDebts += float64(re.WorkedDebt)
		}
	}

	viewSummary.Meta = ReportsMeta{Year: year, Reporter: Name, Manager: Manager}
	viewSummary.Monthes = Report
	viewSummary.Vacations = (holyDays + sickDays) * workdayDuration // in hours
	ActualDebts := actualSummaryDebts - float64(viewSummary.Vacations)
	viewSummary.Summary = SummaryWorktimes{
		Needed:         neededDaysSumm,
		WorkedOut:      workedoutDaysSumm / 8,
		WorkedDebt:     ActualDebts,
		WorkedDebtDays: ActualDebts / 8}

	return viewSummary, nil
}
