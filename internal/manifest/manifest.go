// Package manifest loads batch render jobs from JSON and CSV files.
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Job contains the per-thumbnail values supplied by a batch manifest.
type Job struct {
	Input    string `json:"input"`
	Output   string `json:"output"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Font     string `json:"font"`
	Quality  *int   `json:"quality"`
	Row      int    `json:"-"`
}

var requiredColumns = map[string]bool{
	"input":  true,
	"output": true,
	"title":  true,
}

var allowedColumns = map[string]bool{
	"input":    true,
	"output":   true,
	"title":    true,
	"subtitle": true,
	"font":     true,
	"quality":  true,
}

// Load reads jobs from a JSON or CSV manifest. Relative image paths resolve
// against the directory containing the manifest.
func Load(path string) ([]Job, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening manifest %q: %w", path, err)
	}
	defer file.Close()

	var jobs []Job
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		jobs, err = loadJSON(file)
	case ".csv":
		jobs, err = loadCSV(file)
	default:
		return nil, fmt.Errorf("unsupported manifest extension %q; use .json or .csv", filepath.Ext(path))
	}
	if err != nil {
		return nil, fmt.Errorf("reading manifest %q: %w", path, err)
	}
	if len(jobs) == 0 {
		return nil, errors.New("manifest must contain at least one job")
	}

	directory := filepath.Dir(path)
	for index := range jobs {
		jobs[index].Input = resolvePath(directory, jobs[index].Input)
		jobs[index].Output = resolvePath(directory, jobs[index].Output)
		if jobs[index].Row == 0 {
			jobs[index].Row = index + 1
		}
	}
	return jobs, nil
}

func loadJSON(reader io.Reader) ([]Job, error) {
	var document struct {
		Jobs []Job `json:"jobs"`
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("JSON manifest must contain one document")
	}
	return document.Jobs, nil
}

func loadCSV(reader io.Reader) ([]Job, error) {
	rows, err := readCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("decoding CSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, errors.New("CSV manifest must include a header row")
	}

	columns, err := csvColumns(rows[0])
	if err != nil {
		return nil, err
	}
	jobs := make([]Job, 0, len(rows)-1)
	for index, row := range rows[1:] {
		rowNumber := index + 2
		if len(row) != len(rows[0]) {
			return nil, fmt.Errorf("CSV row %d has %d fields, want %d", rowNumber, len(row), len(rows[0]))
		}
		job, err := jobFromCSV(row, columns, rowNumber)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func readCSV(reader io.Reader) ([][]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var records [][]string
	var record []string
	var field strings.Builder
	quoted := false
	quoteClosed := false
	fieldStarted := false
	line := 1

	appendField := func() {
		record = append(record, field.String())
		field.Reset()
		fieldStarted = false
		quoteClosed = false
	}
	appendRecord := func() {
		appendField()
		records = append(records, record)
		record = nil
	}

	for index := 0; index < len(data); index++ {
		value := data[index]
		if value == '\r' && index+1 < len(data) && data[index+1] == '\n' {
			continue
		}

		if quoted {
			if value == '"' {
				if index+1 < len(data) && data[index+1] == '"' {
					field.WriteByte(value)
					index++
					continue
				}
				quoted = false
				quoteClosed = true
				continue
			}
			field.WriteByte(value)
			if value == '\n' {
				line++
			}
			continue
		}

		if quoteClosed {
			switch value {
			case ',':
				appendField()
			case '\n':
				appendRecord()
				line++
			default:
				return nil, fmt.Errorf("CSV line %d has characters after a closing quote", line)
			}
			continue
		}

		switch value {
		case '"':
			if fieldStarted {
				return nil, fmt.Errorf("CSV line %d has a quote in an unquoted field", line)
			}
			quoted = true
			fieldStarted = true
		case ',':
			appendField()
		case '\n':
			if fieldStarted || len(record) > 0 {
				appendRecord()
			}
			line++
		default:
			field.WriteByte(value)
			fieldStarted = true
		}
	}

	if quoted {
		return nil, fmt.Errorf("CSV line %d has an unclosed quoted field", line)
	}
	if fieldStarted || quoteClosed || len(record) > 0 {
		appendRecord()
	}
	return records, nil
}

func csvColumns(header []string) (map[string]int, error) {
	columns := make(map[string]int, len(header))
	for index, value := range header {
		name := strings.TrimSpace(value)
		if !allowedColumns[name] {
			return nil, fmt.Errorf("unsupported CSV column %q", value)
		}
		if _, exists := columns[name]; exists {
			return nil, fmt.Errorf("CSV column %q appears more than once", name)
		}
		columns[name] = index
	}
	for name := range requiredColumns {
		if _, exists := columns[name]; !exists {
			return nil, fmt.Errorf("CSV manifest is missing required column %q", name)
		}
	}
	return columns, nil
}

func jobFromCSV(row []string, columns map[string]int, rowNumber int) (Job, error) {
	job := Job{
		Input:  row[columns["input"]],
		Output: row[columns["output"]],
		Title:  row[columns["title"]],
		Row:    rowNumber,
	}
	if index, exists := columns["subtitle"]; exists {
		job.Subtitle = row[index]
	}
	if index, exists := columns["font"]; exists {
		job.Font = row[index]
	}
	if index, exists := columns["quality"]; exists && strings.TrimSpace(row[index]) != "" {
		quality, err := strconv.Atoi(row[index])
		if err != nil {
			return Job{}, fmt.Errorf("CSV row %d has invalid quality %q: %w", rowNumber, row[index], err)
		}
		job.Quality = &quality
	}
	return job, nil
}

func resolvePath(directory, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(directory, path)
}
