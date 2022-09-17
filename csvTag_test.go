package csv

import (
	"io"
	"strings"
	"testing"
)

const (
	headerTestData = `field1,fieldTwo,uselessGarbage,Field3
String,12,asdf65434,123456
OtherString,14, f8jf8j,48484848`

	indexTestData = `asdlkeim	firstData	asdon	46	afd&&svdsaf	4g5245g254
asd5g4lkeim	"second	Data"	a5g5on	47	afd&&5h67af	4g5sbg254
asdlk654eim	thirdData	a$&*^on	48	a$%&*af	4g5254654`

	typesTestData = `string,int,int8,int16,int32,int64,customField
blah,1,8,16,32,64,value`
)

var (
	headerTestResults = []headerTest{
		{Field1: "String", Field2: 12, Field3: 123456},
		{Field1: "OtherString", Field2: 14, Field3: 48484848},
	}
	indexTestResults = []indexTest{
		{Field1: "firstData", Field2: 46},
		{Field1: "second	Data", Field2: 47},
		{Field1: "thirdData", Field2: 48},
	}
	typesTestResults = []dataTypesTest{
		{
			String:      "blah",
			Int:         1,
			Int8:        int8(8),
			Int16:       int16(16),
			Int32:       int32(32),
			Int64:       int64(64),
			Float32:     float32(12.8),
			Float64:     float64(12.8),
			CustomField: "VALUE!!",
		},
	}
)

type headerTest struct {
	IgnoredField int
	Field1       string `csv:"header:field1"`
	Field2       int    `csv:"header:fieldTwo"`
	Field3       int    `csv:"header:Field3"`
}

type indexTest struct {
	Field1 string `csv:"index:1"`
	Field2 int    `csv:"index:3"`
}

type dataTypesTest struct {
	String      string `csv:"header:string"`
	Int         int    `csv:"header:int"`
	Int8        int8   `csv:"header:int8"`
	Int16       int16  `csv:"header:int16"`
	Int32       int32  `csv:"header:int32"`
	Int64       int64  `csv:"header:int64"`
	Float32     float32
	Float64     float64
	CustomField string `csv:"header:customField;useCustomSetter"`
}

func (dtt *dataTypesTest) CustomSetter(fieldName string, value string) (err error) {
	if fieldName == "customField" {
		dtt.CustomField = strings.ToUpper(value) + "!!"
	}

	return nil
}

func TestCsvWithHeaders(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	data := headerTest{}
	err := p.ParseHeader(&headerTest{})
	if err != nil {
		t.Errorf("encountered error parsing csv header: %v", err)
	}

	for i := 0; true; i++ {
		data.IgnoredField = i
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

		if data.Field1 != headerTestResults[i].Field1 {
			t.Errorf("improperly parsed Field1 data from csv with header. Got '%v' but expected '%v'", data.Field1, headerTestResults[i].Field1)
		}
		if data.Field2 != headerTestResults[i].Field2 {
			t.Errorf("improperly parsed Field2 data from csv with header. Got '%v' but expected '%v'", data.Field2, headerTestResults[i].Field2)
		}
		if data.Field3 != headerTestResults[i].Field3 {
			t.Errorf("improperly parsed Field3 data from csv with header. Got '%v' but expected '%v'", data.Field3, headerTestResults[i].Field3)
		}
	}
}

func TestCsvWithoutHeaders(t *testing.T) {
	p := NewParser(strings.NewReader(indexTestData),
		ParserOptions{
			Delimiter: '\t',
		},
	)

	data := indexTest{}

	for i := 0; true; i++ {
		err := p.ReadRecord(&data)

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}

		if data.Field1 != indexTestResults[i].Field1 {
			t.Errorf("improperly parsed Field1 data from csv without header. Got '%v' but expected '%v'", data.Field1, indexTestResults[i].Field1)
		}
		if data.Field2 != indexTestResults[i].Field2 {
			t.Errorf("improperly parsed Field2 data from csv without header. Got '%v' but expected '%v'", data.Field2, indexTestResults[i].Field2)
		}
	}
}

func TestCsvDataTypes(t *testing.T) {
	p := NewParser(strings.NewReader(typesTestData), ParserOptions{})

	data := dataTypesTest{}
	err := p.ParseHeader(&dataTypesTest{})
	if err != nil {
		t.Errorf("encountered error parsing csv header: %v", err)
	}

	for i := 0; true; i++ {
		err := p.ReadRecord(&data)

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}

		if data.String != typesTestResults[i].String {
			t.Errorf("improperly parsed stringData data from csv with header. Got '%v' but expected '%v'", data.String, typesTestResults[i].String)
		}

		if data.Int != typesTestResults[i].Int {
			t.Errorf("improperly parsed intData data from csv with header. Got '%v' but expected '%v'", data.Int, typesTestResults[i].Int)
		}

		if data.Int8 != typesTestResults[i].Int8 {
			t.Errorf("improperly parsed int8Data data from csv with header. Got '%v' but expected '%v'", data.Int8, typesTestResults[i].Int8)
		}

		if data.Int16 != typesTestResults[i].Int16 {
			t.Errorf("improperly parsed int16Data data from csv with header. Got '%v' but expected '%v'", data.Int16, typesTestResults[i].Int16)
		}

		if data.Int32 != typesTestResults[i].Int32 {
			t.Errorf("improperly parsed int32Data data from csv with header. Got '%v' but expected '%v'", data.Int32, typesTestResults[i].Int32)
		}

		if data.Int64 != typesTestResults[i].Int64 {
			t.Errorf("improperly parsed int64Data data from csv with header. Got '%v' but expected '%v'", data.Int64, typesTestResults[i].Int64)
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
