# go-csv
This project defines a csv struct decorator tag, and wraper for the base csv package that can understand and use the tag in order to more easily parse csv data.

## How to decorate a struct with the csv tag

If your csv file uses a header, define the mapping with the header attribute. For example, given the following struct definition:

```
type csvWithHeader struct {
	Field1       string `csv:"header:field1Header"`
	Field2       int    `csv:"header:field2Header"`
	Field3       int    `csv:"header:Field3Header"`
}
```

... and the folling csv-formatted string

```
`field1Header,ignoredColumn,field3Header,field2Header
value1,thisIsUnusedData,3,2`
```

... we would retrieve the following data

```
data := []csvWithHeader [
  {
    Field1: "value1",
    Field2: 2,
    Field3: 3,
  },
]
```

If your csv file does not use a header, define the mapping with the index attribute. The index is a zero-indexed integer. For example, given the following struct definition:

```
type csvWithIndex struct {
	Field1 string `csv:"index:0"`
	Field2 int    `csv:"index:3"`
}
```

... and the folling csv-formatted string

```
`10,11,12,13,14`
```

... we would retrieve the following data

```
data := []csvWithIndex [
  {
    Field1: 10,
    Field2: 13,
  },
]
```

Fields in a struct that don't have the csv tag applied will be skipped over.

```
type ignoredField struct {
  ThisFieldWontBeWrittenTo string
  ThisFieldWill `csv:"index:0"`
}
```

Fields must be exported for the csv tag to work. The following struct definition is invalid:

```
type improperlyAppliedTag struct {
  unexportedField `csv:"index:0"`
}
```


CSV tags must defined either the header, or the index attribute to be valid. The following struct definition is invalid:

```
type improperlyDefinedTag struct {
  InvalidTag `csv:""`
}
```

If you are setting data that needs additional handling beyond the default, or you are setting a data type that isn't supported, implement the CustomSetter interface for your struct. For example, given the following struct definition:

```
type implementsCustomSetter struct {
  CustomField1 string `csv:"index:0;useCustomSetter"`
  CustomField2 string `csv:"index:0;useCustomSetter"`
}

func (isc *implementsCustomSetter) CustomSetter(fieldName string, value string) (err error) {
  if fieldName = "CustomField1" {
    isc.CustomField1 = strings.ToUpper(value)
    return nil
  }
  
  if fieldName = "CustomField2" {
    isc.CustomField2 = ToLower(value)
    return nil
  }
  
  return fmt.Errorf("custom setter called for unexpected field")
}
```

... and the folling csv-formatted string

```
`hErE Is SoMe wOnKeY DaTa
hErE Is SoMe mOrE WoNkEy dAtA`
```

... we would retrieve the following data

```
data := []implementsCustomSetter [
  {
    CustomField1: "HERE IS SOME WONKEY DATA",
    CustomField2: "here is some wonkey data",
  },
  {
    CustomField1: "HERE IS SOME MORE WONKEY DATA",
    CustomField2: "here is some more wonkey data",
  },
]

```

## How to parse csv data
Once you have defined a struct with csv tags, you'll need to create a new csv parser for the file you want to parse. Then, if your data uses headers, parse the header.
Once you have done that, read the csv data into your struct.

Here is an example with headers:

```
import (
	csv "github.com/AidanJHMurphy/go-csv"
)

const csvWithHeaderData = `field1Header,ignoredColumn,field3Header,field2Header
value1,thisIsUnusedData,3,2`

type csvWithHeader struct {
	Field1       string `csv:"header:field1Header"`
	Field2       int    `csv:"header:field2Header"`
	Field3       int    `csv:"header:Field3Header"`
}

func main() {
  p := csv.NewParser(strings.NewReader(csvWithHeaderData), ParserOptions{}))
    
  err := p.ParseHeader(&csvWithIndex{})
	if err != nil {
		t.Errorf("encountered error parsing csv header: %v", err)
	}
  
  for {
    data := csvWithHeader{}
    err := p.ReadRecord(&data)
    if err == io.EOF {
			break
		}
    
    if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}
  }
}
```

Here is an example without headers:


```
import (
	csv "github.com/AidanJHMurphy/go-csv"
)

const csvWithoutHeaderData = `10,11,12,13,14
20,21,22,23,24`

type csvWithoutHeader struct {
	Field1 string `csv:"index:0"`
	Field2 int    `csv:"index:3"`
}

func main() {
  p := NewParser(strings.NewReader(csvWithoutHeaderData), ParserOptions{}))
  
  for {
    data := csvWithoutHeader{}
    err := p.ReadRecord(&data)
    if err == io.EOF {
			break
		}
    
    if err != nil {
			t.Errorf("encountered error parsing csv with header: %v", err)
			break
		}
  }
}
```
