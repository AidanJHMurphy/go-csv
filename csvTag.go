package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const (
	tagName    = "csv"
	attrDelim  = ";"
	valueDelim = ":"

	headerAttr          = "header"
	indexAttr           = "index"
	useCustomSetterAttr = "useCustomSetter"
)

var (
	ErrorMissingCustomSetter = fmt.Errorf("cannot use custom data type without implementing CustomSetter interface")
	ErrorUnsupportedDataType = fmt.Errorf("must implement CustomSetter interface when using unsupported data types")
	ErrorInvalidIndex        = fmt.Errorf("index must be a non negative integer")
	ErrorMalformedCsvTag     = fmt.Errorf("you need to specify either the header or index")
	ErrorUnexportedField     = fmt.Errorf("csv tags may not be set on unexported fields")
	ErrorFieldNotFound       = fmt.Errorf("field not found in header")
)

type CustomSetter interface {
	CustomSetter(fieldName string, value string) (err error)
}

type csvAttributes struct {
	headerName      string
	columnIndex     int
	useCustomSetter bool
}

func isValidDataType(i interface{}) bool {
	switch i.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		return true
	}
	return false
}

func getCsvAttributes(structPointer interface{}) (csvAttrs map[string]csvAttributes, err error) {
	csvAttrs = make(map[string]csvAttributes)

	structValue := reflect.ValueOf(structPointer).Elem()
	customDataSetter := reflect.TypeOf((*CustomSetter)(nil)).Elem()
	supportsCustomData := reflect.TypeOf(structPointer).Implements(customDataSetter)

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		tag := field.Tag.Get(tagName)
		if tag == "" {
			continue
		}

		if !field.IsExported() {
			return csvAttrs, CsvTagDefError{
				CsvTag:    tag,
				FieldName: field.Name,
				Err:       ErrorUnexportedField,
			}
		}

		fieldAttrs, err := getAttributesFromTag(tag)
		if err != nil {
			return csvAttrs, CsvTagDefError{
				CsvTag:    tag,
				FieldName: field.Name,
				Err:       err,
			}
		}

		if fieldAttrs.useCustomSetter && !supportsCustomData {
			return csvAttrs, CsvTagDefError{
				CsvTag:    tag,
				FieldName: field.Name,
				Err:       ErrorMissingCustomSetter,
			}
		}

		if !isValidDataType(structValue.FieldByIndex([]int{i}).Interface()) && !supportsCustomData {
			return csvAttrs, CsvTagDefError{
				CsvTag:    tag,
				FieldName: field.Name,
				Err:       ErrorUnsupportedDataType,
			}
		}

		csvAttrs[structValue.Type().Field(i).Name] = fieldAttrs
	}

	return csvAttrs, nil
}

func getAttributesFromTag(tag string) (attrs csvAttributes, err error) {
	attributes := strings.Split(tag, attrDelim)
	var hasHeader = false
	var hasIndex = false

	for _, attribute := range attributes {
		attributeArr := strings.Split(attribute, valueDelim)
		key := attributeArr[0]
		var value string
		if len(attributeArr) > 1 {
			value = attributeArr[1]
		}

		switch key {
		case headerAttr:
			hasHeader = true
			attrs.headerName = value
		case indexAttr:
			hasIndex = true
			attrs.columnIndex, err = strconv.Atoi(value)
			if err != nil {
				return attrs, ErrorInvalidIndex
			}
			if attrs.columnIndex < 0 {
				return attrs, ErrorInvalidIndex
			}
		case useCustomSetterAttr:
			attrs.useCustomSetter = true
		}
	}

	if !hasHeader && !hasIndex {
		return attrs, ErrorMalformedCsvTag
	}

	return attrs, nil
}

type Parser struct {
	reader   *csv.Reader
	line     int
	csvAttrs map[string]csvAttributes
}

type ParserOptions struct {
	Delimiter   rune
	CommentChar rune
	ReuseRecord bool
}

func legalDelimiter(d rune) bool {
	if d == 0 {
		return false
	}
	if d == '\n' {
		return false
	}
	if d == '\r' {
		return false
	}

	return true
}

// NewParser creates a new csv parser for the provided file that supports the csv struct decorator tag.
// Use ParserOptions to specify any desired changed from the default behavior as defined in the standard csv parser library.
func NewParser(file io.Reader, options ParserOptions) (p Parser) {
	p.reader = csv.NewReader(file)
	p.csvAttrs = make(map[string]csvAttributes)

	// Keep default value if zero-value rune is passed in
	if legalDelimiter(options.Delimiter) {
		p.reader.Comma = options.Delimiter
	}

	p.reader.Comment = options.CommentChar

	p.reader.ReuseRecord = options.ReuseRecord

	return p
}

// ParseHeader reads the first line of the parser's csv file and interpret's the data as headers described by the csv decorator tags defined on structPointer.
// The structPointer should be pointer to a struct with csv decorator tags applied.
func (p *Parser) ParseHeader(structPointer interface{}) (err error) {
	header, err := p.reader.Read()

	if err != nil {
		return err
	}

	if len(p.csvAttrs) == 0 {
		p.csvAttrs, err = getCsvAttributes(structPointer)
		if err != nil {
			return err
		}
	}

	for fieldName, csvAttrs := range p.csvAttrs {
		var foundIdx = false

		for idx, headerLabel := range header {
			if headerLabel == csvAttrs.headerName {
				csvAttrs.columnIndex = idx
				p.csvAttrs[fieldName] = csvAttrs
				foundIdx = true
				break
			}
		}

		if !foundIdx {
			return FieldNotFoundError{
				FieldName:  fieldName,
				HeaderName: csvAttrs.headerName,
				Err:        ErrorFieldNotFound,
			}
		}
	}

	return nil
}

// ReadRecord reads the next line of the parser's csv file and interprets the data as described by the csv decorator tags defined on structPointer.
// The structPointer should be pointer to a struct with csv decorator tags applied, and data from the appropriate column in the csv file will be set on the fields of structPointer.
func (p *Parser) ReadRecord(structPointer interface{}) (err error) {

	if len(p.csvAttrs) == 0 {
		p.csvAttrs, err = getCsvAttributes(structPointer)
		if err != nil {
			return err
		}
	}

	p.line++
	readRecord, err := p.reader.Read()

	if err != nil {
		return err
	}

	for fieldName, csvAttrs := range p.csvAttrs {
		idx := csvAttrs.columnIndex
		value := readRecord[idx]
		err := p.setFieldValue(structPointer, fieldName, value)

		if err != nil {
			return SetValueError{
				Line:      p.line,
				Value:     value,
				FieldName: fieldName,
				Err:       err,
			}
		}
	}

	return nil
}

func (p *Parser) setFieldValue(structPointer interface{}, fieldName string, value string) (err error) {
	inStruct := reflect.ValueOf(structPointer)
	field := inStruct.Elem().FieldByName(fieldName)

	if p.csvAttrs[fieldName].useCustomSetter {
		method := inStruct.MethodByName("CustomSetter")
		inputs := make([]reflect.Value, 2)
		inputs[0] = reflect.ValueOf(fieldName)
		inputs[1] = reflect.ValueOf(value)

		out := method.Call(inputs)[0]

		if !out.IsZero() {
			return fmt.Errorf("%v", out)
		}

		return nil
	}

	switch field.Interface().(type) {
	case string:
		field.SetString(value)
	case bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case int, int8, int16, int32, int64:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(intValue))
	case uint, uint8, uint16, uint32, uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case float32:
		floatValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)

	case complex64:
		cmplxValue, err := strconv.ParseComplex(value, 64)
		if err != nil {
			return err
		}
		field.SetComplex(cmplxValue)
	case complex128:
		cmplxValue, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return err
		}
		field.SetComplex(cmplxValue)
	default:
		return ErrorUnsupportedDataType
	}

	return nil
}

type CsvTagDefError struct {
	CsvTag    string
	FieldName string
	Err       error
}

func (e CsvTagDefError) Error() string {
	return fmt.Sprintf("problem with csv tag definition %s on field %s: %v", e.CsvTag, e.FieldName, e.Err)
}

func (e CsvTagDefError) Unwrap() error { return e.Err }

type FieldNotFoundError struct {
	FieldName  string
	HeaderName string
	Err        error
}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf("field %s not found in header with label %s", e.FieldName, e.HeaderName)
}

func (e FieldNotFoundError) Unwrap() error { return e.Err }

type SetValueError struct {
	Line      int
	Value     string
	FieldName string
	Err       error
}

func (e SetValueError) Error() string {
	return fmt.Sprintf("record on line %d: problem setting value %s on field %s: %v", e.Line, e.Value, e.FieldName, e.Err)
}

func (e SetValueError) Unwrap() error { return e.Err }
