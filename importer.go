package gofpdi

import "io"

// The Importer class to be used by a pdf generation library
type Importer struct {
	reader *PdfReader
	writer *PdfWriter
	tpl    *TplInfo
}

type TplInfo struct {
	SourceFile string
	Writer     *PdfWriter
	TemplateId int
}

func (this *Importer) GetReader() *PdfReader {
	return this.reader
}

func (this *Importer) GetWriter() *PdfWriter {
	return this.writer
}

func NewImporter() *Importer {
	importer := &Importer{}
	importer.init()

	return importer
}

func (this *Importer) init() {
	this.writer, _ = NewPdfWriter(nil)
}

func (this *Importer) SetSourceFile(f io.ReadSeeker) {
	reader, err := NewPdfReader(f)
	if err != nil {
		panic(err)
	}
	writer, err := NewPdfWriter(nil)
	writer.SetTplIdOffset(0)
	writer.r = reader
}

func (this *Importer) GetPageSizes() map[int]map[string]map[string]float64 {
	result, err := this.GetReader().getAllPageBoxes(1.0)

	if err != nil {
		panic(err)
	}

	return result
}

func (this *Importer) ImportPage(pageno int, box string) int {
	res, err := this.GetWriter().ImportPage(this.GetReader(), pageno, box)
	if err != nil {
		panic(err)
	}

	// Set tpl info
	this.tpl = &TplInfo{TemplateId: res, Writer: this.GetWriter()}

	return 1
}

func (this *Importer) SetNextObjectID(objId int) {
	this.GetWriter().SetNextObjectID(objId)
}

// Put form xobjects and get back a map of template names (e.g. /GOFPDITPL1) and their object ids (int)
func (this *Importer) PutFormXobjects() map[string]int {
	res := make(map[string]int, 0)
	tplNamesIds, err := this.GetWriter().PutFormXobjects(this.GetReader())
	if err != nil {
		panic(err)
	}
	for tplName, pdfObjId := range tplNamesIds {
		res[tplName] = pdfObjId.id
	}
	return res
}

// Put form xobjects and get back a map of template names (e.g. /GOFPDITPL1) and their object ids (sha1 hash)
func (this *Importer) PutFormXobjectsUnordered() map[string]string {
	this.GetWriter().SetUseHash(true)
	res := make(map[string]string, 0)
	tplNamesIds, err := this.GetWriter().PutFormXobjects(this.GetReader())
	if err != nil {
		panic(err)
	}
	for tplName, pdfObjId := range tplNamesIds {
		res[tplName] = pdfObjId.hash
	}
	return res
}

// Get object ids (int) and their contents (string)
func (this *Importer) GetImportedObjects() map[int]string {
	res := make(map[int]string, 0)
	pdfObjIdBytes := this.GetWriter().GetImportedObjects()
	for pdfObjId, bytes := range pdfObjIdBytes {
		res[pdfObjId.id] = string(bytes)
	}
	return res
}

// Get object ids (sha1 hash) and their contents ([]byte)
// The contents may have references to other object hashes which will need to be replaced by the pdf generator library
// The positions of the hashes (sha1 - 40 characters) can be obtained by calling GetImportedObjHashPos()
func (this *Importer) GetImportedObjectsUnordered() map[string][]byte {
	res := make(map[string][]byte, 0)
	pdfObjIdBytes := this.GetWriter().GetImportedObjects()
	for pdfObjId, bytes := range pdfObjIdBytes {
		res[pdfObjId.hash] = bytes
	}
	return res
}

// Get the positions of the hashes (sha1 - 40 characters) within each object, to be replaced with
// actual objects ids by the pdf generator library
func (this *Importer) GetImportedObjHashPos() map[string]map[int]string {
	res := make(map[string]map[int]string, 0)
	pdfObjIdPosHash := this.GetWriter().GetImportedObjHashPos()
	for pdfObjId, posHashMap := range pdfObjIdPosHash {
		res[pdfObjId.hash] = posHashMap
	}
	return res
}

// For a given template id (returned from ImportPage), get the template name (e.g. /GOFPDITPL1) and
// the 4 float64 values necessary to draw the template a x,y for a given width and height.
func (this *Importer) UseTemplate(tplid int, _x float64, _y float64, _w float64, _h float64) (string, float64, float64, float64, float64) {
	// Look up template id in importer tpl map
	tplInfo := this.tpl
	return tplInfo.Writer.UseTemplate(tplInfo.TemplateId, _x, _y, _w, _h)
}
