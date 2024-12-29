package handlers

import (
	"embed"
	"html/template"
	"io"
	"log"
	"os"
)

//go:embed report.template.html
var templateFs embed.FS

// Write HTML Hearbeat report to `wr`
func WriteReport(reports []ServiceReport, wr io.Writer) error {
	t, err := template.ParseFS(templateFs, "report.template.html")
	if err != nil {
		return err
	}

	err = t.Execute(wr, reports)
	if err != nil {
		return err
	}

	return nil
}

func WriteReportToFile(reports []ServiceReport, filename string) error {
	log.Printf("Writing report to %s", filename)
	// Create or truncate the output file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = WriteReport(reports, file)
	return err
}
