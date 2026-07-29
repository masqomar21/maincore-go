package controllers

import (
	"maincore_go/services"

	"github.com/gin-gonic/gin"
)

// TestDownloadPDF generates a sample PDF and streams it directly to the browser
func TestDownloadPDF(c *gin.Context) {
	pdfService := &services.PdfExportService{}

	options := services.PDFStandardExportOptions{
		Title:       "Laporan Pengguna Contoh",
		PageSize:    "A4",
		Orientation: "portrait",
		Columns: []services.PDFColumn{
			{Header: "Nama", Key: "name", Width: "35%"},
			{Header: "Email", Key: "email", Width: "45%"},
			{Header: "Status", Key: "status", Width: "20%"},
		},
		Data: []map[string]interface{}{
			{"name": "Budi Santoso", "email": "budi@example.com", "status": "Aktif"},
			{"name": "Siti Aminah", "email": "siti@example.com", "status": "Non-Aktif"},
			{"name": "Ahmad Dani", "email": "ahmad@example.com", "status": "Aktif"},
		},
		DefaultEmptyValue: "-",
	}

	buffer, err := pdfService.ExportStandardExportToBuffer(options)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat PDF: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="test-report.pdf"`)
	c.Data(200, "application/pdf", buffer)
}
