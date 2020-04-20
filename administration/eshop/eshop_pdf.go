package eshop

import (
	"bytes"

	"github.com/hypertornado/gofpdf"
	"github.com/skip2/go-qrcode"
)

type PDFCouponData struct {
	Name        string
	Description string
	QRCode      string
}

func generatePDFCoupon(items []PDFCouponData) []byte {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddFont("Calligra", "", "public/Arial.json")

	for _, data := range items {
		pdf.AddPage()
		png, err := qrcode.Encode(data.QRCode, qrcode.Medium, 256)
		if err != nil {
			panic(err)
		}

		pdf.RegisterImageReader("qrCode", "png", bytes.NewReader(png))
		pdf.Image("qrCode", 187, 10, 100, 100, false, "png", 0, "")

		pdf.SetFont("Calligra", "", 16)
		pdf.SetCellMargin(0)

		var rectX float64 = 55
		var rectY float64 = 40
		var rectWidth float64 = 140
		var rectHeight float64 = 125

		var textSize float64 = 19
		var largeTextRatio float64 = 1.4
		var minTextSize float64 = 8

		var margin, titleSize, subtitleSize float64
		for ; textSize > minTextSize; textSize -= 1 {
			titleSize = getPdfHeight(pdf, data.Name, textSize*largeTextRatio, rectWidth)
			subtitleSize = getPdfHeight(pdf, data.Description, textSize, rectWidth)
			if titleSize+subtitleSize < rectHeight {
				margin = (rectHeight - (titleSize + subtitleSize)) / 2
				break
			}
		}

		drawTextBox(pdf, data.Name, textSize*largeTextRatio, rectWidth, rectX, rectY+margin)
		drawTextBox(pdf, data.Description, textSize, rectWidth, rectX, rectY+margin+titleSize)

	}

	buffer := bytes.NewBuffer(nil)
	err := pdf.Output(buffer)
	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func getPdfLineHeight(textSize float64) float64 {
	return textSize / 2
}

func drawTextBox(pdf *gofpdf.Fpdf, text string, textSize, width, x, y float64) {
	tf := pdf.UnicodeTranslatorFromDescriptor("cp1250")

	pdf.SetTextColor(38, 153, 214)
	pdf.SetFontSize(textSize)

	lineHeight := getPdfLineHeight(textSize)

	lines := pdf.SplitLines([]byte(tf(text)), width)
	for k, v := range lines {
		pdf.MoveTo(x, y+(float64(k)*lineHeight))
		pdf.CellFormat(width, lineHeight, string(v), "0", 0, "L", false, 0, "")
	}
}

func getPdfHeight(pdf *gofpdf.Fpdf, text string, textSize, width float64) float64 {
	tf := pdf.UnicodeTranslatorFromDescriptor("cp1250")
	lineHeight := getPdfLineHeight(textSize)
	pdf.SetFontSize(textSize)
	lines := pdf.SplitLines([]byte(tf(text)), width)
	return lineHeight * float64(len(lines))
}
