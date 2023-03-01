包csvutil在CSV和Go（golang）值之间提供快速、惯用和无依赖的映射。
这个包不是 CSV 解析器，它基于由例如实现的 Reader 和 Writer 接口。 std Go (golang) csv 包。这提供了选择任何其他可能性能更高的 CSV 写入器或读取器的可能性。

Installation
go get github.com/jszwec/csvutil


Requirements
Go1.8+


Example
1.Unmarshal
Unmarshal 使用的是 Go std csv.Reader 及其默认选项。将解码器用于流媒体和更高级的用例。
var csvInput = []byte(`
	name,age,CreatedAt
	jacek,26,2012-04-01T15:00:00Z
	john,,0001-01-01T00:00:00Z`,
)
type User struct {
	Name      string `csv:"name"`
	Age       int    `csv:"age,omitempty"`
	CreatedAt time.Time
}
var users []User
if err := csvutil.Unmarshal(csvInput, &users); err != nil {
	fmt.Println("error:", err)
}
for _, u := range users {
	fmt.Printf("%+v\n", u)
}
// Output:
// {Name:jacek Age:26 CreatedAt:2012-04-01 15:00:00 +0000 UTC}
// {Name:john Age:0 CreatedAt:0001-01-01 00:00:00 +0000 UTC}

2.Marshal
Marshal 使用 Go std csv.Writer 及其默认选项。使用 Encoder 进行流式传输或使用不同的 Writer。
type Address struct {
	City    string
	Country string
}
type User struct {
	Name 	  string
	Address
	Age       int `csv:"age,omitempty"`
	CreatedAt time.Time
}
users := []User{
	{
		Name:      "John",
		Address:   Address{"Boston", "USA"},
		Age:       26,
		CreatedAt: time.Date(2010, 6, 2, 12, 0, 0, 0, time.UTC),
	},
	{
		Name:    "Alice",
		Address: Address{"SF", "USA"},
	},
}
b, err := csvutil.Marshal(users)
if err != nil {
	fmt.Println("error:", err)
}
fmt.Println(string(b))
// Output:
// Name,City,Country,age,CreatedAt
// John,Boston,USA,26,2010-06-02T12:00:00Z
// Alice,SF,USA,,0001-01-01T00:00:00Z

3.Unmarshal and metadata 反序列化和元数据
您的 CSV 输入可能不会始终具有相同的标题。除了您的基本字段之外，您可能会获得您仍想存储的额外元数据。 Decoder 提供了 Unused 方法，在每次调用
Decode 后，它可以报告在解码过程中哪些标头索引没有被使用。基于此，可以处理和存储所有这些额外值。
type User struct {
	Name      string            `csv:"name"`
	City      string            `csv:"city"`
	Age       int               `csv:"age"`
	OtherData map[string]string `csv:"-"`
}
csvReader := csv.NewReader(strings.NewReader(`
	name,age,city,zip
	alice,25,la,90005
	bob,30,ny,10005`)
)
dec, err := csvutil.NewDecoder(csvReader)
if err != nil {
	log.Fatal(err)
}
header := dec.Header()
var users []User
for {
	u := User{OtherData: make(map[string]string)}
	if err := dec.Decode(&u); err == io.EOF {
		break
	} else if err != nil {
		log.Fatal(err)
	}
	for _, i := range dec.Unused() {
		u.OtherData[header[i]] = dec.Record()[i]
	}
	users = append(users, u)
}
fmt.Println(users)
// Output:
// [{alice la 25 map[zip:90005]} {bob ny 30 map[zip:10005]}]

4.But my CSV file has no header...	但我的CSV文件没有标题。。。
有些CSV文件没有标头，但如果您知道它应该是什么样子，就可以定义一个结构并生成它。剩下要做的就是把它传给解码器。
type User struct {
	ID   int
	Name string
	Age  int 	`csv:",omitempty"`
	City string
}
csvReader := csv.NewReader(strings.NewReader(`
1,John,27,la
2,Bob,,ny`))
// in real application this should be done once in init function.//在实际应用程序中，这应该在init函数中完成一次。
userHeader, err := csvutil.Header(User{}, "csv")
if err != nil {
	log.Fatal(err)
}
dec, err := csvutil.NewDecoder(csvReader, userHeader...)
if err != nil {
	log.Fatal(err)
}
var users []User
for {
	var u User
	if err := dec.Decode(&u); err == io.EOF {
		break
	} else if err != nil {
		log.Fatal(err)
	}
	users = append(users, u)
}
fmt.Printf("%+v", users)
// Output:
// [{ID:1 Name:John Age:27 City:la} {ID:2 Name:Bob Age:0 City:ny}]

5.Decoder.Map - data normalization	解码器.映射-数据规范化
解码器的映射功能是一个强大的工具，可以在实际解码发生之前帮助清理或规范传入数据。
假设我们想解码一些浮点值，csv输入包含一些NaN值，但这些值由“n/a”字符串表示。尝试将“n/a”解码为float将以错误告终，因为strconv.ParseFloat需要
“NaN”。知道了这一点，我们可以实现一个Map函数，该函数将规范化我们的“n/a”字符串，并仅针对浮点类型将其转换为“NaN”。
dec, err := NewDecoder(r)
if err != nil {
	log.Fatal(err)
}
dec.Map = func(field, column string, v interface{}) string {
	if _, ok := v.(float64); ok && field == "n/a" {
		return "NaN"
	}
	return field
}
现在，我们的float64字段将被正确解码为NaN。那么float32、float类型别名和其他NaN格式呢？看看这里的完整示例。

6.Different separator/delimiter	不同的分隔符/分隔符
某些文件可能使用不同的值分隔符，例如TSV文件将使用\t。下面的示例展示了如何为这种用例设置解码器和编码器。
Decoder解码器:
csvReader := csv.NewReader(r)
csvReader.Comma = '\t'
dec, err := NewDecoder(csvReader)
	if err != nil {
	log.Fatal(err)
}
var users []User
for {
	var u User
	if err := dec.Decode(&u); err == io.EOF {
		break
	} else if err != nil {
		log.Fatal(err)
	}
	users = append(users, u)
}
Encoder编码器:
var buf bytes.Buffer
w := csv.NewWriter(&buf)
w.Comma = '\t'
enc := csvutil.NewEncoder(w)
for _, u := range users {
	if err := enc.Encode(u); err != nil {
		log.Fatal(err)
	}
}
w.Flush()
	if err := w.Error(); err != nil {
	log.Fatal(err)
}

7.Custom Types and Overrides	自定义类型和覆盖
有多种方法可以自定义或重写类型的行为。
	1.a type implements csvutil.Marshaler and/or csvutil.Unmarshaler
type Foo int64
func (f Foo) MarshalCSV() ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
}
func (f *Foo) UnmarshalCSV(data []byte) error {
	i, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(i)
	return nil
}
	2.a type implements encoding.TextUnmarshaler and/or encoding.TextMarshaler
type Foo int64
func (f Foo) MarshalText() ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
}
func (f *Foo) UnmarshalText(data []byte) error {
	i, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(i)
	return nil
}
	3.a type is registered using Encoder.Register and/or Decoder.Register
type Foo int64
enc.Register(func(f Foo) ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
})
dec.Register(func(data []byte, f *Foo) error {
	v, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(v)
	return nil
})
	4.a type implements an interface that was registered using Encoder.Register and/or Decoder.Register
type Foo int64
func (f Foo) String() string {
	return strconv.FormatInt(int64(f), 16)
}
func (f *Foo) Scan(state fmt.ScanState, verb rune) error {
	// too long; look here: https://github.com/jszwec/csvutil/blob/master/example_decoder_register_test.go#L19
}
enc.Register(func(s fmt.Stringer) ([]byte, error) {
	return []byte(s.String()), nil
})
dec.Register(func(data []byte, s fmt.Scanner) error {
	_, err := fmt.Sscan(string(data), s)
	return err
})
编码器和解码器的优先顺序为：1.type is registered类型已注册 2.type implements an interface that was registered类型实现已注册的接口
						3.csvutil.{Un,M}arshaler	4.encoding.Text{Un,M}arshaler

8.Custom time.Time format	自定义时间格式
由于解码器和编码器都内置编码支持encoding.TextUnmarshaler和encoding.TextMarshaler，因此可以在结构字段中按原样使用类型时间。这意味着默认情况
下，时间具有特定格式；查看MarshalText和UnmarshalText。有两种方法可以覆盖它，您选择哪一种取决于您的用例：
	1.Via Register func (based on encoding/json)
const format = "2006/01/02 15:04:05"
marshalTime := func(t time.Time) ([]byte, error) {
	return t.AppendFormat(nil, format), nil
}
unmarshalTime := func(data []byte, t *time.Time) error {
	tt, err := time.Parse(format, string(data))
	if err != nil {
		return err
	}
	*t = tt
	return nil
}
enc := csvutil.NewEncoder(w)
enc.Register(marshalTime)
dec, err := csvutil.NewDecoder(r)
if err != nil {
	return err
}
dec.Register(unmarshalTime)
	2.With custom type:
type Time struct {
	time.Time
}
const format = "2006/01/02 15:04:05"
func (t Time) MarshalCSV() ([]byte, error) {
	var b [len(format)]byte
	return t.AppendFormat(b[:0], format), nil
}
func (t *Time) UnmarshalCSV(data []byte) error {
	tt, err := time.Parse(format, string(data))
	if err != nil {
		return err
	}
	*t = Time{Time: tt}
	return nil
}

9.Custom struct tags	自定义结构标记
与其他Go编码包一样，结构字段标记可以用于设置自定义名称或选项。默认情况下，编码器和解码器正在查看csv标记。但是，可以通过手动设置“标记”字段来覆盖此设置。
type Foo struct {
	Bar int `custom:"bar"`
}
dec, err := csvutil.NewDecoder(r)
if err != nil {
	log.Fatal(err)
}
dec.Tag = "custom"
enc := csvutil.NewEncoder(w)
enc.Tag = "custom"

10.Slice and Map fields	切片和映射字段
切片和映射字段没有默认的编码/解码支持，因为这些值没有CSV规范。在这种情况下，建议创建自定义类型别名并实现Marshaler和Unmarshaler接口。请注意，切片
和映射别名的行为与其他类型的别名不同-不需要类型转换。
type Strings []string
func (s Strings) MarshalCSV() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil // strings.Join takes []string but it will also accept Strings
}
type StringMap map[string]string
func (sm StringMap) MarshalCSV() ([]byte, error) {
	return []byte(fmt.Sprint(sm)), nil
}
func main() {
	b, err := csvutil.Marshal([]struct {
		Strings Strings   `csv:"strings"`
		Map     StringMap `csv:"map"`
	}{
		{[]string{"a", "b"}, map[string]string{"a": "1"}}, // no type casting is required for slice and map aliases切片和映射别名不需要类型转换
		{Strings{"c", "d"}, StringMap{"b": "1"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", b)
	// Output:
	// strings,map
	// "a,b",map[a:1]
	// "c,d",map[b:1]
}

11.Nested/Embedded structs	嵌套/嵌入结构
编码器和解码器都支持嵌套或嵌入式结构。
package main
import (
	"fmt"
	"github.com/jszwec/csvutil"
)
type Address struct {
	Street string `csv:"street"`
	City   string `csv:"city"`
}
type User struct {
	Name string `csv:"name"`
	Address
}
func main() {
	users := []User{
		{
			Name: "John",
			Address: Address{
				Street: "Boylston",
				City:   "Boston",
			},
		},
	}
	b, err := csvutil.Marshal(users)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", b)
	var out []User
	if err := csvutil.Unmarshal(b, &out); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", out)
	// Output:
	//
	// name,street,city
	// John,Boylston,Boston
	//
	// [{Name:John Address:{Street:Boylston City:Boston}}]
}

12.Inline tag	内联标签
带有内联标记的字段的行为类似于嵌入的结构字段。但是，它提供了为所有基础字段指定前缀的可能性。当一个结构可以定义多个CSV列时，这可能很有用，因为它们之间
的差异仅限于某个前缀。看看下面的例子。
package main
import (
	"fmt"
	"github.com/jszwec/csvutil"
)
func main() {
	type Address struct {
		Street string `csv:"street"`
		City   string `csv:"city"`
	}
	type User struct {
		Name        string  `csv:"name"`
		Address     Address `csv:",inline"`
		HomeAddress Address `csv:"home_address_,inline"`
		WorkAddress Address `csv:"work_address_,inline"`
		Age         int     `csv:"age,omitempty"`
	}
	users := []User{
		{
			Name:        "John",
			Address:     Address{"Washington", "Boston"},
			HomeAddress: Address{"Boylston", "Boston"},
			WorkAddress: Address{"River St", "Cambridge"},
			Age:         26,
		},
	}
	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%s\n", b)
	// Output:
	// name,street,city,home_address_street,home_address_city,work_address_street,work_address_city,age
	// John,Washington,Boston,Boylston,Boston,River St,Cambridge,26
}