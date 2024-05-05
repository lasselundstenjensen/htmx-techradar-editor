package main

import (
	"encoding/csv"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CSVRow struct {
	Name        string
	Ring        string
	Quadrant    string
	IsNew       bool
	Move        int
	Description string
}

type CSVData struct {
	Headers []string
	Rows    []CSVRow
}

var csvData CSVData

func parseCSVRecord(record []string) (CSVRow, error) {
	if len(record) != 6 {
		return CSVRow{}, errors.New("incorrect number of columns")
	}

	isNew, err := strconv.ParseBool(record[3])
	if err != nil {
		return CSVRow{}, errors.New("invalid boolean value")
	}

	move, err := strconv.Atoi(record[4])
	if err != nil {
		return CSVRow{}, errors.New("invalid integer value")
	}

	return CSVRow{
		Name:        record[0],
		Ring:        record[1],
		Quadrant:    record[2],
		IsNew:       isNew,
		Move:        move,
		Description: record[5],
	}, nil
}

func saveChanges(form *http.Request) {
	form.ParseMultipartForm(1024)
	var updatedRows []CSVRow
	names := form.PostForm["name"]
	rings := form.PostForm["ring"]
	quadrants := form.PostForm["quadrant"]
	isNews := form.PostForm["isNew"]
	moves := form.PostForm["move"]
	descriptions := form.PostForm["description"]

	for i := 0; i < len(names); i++ {
		isNew := isNews[i] == "true"
		move, _ := strconv.Atoi(moves[i])

		updatedRows = append(updatedRows, CSVRow{
			Name:        names[i],
			Ring:        rings[i],
			Quadrant:    quadrants[i],
			IsNew:       isNew,
			Move:        move,
			Description: descriptions[i],
		})
	}

	csvData.Rows = updatedRows
}

func main() {
	r := gin.Default()

	// Create a custom template function
	r.SetFuncMap(template.FuncMap{
		"safe": safeHTML,
	})

	r.LoadHTMLGlob("templates/*")

	// Serve the HTML page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Handle file upload
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Failed to upload file")
			return
		}

		// Read the file content
		src, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read file")
			return
		}
		defer src.Close()

		r := csv.NewReader(src)
		r.Comma = ',' // Set the delimiter to a comma

		records, err := r.ReadAll()
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid CSV data")
			return
		}

		headers := []string{"name", "ring", "quadrant", "isNew", "move", "description"}
		var rows []CSVRow
		for _, record := range records[1:] {
			row, err := parseCSVRecord(record)
			if err == nil {
				rows = append(rows, row)
			}
		}

		csvData = CSVData{
			Headers: headers,
			Rows:    rows,
		}

		c.HTML(http.StatusOK, "result.html", gin.H{
			"Data": csvData,
		})
	})

	// Save changes made to the CSV table
	r.POST("/save", func(c *gin.Context) {
		saveChanges(c.Request)
		c.HTML(http.StatusOK, "result.html", gin.H{
			"Data": csvData,
		})
	})

	// Add a new row
	r.POST("/add", func(c *gin.Context) {
		saveChanges(c.Request)
		csvData.Rows = append(csvData.Rows, CSVRow{
			Name:        "",
			Ring:        "",
			Quadrant:    "",
			IsNew:       false,
			Move:        0,
			Description: "",
		})
		c.HTML(http.StatusOK, "result.html", gin.H{
			"Data": csvData,
		})
	})

	// Delete a row
	r.POST("/delete/:index", func(c *gin.Context) {
		saveChanges(c.Request)
		index, _ := strconv.Atoi(c.Param("index"))
		if index >= 0 && index < len(csvData.Rows) {
			csvData.Rows = append(csvData.Rows[:index], csvData.Rows[index+1:]...)
		}
		c.HTML(http.StatusOK, "result.html", gin.H{
			"Data": csvData,
		})
	})

	// Download CSV file
	r.GET("/download", func(c *gin.Context) {
		saveChanges(c.Request)

		c.Header("Content-Disposition", "attachment; filename=table.csv")
		c.Header("Content-Type", "text/csv")

		writer := csv.NewWriter(c.Writer)
		writer.Write(csvData.Headers)
		for _, row := range csvData.Rows {
			writer.Write([]string{
				row.Name,
				row.Ring,
				row.Quadrant,
				strconv.FormatBool(row.IsNew),
				strconv.Itoa(row.Move),
				row.Description,
			})
		}
		writer.Flush()
	})

	r.Run(":8080")
}

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}
