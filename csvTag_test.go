package csv

import (
	"errors"
	"fmt"
	"io"
	"reflect"
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

	typesTestData = `string,boolean,int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,float32,float64,customField
blah,true,-1,-8,-16,-32,-64,1,8,16,32,64,12.8,25.6,value`
)

var (
	headerTestResults = []headerTest{
		{IgnoredField: 0, Field1: "String", Field2: 12, Field3: 123456},
		{IgnoredField: 1, Field1: "OtherString", Field2: 14, Field3: 48484848},
	}
	indexTestResults = []indexTest{
		{Field1: "firstData", Field2: 46},
		{Field1: "second	Data", Field2: 47},
		{Field1: "thirdData", Field2: 48},
	}
	typesTestResults = []dataTypesTest{
		{
			String:      "blah",
			Boolean:     true,
			Int:         int(-1),
			Int8:        int8(-8),
			Int16:       int16(-16),
			Int32:       int32(-32),
			Int64:       int64(-64),
			UInt:        uint(1),
			UInt8:       uint8(8),
			UInt16:      uint16(16),
			UInt32:      uint32(32),
			UInt64:      uint64(64),
			Float32:     float32(12.8),
			Float64:     float64(25.6),
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
	String      string  `csv:"header:string"`
	Boolean     bool    `csv:"header:boolean"`
	Int         int     `csv:"header:int"`
	Int8        int8    `csv:"header:int8"`
	Int16       int16   `csv:"header:int16"`
	Int32       int32   `csv:"header:int32"`
	Int64       int64   `csv:"header:int64"`
	UInt        uint    `csv:"header:uint"`
	UInt8       uint8   `csv:"header:uint8"`
	UInt16      uint16  `csv:"header:uint16"`
	UInt32      uint32  `csv:"header:uint32"`
	UInt64      uint64  `csv:"header:uint64"`
	Float32     float32 `csv:"header:float32"`
	Float64     float64 `csv:"header:float64"`
	CustomField string  `csv:"header:customField;useCustomSetter"`
}

func (dtt *dataTypesTest) CustomSetter(fieldName string, value string) (err error) {
	if fieldName == "CustomField" {
		dtt.CustomField = strings.ToUpper(value) + "!!"
		return nil
	}

	return fmt.Errorf("unexpected call to CustomSetter")
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

		for fieldIndex := 0; fieldIndex < reflect.ValueOf(data).NumField(); fieldIndex++ {
			fieldName := reflect.ValueOf(data).Type().Field(fieldIndex).Name
			dataValue := reflect.ValueOf(data).FieldByName(fieldName)
			expectedValue := reflect.ValueOf(headerTestResults[i]).FieldByName(fieldName)

			if dataValue.Interface() != expectedValue.Interface() {
				t.Errorf("improperly parsed %s data from csv with header. Got '%v' but expected '%v'", fieldName, dataValue, expectedValue)
			}
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
			t.Errorf("encountered error parsing csv without header: %v", err)
			break
		}

		for fieldIndex := 0; fieldIndex < reflect.ValueOf(data).NumField(); fieldIndex++ {
			fieldName := reflect.ValueOf(data).Type().Field(fieldIndex).Name
			dataValue := reflect.ValueOf(data).FieldByName(fieldName)
			expectedValue := reflect.ValueOf(indexTestResults[i]).FieldByName(fieldName)

			if dataValue.Interface() != expectedValue.Interface() {
				t.Errorf("improperly parsed %s data from csv with header. Got '%v' but expected '%v'", fieldName, dataValue, expectedValue)
			}
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

		for fieldIndex := 0; fieldIndex < reflect.ValueOf(data).NumField(); fieldIndex++ {
			fieldName := reflect.ValueOf(data).Type().Field(fieldIndex).Name
			dataValue := reflect.ValueOf(data).FieldByName(fieldName)
			expectedValue := reflect.ValueOf(typesTestResults[i]).FieldByName(fieldName)

			if dataValue.Interface() != expectedValue.Interface() {
				t.Errorf("improperly parsed %s data from csv with header. Got '%v' but expected '%v'", fieldName, dataValue, expectedValue)
			}
		}
	}
}

type MissingCustomSetter struct {
	CustomField string `csv:"header:field1;useCustomSetter"`
}

func TestCustomSetterInterfaceError(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&MissingCustomSetter{})
	if err == nil {
		t.Errorf("expected to encounter Missing Custom Setter error, but got none")
	}
	if !errors.Is(err, ErrorMissingCustomSetter) {
		t.Errorf("expected to encounter Missing Custom Setter error, but got %v", err)
	}
}

type UnsupportedDataType1 struct {
	UnsupportedField interface{} `csv:"header:field1"`
}

func TestUnsupportedDataTypeError1(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&UnsupportedDataType1{})
	if err == nil {
		t.Errorf("expected to encounter Unsupported Data Type error, but got none")
	}
	if !errors.Is(err, ErrorUnsupportedDataType) {
		t.Errorf("expected to encounter Unsupported Data Type error, but got %v", err)
	}
}

type UnsupportedDataType2 struct {
	UnsupportedField interface{} `csv:"header:field1"`
}

func (UnsupportedDataType2) CustomSetter(fieldName string, value string) (err error) {
	return nil
}

func TestUnsupportedDataTypeError2(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&UnsupportedDataType2{})
	fmt.Println(err)
	if err != nil {
		t.Errorf("expected to not encounter an error, but got %v", err)
	}
}

type InvalidIndex1 struct {
	AlphaIndex string `csv:"index:a"`
}

func TestInvalidIndexError1(t *testing.T) {
	p := NewParser(strings.NewReader(indexTestData), ParserOptions{Delimiter: '\t'})

	err := p.ReadRecord(&InvalidIndex1{})
	fmt.Println(err)
	if err == nil {
		t.Errorf("expected to encounter Invalid Index error, but got none")
	}
	if !errors.Is(err, ErrorInvalidIndex) {
		t.Errorf("expected to encounter Invalid Index error, but got %v", err)
	}
}

type InvalidIndex2 struct {
	AlphaIndex string `csv:"index:-1"`
}

func TestInvalidIndexError2(t *testing.T) {
	p := NewParser(strings.NewReader(indexTestData), ParserOptions{Delimiter: '\t'})

	err := p.ReadRecord(&InvalidIndex2{})
	fmt.Println(err)
	if err == nil {
		t.Errorf("expected to encounter Invalid Index error, but got none")
	}
	if !errors.Is(err, ErrorInvalidIndex) {
		t.Errorf("expected to encounter Invalid Index error, but got %v", err)
	}
}

type MalformedTag struct {
	BadDef string `csv:"Header:field1"`
}

func TestMalformedTagError(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&MalformedTag{})
	if err == nil {
		t.Errorf("expected to encounter Malformed Tag error, but got none")
	}
	if !errors.Is(err, ErrorMalformedCsvTag) {
		t.Errorf("expected to encounter Malformed Tag error, but got %v", err)
	}
}

type UnexportedField struct {
	unexportedField string `csv:"header:field1"`
}

func TestUnexportedFieldError(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&UnexportedField{})
	if err == nil {
		t.Errorf("expected to encounter Unexported Field error, but got none")
	}
	if !errors.Is(err, ErrorUnexportedField) {
		t.Errorf("expected to encounter Unexported Field error, but got %v", err)
	}
}

type FieldNotFound struct {
	Field1 string `csv:"header:thiswontbefound"`
}

func TestFieldNotFoundError(t *testing.T) {
	p := NewParser(strings.NewReader(headerTestData), ParserOptions{})

	err := p.ParseHeader(&FieldNotFound{})
	if err == nil {
		t.Errorf("expected to encounter Field Not Found error, but got none")
	}
	if !errors.Is(err, ErrorFieldNotFound) {
		t.Errorf("expected to encounter Field Not Found error, but got %v", err)
	}
}
