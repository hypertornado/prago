package utils

import (
	stripmd "github.com/writeas/go-strip-markdown"
)

func filterMarkdown(in string) string {

	return stripmd.Strip(in)

	/*ret := string(blackfriday.Markdown(
		[]byte(in),
		NewPlaintextRenderer(),
		0,
	))
	ret = strings.Trim(ret, " ")
	ret = strings.Replace(ret, "\n", " ", -1)
	ret = strings.Replace(ret, "---", "—", -1)
	ret = strings.Replace(ret, "--", "–", -1)
	return ret*/
}

//CropMarkdown remove all markdown special characters
func CropMarkdown(text string, count int) string {
	text = filterMarkdown(text)
	return Crop(text, count)
}

/*func NewPlaintextRenderer() blackfriday.Renderer {
	return PlaintextRendered{}
	//return blackfriday.PlainText{}
}

type PlaintextRendered struct{}

func (PlaintextRendered) AutoLink(out *bytes.Buffer, link []byte, kind int) {}
func (PlaintextRendered) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.WriteString(" ")
	out.Write(text)
}
func (PlaintextRendered) BlockHtml(out *bytes.Buffer, text []byte) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) CodeSpan(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (PlaintextRendered) DocumentFooter(out *bytes.Buffer) {}
func (PlaintextRendered) DocumentHeader(out *bytes.Buffer) {}
func (PlaintextRendered) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}
func (PlaintextRendered) Emphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (PlaintextRendered) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (PlaintextRendered) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {}

func (PlaintextRendered) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {}

func (PlaintextRendered) Footnotes(out *bytes.Buffer, text func() bool) {}

func (PlaintextRendered) GetFlags() int           { return 0 }
func (PlaintextRendered) HRule(out *bytes.Buffer) {}
func (PlaintextRendered) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	out.WriteString(" ")
	text()
}
func (PlaintextRendered) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {}

func (PlaintextRendered) LineBreak(out *bytes.Buffer) {}
func (PlaintextRendered) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.Write(content)
}

func (PlaintextRendered) List(out *bytes.Buffer, text func() bool, flags int) {
	out.WriteString(" ")
	text()
}
func (PlaintextRendered) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (PlaintextRendered) Paragraph(out *bytes.Buffer, text func() bool) {
	out.WriteString(" ")
	text()
}
func (PlaintextRendered) RawHtmlTag(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (PlaintextRendered) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (PlaintextRendered) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {}
func (PlaintextRendered) TableRow(out *bytes.Buffer, text []byte) {
	out.WriteString(" ")
	out.Write(text)
}
func (PlaintextRendered) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) TableCell(out *bytes.Buffer, text []byte, align int) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) TitleBlock(out *bytes.Buffer, text []byte) {
	out.WriteString(" ")
	out.Write(text)
}

func (PlaintextRendered) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.Write(text)
}*/
