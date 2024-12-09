package handlers

import (
	"html/template"
	"io"
)

// Write HTML Hearbeat report to `wr`
func WriteReport(reports []ServiceReport, wr io.Writer) error {
	t, err := template.ParseFiles("report.template.html")
	if err != nil {
		return err
	}

	err = t.Execute(wr, reports)
	if err != nil {
		return err
	}

	return nil
}
