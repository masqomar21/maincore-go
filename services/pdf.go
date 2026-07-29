package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type PDFMargin struct {
	Top    string
	Right  string
	Bottom string
	Left   string
}

type PDFColumn struct {
	Header string
	Key    string
	Width  string
	Align  string
}

type PDFExportOptions struct {
	PageSize          string // A4, Letter, etc
	Orientation       string // portrait, landscape
	PrintBackground   bool
	Margin            *PDFMargin
	PreferCSSPageSize bool
}

type PDFFileExportOptions struct {
	PDFExportOptions
	OutputDir string
	FileName  string
}

type PDFStandardExportOptions struct {
	Title             string
	Columns           []PDFColumn
	Data              []map[string]interface{}
	DefaultEmptyValue string
	PageSize          string
	Orientation       string
	PrintBackground   bool
	PreferCSSPageSize bool
	Margin            *PDFMargin
}

type PDFStandardFileExportOptions struct {
	PDFStandardExportOptions
	OutputDir string
	FileName  string
}



type PdfExportService struct{}

func (s *PdfExportService) generateStandardExportHTML(options PDFStandardExportOptions) string {
	title := options.Title
	if title == "" {
		title = "Data Export"
	}

	var headers strings.Builder
	for _, col := range options.Columns {
		width := col.Width
		if width == "" {
			width = "auto"
		}
		align := col.Align
		if align == "" {
			align = "left"
		}
		headers.WriteString(fmt.Sprintf("<th style=\"width: %s; text-align: %s\">%s</th>", width, align, col.Header))
	}

	var rows strings.Builder
	for _, item := range options.Data {
		rows.WriteString("<tr>")
		for _, col := range options.Columns {
			align := col.Align
			if align == "" {
				align = "left"
			}
			valRaw, exists := item[col.Key]
			val := ""
			if exists && valRaw != nil {
				val = fmt.Sprintf("%v", valRaw)
			}
			if val == "" {
				val = options.DefaultEmptyValue
			}
			rows.WriteString(fmt.Sprintf("<td style=\"text-align: %s\">%s</td>", align, val))
		}
		rows.WriteString("</tr>")
	}

	pageSize := options.PageSize
	if pageSize == "" {
		pageSize = "A4"
	}
	orientation := options.Orientation
	if orientation == "" {
		orientation = "portrait"
	}

	now := time.Now().Format("02/01/2006, 15:04:05")

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>%s</title>
      <style>
        @page {
          margin: 20mm;
          size: %s %s;
        }
        body {
          font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
          margin: 0;
          padding: 20px;
          font-size: 12px;
          line-height: 1.4;
          color: #333;
        }
        .header {
          text-align: center;
          margin-bottom: 30px;
          border-bottom: 2px solid #2F75B5;
          padding-bottom: 15px;
        }
        .header h1 {
          color: #2F75B5;
          font-size: 24px;
          margin: 0;
        }
        .export-info {
          text-align: right;
          font-size: 10px;
          color: #666;
          margin-bottom: 20px;
        }
        .table-container {
          width: 100%%;
        }
        table {
          width: 100%%;
          border-collapse: collapse;
          background: white;
          box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        thead { display: table-header-group; }
        tfoot { display: table-footer-group; }

        th {
          background: #2F75B5;
          color: white;
          font-weight: bold;
          padding: 12px 8px;
          border: 1px solid #1a5490;
          font-size: 11px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }
        td {
          padding: 10px 8px;
          border: 1px solid #ddd;
          font-size: 11px;
          word-wrap: break-word;
        }
        tr:nth-child(even) {
          background-color: #f8f9fa;
        }
        tr:hover {
          background-color: #e3f2fd;
        }
        tr, td, th {
          page-break-inside: auto !important;
        }
        @media print {
          body { padding: 0 }
          .header, th { page-break-after: avoid }
          .table-container { overflow: visible !important; }
        }
      </style>
    </head>
    <body>
      <div class="header">
        <h1>%s</h1>
      </div>
      <div class="export-info">
        Generated on: %s | Total Records: %d
      </div>
      <div class="table-container">
        <table>
          <thead><tr>%s</tr></thead>
          <tbody>%s</tbody>
        </table>
      </div>
    </body>
    </html>
	`, title, pageSize, orientation, title, now, len(options.Data), headers.String(), rows.String())

	return html
}

func (s *PdfExportService) launchBrowser() *rod.Browser {
	return rod.New().MustConnect()
}

func getPrintToPDF(landscape bool, printBackground bool, preferCSSPageSize bool, margin *PDFMargin) *proto.PagePrintToPDF {
	req := &proto.PagePrintToPDF{
		Landscape:         landscape,
		PrintBackground:   printBackground,
		PreferCSSPageSize: preferCSSPageSize,
	}
	return req
}

func (s *PdfExportService) ExportFormPageSourceToBuffer(html string, options *PDFExportOptions) ([]byte, error) {
	browser := s.launchBrowser()
	defer browser.MustClose()

	page := browser.MustPage()
	
	page.MustSetDocumentContent(html)
	page.MustWaitLoad()

	landscape := false
	printBg := true
	preferCss := true
	var margin *PDFMargin

	if options != nil {
		landscape = options.Orientation == "landscape"
		printBg = options.PrintBackground
		preferCss = options.PreferCSSPageSize
		margin = options.Margin
	}

	req := getPrintToPDF(landscape, printBg, preferCss, margin)
	stream, err := page.PDF(req)
	if err != nil {
		return nil, err
	}
	
	buf, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	
	return buf, nil
}

func (s *PdfExportService) ExportFormPageSourceToFile(html string, options *PDFFileExportOptions) (string, error) {
	var exportOptions *PDFExportOptions
	outputDir := "exports"
	fileName := "export.pdf"

	if options != nil {
		exportOptions = &options.PDFExportOptions
		if options.OutputDir != "" {
			outputDir = options.OutputDir
		}
		if options.FileName != "" {
			fileName = options.FileName
		}
	}

	buf, err := s.ExportFormPageSourceToBuffer(html, exportOptions)
	if err != nil {
		return "", err
	}

	cwd, _ := os.Getwd()
	publicDir := filepath.Join(cwd, "public", outputDir)
	os.MkdirAll(publicDir, 0755)

	filePath := filepath.Join(publicDir, fileName)
	err = os.WriteFile(filePath, buf, 0644)
	return filePath, err
}

func (s *PdfExportService) ExportStandardExportToBuffer(options PDFStandardExportOptions) ([]byte, error) {
	html := s.generateStandardExportHTML(options)
	
	exportOpts := &PDFExportOptions{
		PageSize:          options.PageSize,
		Orientation:       options.Orientation,
		PrintBackground:   options.PrintBackground,
		PreferCSSPageSize: options.PreferCSSPageSize,
		Margin:            options.Margin,
	}
	return s.ExportFormPageSourceToBuffer(html, exportOpts)
}

func (s *PdfExportService) ExportStandardExportToFile(options PDFStandardFileExportOptions) (string, error) {
	buf, err := s.ExportStandardExportToBuffer(options.PDFStandardExportOptions)
	if err != nil {
		return "", err
	}

	outputDir := "exports"
	if options.OutputDir != "" {
		outputDir = options.OutputDir
	}

	cwd, _ := os.Getwd()
	publicDir := filepath.Join(cwd, "public", outputDir)
	os.MkdirAll(publicDir, 0755)

	filePath := filepath.Join(publicDir, options.FileName)
	err = os.WriteFile(filePath, buf, 0644)
	return filePath, err
}
