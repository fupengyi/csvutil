package CSVutil

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

1.Examples Header
import (
	"fmt"
	"log"
	"github.com/jszwec/csvutil"
)
func main() {
	type User struct {
		ID    int
		Name  string
		Age   int `csv:",omitempty"`
		State int `csv:"-"`
		City  string
		ZIP   string `csv:"zip_code"`
	}
	header, err := csvutil.Header(User{}, "csv")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(header)		// Output: [ID Name Age City zip_code]
}

2.Examples Marshal
package main
import (
	"fmt"
	"time"
	"github.com/jszwec/csvutil"
)

func main() {
	type Address struct {
		City    string
		Country string
	}
	
	type User struct {
		Name string
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
	//Output:
	//Name,City,Country,age,CreatedAt
	//John,Boston,USA,26,2010-06-02T12:00:00Z
	//Alice,SF,USA,,0001-01-01T00:00:00Z
}

3.Example (CustomMarshalCSV) ¶
package main

import (
	"fmt"
	
	"github.com/jszwec/csvutil"
)

type Status uint8

const (
	Unknown = iota
	Success
	Failure
)

func (s Status) MarshalCSV() ([]byte, error) {
	switch s {
	case Success:
		return []byte("success"), nil
	case Failure:
		return []byte("failure"), nil
	default:
		return []byte("unknown"), nil
	}
}

type Job struct {
	ID     int
	Status Status
}

func main() {
	jobs := []Job{
		{1, Success},
		{2, Failure},
	}
	
	b, err := csvutil.Marshal(jobs)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
	//Output:
	//
	//ID,Status
	//1,success
	//2,failure
}

4.Example (SliceMap)
package main

import (
	"fmt"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

type Strings []string

func (s Strings) MarshalCSV() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil 	// strings.Join takes []string but it will also accept Strings
}													// strings.Join 接受 []string 但它也接受字符串


type StringMap map[string]string

func (sm StringMap) MarshalCSV() ([]byte, error) {
	return []byte(fmt.Sprint(sm)), nil
}

func main() {
	b, err := csvutil.Marshal([]struct {
		Strings Strings   `csv:"strings"`
		Map     StringMap `csv:"map"`
	}{
		{[]string{"a", "b"}, map[string]string{"a": "1"}}, 	// no type casting is required for slice and map aliases
		{Strings{"c", "d"}, StringMap{"b": "1"}},					// slice 和 map 别名不需要类型转换
		
	})
	
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("%s\n", b)
	//Output:
	//
	//strings,map
	//"a,b",map[a:1]
	//"c,d",map[b:1]
}

5.Example
package main

import (
	"fmt"
	"time"
	
	"github.com/jszwec/csvutil"
)

func main() {
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
	//Output:
	//
	//{Name:jacek Age:26 CreatedAt:2012-04-01 15:00:00 +0000 UTC}
	//{Name:john Age:0 CreatedAt:0001-01-01 00:00:00 +0000 UTC}
}

6.Example (CustomUnmarshalCSV) ¶
package main

import (
	"fmt"
	"strconv"
	
	"github.com/jszwec/csvutil"
)

type Bar int

func (b *Bar) UnmarshalCSV(data []byte) error {
	n, err := strconv.Atoi(string(data))
	*b = Bar(n)
	return err
}

type Foo struct {
	Int int `csv:"int"`
	Bar Bar `csv:"bar"`
}

func main() {
	var csvInput = []byte(`
		int,bar
		5,10
		6,11`
	)
	
	var foos []Foo
	if err := csvutil.Unmarshal(csvInput, &foos); err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Printf("%+v", foos)
	//Output:
	//
	//[{Int:5 Bar:10} {Int:6 Bar:11}]
}

7.Example (DecodeEmbedded) ¶
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type Address struct {
		ID    int    `csv:"id"` // same field as in User - this one will be empty // 与 User 中相同的字段 - 这个字段将为空
		City  string `csv:"city"`
		State string `csv:"state"`
	}
	
	type User struct {
		Address
		ID   int    `csv:"id"` // same field as in Address - this one wins	// 与 Address 中相同的字段 - 这个获胜
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}
	
	csvReader := csv.NewReader(strings.NewReader(
		"id,name,age,city,state\n" +
			"1,alice,25,la,ca\n" +
			"2,bob,30,ny,ny"))
	
	dec, err := csvutil.NewDecoder(csvReader)
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
	
	fmt.Println(users)
	//Output:
	//
	//[{{0 la ca} 1 alice 25} {{0 ny ny} 2 bob 30}]
}

8.Example (DecodingDataWithNoHeader) ¶
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	
	"github.com/jszwec/csvutil"
)

type User struct {
	ID    int
	Name  string
	Age   int `csv:",omitempty"`
	State int `csv:"-"`
	City  string
	ZIP   string `csv:"zip_code"`
}

var userHeader []string

func init() {
	h, err := csvutil.Header(User{}, "csv")
	if err != nil {
		log.Fatal(err)
	}
	userHeader = h
}

func main() {
	data := []byte(`
		1,John,27,la,90005
		2,Bob,,ny,10005`)
	
	r := csv.NewReader(bytes.NewReader(data))
	
	dec, err := csvutil.NewDecoder(r, userHeader...)
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
	//Output:
	//
	//[{ID:1 Name:John Age:27 State:0 City:la ZIP:90005} {ID:2 Name:Bob Age:0 State:0 City:ny ZIP:10005}]
}

9.Example (InterfaceValues) ¶
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	
	"github.com/jszwec/csvutil"
)

// Value defines one record in the csv input. In this example it is important	// 值定义 csv 输入中的一条记录。在这个例子中很重要
// that Type field is defined before Value. Decoder reads headers and values	// 该类型字段在值之前定义。解码器读取标头和值
// in the same order as struct fields are defined.								// 与定义结构字段的顺序相同。
type Value struct {
	Type  string `csv:"type"`
	Value any    `csv:"value"`
}

func main() {
	// lets say our csv input defines variables with their types and values.	// 假设我们的 csv 输入定义变量及其类型和值。
	data := []byte(`
		type,value
		string,string_value
		int,10
	`)
	
	dec, err := csvutil.NewDecoder(csv.NewReader(bytes.NewReader(data)))
	if err != nil {
		log.Fatal(err)
	}
	
	// we would like to read every variable and store their already parsed values	// 我们想读取每个变量并存储它们已经解析的值
	// in the interface field. We can use Decoder.Map function to initialize		// 在接口字段中。我们可以使用 Decoder.Map 函数来初始化
	// interface with proper values depending on the input.							// 根据输入使用适当的值进行接口。
	var value Value
	dec.Map = func(field, column string, v any) string {
		if column == "type" {
			switch field {
			case "int": // csv input tells us that this variable contains an int.	// csv 输入告诉我们这个变量包含一个 int。
				var n int
				value.Value = &n // lets initialize interface with an initialized int pointer.	// 让我们用一个初始化的 int 指针来初始化接口。
			default:
				return field
			}
		}
		return field
	}
	
	for {
		value = Value{}
		if err := dec.Decode(&value); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		
		if value.Type == "int" {
			// our variable type is int, Map func already initialized our interface		// 我们的变量类型是 int，Map func 已经初始化了我们的接口
			// as int pointer, so we can safely cast it and use it.						// 作为 int 指针，所以我们可以安全地转换和使用它。
			n, ok := value.Value.(*int)
			if !ok {
				log.Fatal("expected value to be *int")
			}
			fmt.Printf("value_type: %s; value: (%T) %d\n", value.Type, value.Value, *n)				// value_type: int; value: (*int) 10
		} else {
			fmt.Printf("value_type: %s; value: (%T) %v\n", value.Type, value.Value, value.Value)		// value_type: string; value: (string) string_value
		}
	}
}

10.Example
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type User struct {
		ID   *int   `csv:"id,omitempty"`
		Name string `csv:"name"`
		City string `csv:"city"`
		Age  int    `csv:"age"`
	}
	
	csvReader := csv.NewReader(strings.NewReader(`
		id,name,age,city
		,alice,25,la
		,bob,30,ny`)
	)
	
	dec, err := csvutil.NewDecoder(csvReader)
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
	
	fmt.Println(users)
	//Output:
	//
	//[{<nil> alice la 25} {<nil> bob ny 30}]
}

11.Example (Array)
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type User struct {
		ID   *int   `csv:"id,omitempty"`
		Name string `csv:"name"`
		City string `csv:"city"`
		Age  int    `csv:"age"`
	}
	
	csvReader := csv.NewReader(strings.NewReader(`
		id,name,age,city
		,alice,25,la
		,bob,30,ny
		,john,29,ny`)
	)
	
	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}
	
	var users [2]User
	if err := dec.Decode(&users); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(users)
	//Output:
	//
	//[{<nil> alice la 25} {<nil> bob ny 30}]
}

12.Example (Inline)
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
	
	data := []byte(
		"name,street,city,home_address_street,home_address_city,work_address_street,work_address_city,age\n" +
			"John,Washington,Boston,Boylston,Boston,River St,Cambridge,26",
	)
	
	var users []User
	if err := csvutil.Unmarshal(data, &users); err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Println(users)
	//Output:
	//
	//[{John {Washington Boston} {Boylston Boston} {River St Cambridge} 26}]
}

13.Example (Slice)
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type User struct {
		ID   *int   `csv:"id,omitempty"`
		Name string `csv:"name"`
		City string `csv:"city"`
		Age  int    `csv:"age"`
	}
	
	csvReader := csv.NewReader(strings.NewReader(`
		id,name,age,city
		,alice,25,la
		,bob,30,ny`)
	)
	
	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}
	
	var users []User
	if err := dec.Decode(&users); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(users)
	//Output:
	//
	//[{<nil> alice la 25} {<nil> bob ny 30}]
}

14.Example
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type User struct {
		Name      string            `csv:"name"`
		City      string            `csv:"city"`
		Age       int               `csv:"age"`
		OtherData map[string]string `csv:"-"`
	}
	
	csvReader := csv.NewReader(strings.NewReader(`
		name,age,city,zip
		alice,25,la,90005
		bob,30,ny,10005`))
	
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
	//Output:
	//
	//[{alice la 25 map[zip:90005]} {bob ny 30 map[zip:10005]}]
}

15.Example (All)
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type Address struct {
		City    string
		Country string
	}
	
	type User struct {
		Name string
		Address
		Age int `csv:"age,omitempty"`
	}
	
	users := []User{
		{Name: "John", Address: Address{"Boston", "USA"}, Age: 26},
		{Name: "Bob", Address: Address{"LA", "USA"}, Age: 27},
		{Name: "Alice", Address: Address{"SF", "USA"}},
	}
	
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := csvutil.NewEncoder(w).Encode(users); err != nil {
		fmt.Println("error:", err)
	}
	
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Println(buf.String())
	//Output:
	//
	//Name,City,Country,age
	//John,Boston,USA,26
	//Bob,LA,USA,27
	//Alice,SF,USA,
}

16.Example (Inline)
package main

import (
	"fmt"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type Owner struct {
		Name string `csv:"name"`
	}
	
	type Address struct {
		Street string `csv:"street"`
		City   string `csv:"city"`
		Owner  Owner  `csv:"owner_,inline"`
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
			Address:     Address{"Washington", "Boston", Owner{"Steve"}},
			HomeAddress: Address{"Boylston", "Boston", Owner{"Steve"}},
			WorkAddress: Address{"River St", "Cambridge", Owner{"Steve"}},
			Age:         26,
		},
	}
	
	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Printf("%s\n", b)
	//Output:
	//
	//name,street,city,owner_name,home_address_street,home_address_city,home_address_owner_name,work_address_street,work_address_city,work_address_owner_name,age
	//John,Washington,Boston,Steve,Boylston,Boston,Steve,River St,Cambridge,Steve,26
}

17.Example (Streaming)
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type Address struct {
		City    string
		Country string
	}
	
	type User struct {
		Name string
		Address
		Age int `csv:"age,omitempty"`
	}
	
	users := []User{
		{Name: "John", Address: Address{"Boston", "USA"}, Age: 26},
		{Name: "Bob", Address: Address{"LA", "USA"}, Age: 27},
		{Name: "Alice", Address: Address{"SF", "USA"}},
	}
	
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)
	
	for _, u := range users {
		if err := enc.Encode(u); err != nil {
			fmt.Println("error:", err)
		}
	}
	
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Println(buf.String())
	//Output:
	//
	//Name,City,Country,age
	//John,Boston,USA,26
	//Bob,LA,USA,27
	//Alice,SF,USA,
}

18.Example
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	
	"github.com/jszwec/csvutil"
)

func main() {
	type User struct {
		Name string
		Age  int `csv:"age,omitempty"`
	}
	
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)
	
	if err := enc.EncodeHeader(User{}); err != nil {
		fmt.Println("error:", err)
	}
	
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}
	
	fmt.Println(buf.String())
	Output:
	
	Name,age
}

19.





