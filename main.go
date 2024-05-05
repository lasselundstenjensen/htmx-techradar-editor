package main

import (
	"encoding/csv"
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
		records, err := r.ReadAll()
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid CSV data")
			return
		}

		headers := []string{"name", "ring", "quadrant", "isNew", "move", "description"}
		var rows []CSVRow
		for _, record := range records[1:] {
			if len(record) != len(headers) {
				continue // Skip rows with incorrect column count
			}

			isNew, err := strconv.ParseBool(record[3])
			if err != nil {
				continue // Skip invalid boolean value
			}

			move, err := strconv.Atoi(record[4])
			if err != nil {
				continue // Skip invalid integer value
			}

			rows = append(rows, CSVRow{
				Name:        record[0],
				Ring:        record[1],
				Quadrant:    record[2],
				IsNew:       isNew,
				Move:        move,
				Description: record[5],
			})
		}

		c.HTML(http.StatusOK, "result.html", gin.H{
			"Data": CSVData{
				Headers: headers,
				Rows:    rows,
			},
		})
	})

	r.Run(":8080")
}

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}
