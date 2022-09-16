package csv

import (
	"io"
	"strings"
	"testing"
)

const (
	csvFileWithHeader = `client,rptid,uselessGarbage,custid
Xcel,12,asdf65434,123456
Dom,14, f8jf8j,48484848`

	csvFileWithoutHeader = `asdlkeim	firstData	asdon	46	afd&&svdsaf	4g5245g254
asd5g4lkeim	"second	Data"	a5g5on	47	afd&&5h67af	4g5sbg254
asdlk654eim	thirdData	a$&*^on	48	a$%&*af	4g5254654`
)

var (
	dataFromCsvWithHeader = []headerTest{
		{WhiteLabel: "Xcel", ReportId: 12, CustomerId: 123456},
		{WhiteLabel: "Dom", ReportId: 14, CustomerId: 48484848},
	}
	dataFromCsvWithoutHeader = []noHeaderTest{
		{Field1: "firstData", Field2: 46},
		{Field1: "second	Data", Field2: 47},
		{Field1: "thirdData", Field2: 48},
	}
)

type headerTest struct {
	IgnoredField int
	WhiteLabel   string `csv:"header:client"`
	ReportId     int    `csv:"header:rptid"`
	CustomerId   int    `csv:"header:custid"`
	CustomField  string `csv:"header:uselessGarbage;useCustomSetter"`
}

func (h *headerTest) CustomSetter(fieldName string, value string) (err error) {
	if fieldName == "CustomField" {
		h.CustomField = "you passed the test"
	}

	return nil
}

type noHeaderTest struct {
	Field1 string `csv:"index:1"`
	Field2 int    `csv:"index:3"`
}

func TestCsvWithHeaders(t *testing.T) {
	p := NewParser(strings.NewReader(csvFileWithHeader), ParserOptions{})

	data := headerTest{}
	err := p.ParseHeader(&headerTest{})
	if err != nil {
		t.Errorf("encountered error parsing csv header: %v", err)
	}

	for i := 0; true; i++ {
		data.IgnoredField = i
		data.CustomField = "not set"
		err := p.ReadRecord(&data)

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}

		if data.IgnoredField != i {
			t.Errorf("expected IgnoredField to be %d but got %d", i, data.IgnoredField)
		}

		if data.WhiteLabel != dataFromCsvWithHeader[i].WhiteLabel {
			t.Errorf("improperly parsed WhiteLabel data from csv with header. Got '%v' but expected '%v'", data.WhiteLabel, dataFromCsvWithHeader[i].WhiteLabel)
		}
		if data.ReportId != dataFromCsvWithHeader[i].ReportId {
			t.Errorf("improperly parsed ReportId data from csv with header. Got '%v' but expected '%v'", data.ReportId, dataFromCsvWithHeader[i].ReportId)
		}
		if data.CustomerId != dataFromCsvWithHeader[i].CustomerId {
			t.Errorf("improperly parsed CustomerId data from csv with header. Got '%v' but expected '%v'", data.CustomerId, dataFromCsvWithHeader[i].CustomerId)
		}

		if data.CustomField != "you passed the test" {
			t.Errorf("Custom Field not properly set")
		}
	}
}

func TestCsvWithoutHeaders(t *testing.T) {
	p := NewParser(strings.NewReader(csvFileWithoutHeader),
		ParserOptions{
			Delimiter: '\t',
		},
	)

	data := noHeaderTest{}

	for i := 0; true; i++ {
		err := p.ReadRecord(&data)

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}

		if data.Field1 != dataFromCsvWithoutHeader[i].Field1 {
			t.Errorf("improperly parsed Field1 data from csv without header. Got '%v' but expected '%v'", data.Field1, dataFromCsvWithoutHeader[i].Field1)
		}
		if data.Field2 != dataFromCsvWithoutHeader[i].Field2 {
			t.Errorf("improperly parsed Field2 data from csv without header. Got '%v' but expected '%v'", data.Field2, dataFromCsvWithoutHeader[i].Field2)
		}
	}
}

/*
TODO -- add tests for the various error cases:
 - csv tag definition error
 - malformed header error
 - field not found error
 -
*/
