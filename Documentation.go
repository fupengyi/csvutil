package CSVutil

import (
	"errors"
	"reflect"
)

Documentation
包 csvutil 提供 CSV 和 Go 值之间的快速和惯用映射。
这个包本身不提供 CSV 解析器，它基于由例如实现的 Reader 和 Writer 接口的标准 csv 包。这提供了选择任何其他可能性能更高的 CSV 写入器或读取器的可能性。

Variables ¶
var ErrFieldCount = errors.New("wrong number of fields in record")		// 当 header 的长度与读取记录的长度不匹配时，返回 ErrFieldCount。

Functions ¶
func Header(v any, tag string) ([]string, error)	// Header 扫描提供的结构类型并为其生成 CSV 标头。
// 字段名称的书写顺序与结构字段的定义顺序相同。嵌入式结构的字段被视为外部结构的一部分。嵌入类型和标记的字段与任何其他字段一样对待。
// 未导出的字段和带有标记“-”的字段将被忽略。
// 标记字段优先于同名的非标记字段。
// 遵循 Go 可见性规则，如果在同一级别上有多个具有相同名称（标记或未标记）的字段并且它们之间的选择不明确，那么所有这些字段都将被忽略。
// 为每种类型调用一次 Header 是一种很好的做法。适合调用它的地方是init函数。查看 Decoder.DecodingDataWithNoHeader 示例。
// 如果标签留空，将使用默认的“csv”。
// 如果提供的值为 nil 或不是结构，标头将返回 UnsupportedTypeError。
// Example

func Marshal(v any) ([]byte, error)		// Marshal 返回切片或数组 v 的 CSV 编码。如果 v 不是切片或元素不是结构，则 Marshal 返回 InvalidMarshalError。
// Marshal 使用 std encoding/csv.Writer 及其 csv 编码的默认设置。
// 即使对于空切片，Marshal 也会始终对 CSV 标头进行编码。
// 有关确切的编码规则，请查看 Encoder.Encode 方法。
// Example	Example (CustomMarshalCSV)	Example (SliceMap)

func Unmarshal(data []byte, v any) error	// Unmarshal 解析 CSV 编码数据并将结果存储在 v 指向的切片或数组中。如果 v 为 nil 或者不是指向结构切片或结构数组的指针，则 Unmarshal 返回 InvalidUnmarshalError。
// Unmarshal 使用 std encoding/csv.Reader 进行解析，使用 csvutil.Decoder 填充提供的切片中的结构元素。有关确切的解码规则，请查看解码器的文档。
// 数据中的第一行被视为标题。解码器将使用它将 csv 列映射到结构的字段。
// 如果成功，提供的切片将被重新初始化，其内容将完全替换为解码数据。
// Example

Types ¶
type DecodeError struct {															// DecodeError 提供解码错误的上下文（如果可用）。
	// Field describes the struct's tag or field name on which the error happened.	// Field 描述发生错误的结构的标记或字段名称。
	Field string
	
	// Line is 1-indexed line number taken from FieldPost method. It is only		// Line 是从 FieldPost 方法中获取的 1 索引行号。它只是
	// available if the used Reader supports FieldPos method.						// 如果使用的 Reader 支持 FieldPos 方法，则可用。
	Line int
	
	// Column is 1-indexed column index taken from FieldPost method. It is only		// Column 是从 FieldPost 方法中获取的 1-indexed 列索引。它只是
	// available if the used Reader supports FieldPos method.						// 如果使用的 Reader 支持 FieldPos 方法，则可用。
	Column int
	
	// Error is the actual error that was returned while attempting to decode		// 错误是尝试解码时返回的实际错误
	// a field.																		// 一个字段。
	Err error
}
// 调用者应该使用 errors.As 以便在需要时获取底层错误。
// 一些 DecodeError 的字段只有在 Reader 支持 PosField 方法时才会被填充。特别是行和列。从 Go1.17 开始，FieldPos 在 csv.Reader 中可用。
1.func (e *DecodeError) Error() string
2.func (e *DecodeError) Unwrap() error

type Decoder struct {																// 解码器读取字符串记录并将其解码为结构。
	// Tag defines which key in the struct field's tag to scan for names and		// 标签定义结构字段标签中的哪个键来扫描名称和
	// options (Default: 'csv').													// 选项（默认：'csv'）。
	Tag string
	
	// If true, Decoder will return a MissingColumnsError if it discovers			// 如果为真，Decoder 将在检测到错误时返回 MissingColumnsError
	// that any of the columns are missing. This means that a CSV input				// 缺少任何列。这意味着 CSV 输入
	// will be required to contain all columns that were defined in the				// 将需要包含在
	// provided struct.																// 提供的结构。
	DisallowMissingColumns bool
	
	// If not nil, Map is a function that is called for each field in the csv		// 如果不为 nil，则 Map 是为 csv 中的每个字段调用的函数
	// record before decoding the data. It allows mapping certain string values		// 在解码数据之前记录。它允许映射某些字符串值
	// for specific columns or types to a known format. Decoder calls Map with		// 将特定列或类型转换为已知格式。解码器调用 Map
	// the current column name (taken from header) and a zero non-pointer value		// 当前列名（取自标题）和一个零非指针值
	// of a type to which it is going to decode data into. Implementations			// 要将数据解码成的类型。实现
	// should use type assertions to recognize the type.							// 应该使用类型断言来识别类型。
	//
	// The good example of use case for Map is if NaN values are represented by		// Map 用例的一个很好的例子是，如果 NaN 值表示为
	// eg 'n/a' string, implementing a specific Map function for all floats			// 例如 'n/a' 字符串，为所有浮点数实现特定的 Map 函数
	// could map 'n/a' back into 'NaN' to allow successful decoding.				// 可以将 'n/a' 映射回 'NaN' 以允许成功解码。
	//
	// Use Map with caution. If the requirements of column or type are not met		// 谨慎使用 Map。如果不满足列或类型的要求
	// Map should return 'field', since it is the original value that was			// Map 应该返回 'field'，因为它是原来的值
	// read from the csv input, this would indicate no change.						// 从 csv 输入中读取，这表明没有变化。
	//
	// If struct field is an interface v will be of type string, unless the			// 如果 struct field 是接口 v 将是字符串类型，除非
	// struct field contains a settable pointer value - then v will be a zero		// struct 字段包含一个可设置的指针值 - 那么 v 将为零
	// value of that type.															// 该类型的值。
	//
	// Map must be set before the first call to Decode and not changed after it.	// Map 必须在第一次调用 Decode 之前设置，并且在它之后不能更改。
	Map func(field, col string, v any) string
	// contains filtered or unexported fields
}
// Example (CustomUnmarshalCSV) 	Example (DecodeEmbedded)	Example (DecodingDataWithNoHeader)	Example (InterfaceValues)
1.func NewDecoder(r Reader, header ...string) (dec *Decoder, err error)			// NewDecoder 返回一个从 r 读取的新解码器。
																				// 解码器将根据给定的 header 匹配结构字段。
																				// 如果 header 为空，NewDecoder 将读取一行并将其视为 header。
																				// 来自 r 的记录必须与 header 的长度相同。
																				// 如果 r 中没有数据并且调用者没有提供 header，NewDecoder 可能会返回 io.EOF。
2.func (d *Decoder) Decode(v any) (err error)		// Decode 从其输入读取下一个字符串记录或记录，并将其存储在 v 指向的值中，该值必须是指向结构、结构切片或结构数组的指针。
													// Decode 根据 header 匹配所有导出的结构字段。可以使用标签调整结构字段。
													// “omitempty”选项指定如果记录的字段为空字符串，则应从解码中省略该字段。
													// struct 字段标签的示例及其含义：
													// Decode matches this field with "myName" header column.	// 解码将此字段与“myName”标题列匹配。
													Field int `csv:"myName"`

													// Decode matches this field with "Field" header column.	// 解码将此字段与“Field”标题列匹配。
													Field int
													
													// Decode matches this field with "myName" header column and decoding is not	// 解码将此字段与“myName”标题列匹配并且解码不是
													// called if record's field is an empty string.									// 如果记录的字段为空字符串，则调用。
													Field int `csv:"myName,omitempty"`
													
													// Decode matches this field with "Field" header column and decoding is not		// 解码将此字段与“Field”标题列匹配并且解码不是
													// called if record's field is an empty string.									// 如果记录的字段为空字符串，则调用。
													Field int `csv:",omitempty"`
													
													// Decode ignores this field.													// 解码忽略这个字段。
													Field int `csv:"-"`
													
													// Decode treats this field exactly as if it was an embedded field and			// Decode 将此字段完全视为嵌入字段，并且
													// matches header columns that start with "my_prefix_" to all fields of this	// 将以“my_prefix_”开头的 header 列匹配到此的所有字段
													// type.																		// 类型。
													Field Struct `csv:"my_prefix_,inline"`
													
													// Decode treats this field exactly as if it was an embedded field.				// Decode 将此字段完全视为嵌入字段。
													Field Struct `csv:",inline"`
													// 默认情况下，解码会查找“csv”标签，但这可以通过设置 Decoder.Tag 字段来更改。
													// 要解码为自定义类型，v 必须实现 csvutil.Unmarshaler 或 encoding.TextUnmarshaler。
													// 带标签的匿名结构字段被视为普通字段，除非指定内联标签，否则它们必须实现 csvutil.Unmarshaler 或 encoding.TextUnmarshaler。
													// 没有标签的匿名结构字段被填充，就好像它们是主结构的一部分一样。但是，主结构中的字段具有更高的优先级，它们首先被填充。如果主结构和匿名结构字段具有相同的字段，则将填充主结构的字段。
													// []byte 类型的字段期望数据是 base64 编码的字符串。
													// 如果字符串值为“NaN”，浮点字段将解码为 NaN。此检查不区分大小写。
													// 接口字段被解码为字符串，除非它们包含可设置的指针值。
													// 如果字符串值为空，指针字段将解码为 nil。
													// 如果 v 是一个切片，Decode 会重置它并读取输入直到 EOF，将所有解码值存储在给定的切片中。解码在 EOF 时返回 nil。
													// 如果 v 是一个数组，Decode 读取输入直到 EOF 或直到它解码所有对应的数组元素。如果输入包含的元素少于数组，则额外的 Go 数组元素将设置为零值。解码在 EOF 时返回 nil，除非没有解码的记录。
													// 具有非空前缀的内联标签的字段不能是循环结构。将此类值传递给 Decode 将导致无限循环。
// Example		Example (Array)		Example (Inline) 	Example (Slice)
3.func (d *Decoder) Header() []string								//	Header 返回来自阅读器的第一行，或返回调用者定义的标题。
4.func (d *Decoder) NormalizeHeader(f func(string) string) error	// NormalizeHeader 将 f 应用于标题中的每一列。如果调用 f 导致标题列冲突，它会返回错误。
																	// NormalizeHeader 必须在 Decode 之前调用。
5.func (d *Decoder) Record() []string								// Record 返回最近读取的记录。切片在下一次调用 Decode 之前一直有效。
6.func (d *Decoder) Unused() []int									// Unused 返回由于缺少匹配的结构字段而在解码期间未使用的列索引列表。Example
7.func (d *Decoder) WithUnmarshalers(u *Unmarshalers)				// WithUnmarshalers 为解码器设置提供的 Unmarshalers。
																	// WithUnmarshalers 基于 encoding/json 提案：https://github.com/golang/go/issues/5901。

type Encoder struct {																// 编码器将结构 CSV 表示写入输出流。
	// Tag defines which key in the struct field's tag to scan for names and		// 标签定义结构字段标签中的哪个键来扫描名称和
	// options (Default: 'csv').													// 选项（默认：'csv'）。
	Tag string
	
	// If AutoHeader is true, a struct header is encoded during the first call		// 如果 AutoHeader 为真，则在第一次调用期间对结构头进行编码
	// to Encode automatically (Default: true).										// 自动编码（默认值：true）。
	AutoHeader bool
	// contains filtered or unexported fields
}
1.func NewEncoder(w Writer) *Encoder				// NewEncoder 返回一个写入 w 的新编码器。
2.func (e *Encoder) Encode(v any) error				// Encode 将 v 的 CSV 编码写入输出流。提供的参数 v 必须是结构、结构切片或结构数组。
													// 只有导出的字段将被编码。
													// 除非首先调用 EncodeHeader 或 AutoHeader 为 false，否则首次调用 Encode 将写入标头。可以使用标签自定义标题名称（默认为“csv”），否则使用原始字段名称。
													// 如果标头是通过 SetHeader 提供的，那么它会覆盖提供的数据类型的默认标头。字段按照提供的标头的顺序进行编码。如果标题中指定的列在提供的类型中不存在，它将被编码为空列。不属于提供的标头的字段将被忽略。如果提供的标头包含重复的列名，编码器无法保证正确的顺序。
													// 标头和字段的编写顺序与结构字段的定义顺序相同。嵌入式结构的字段被视为外部结构的一部分。嵌入类型和标记的字段与任何其他字段一样对待，但它们必须实现 Marshaler 或 encoding.TextMarshaler 接口。
													// Marshaler 接口优先于 encoding.TextMarshaler。
													// 标记字段优先于同名的非标记字段。
													// 遵循 Go 可见性规则，如果在同一级别上有多个具有相同名称（标记或未标记）的字段并且它们之间的选择不明确，那么所有这些字段都将被忽略。
													// Nil 值将被编码为空字符串。如果设置了 'omitempty' 标签，也会发生同样的情况，并且该值为默认值，如 0、false 或 nil 接口。
													// Bool 类型被编码为“true”或“false”。
													// 浮点类型使用 strconv.FormatFloat 以精度 -1 和“G”格式进行编码。 NaN 值被编码为“NaN”字符串。
													// []byte 类型的字段被编码为 base64 编码的字符串。
													// 可以使用“-”标签选项将字段从编码中排除。
													// 结构标签的例子：
													// Field appears as 'myName' header in CSV encoding.	// 字段在 CSV 编码中显示为“myName”标头。
													Field int `csv:"myName"`
													
													// Field appears as 'Field' header in CSV encoding.		// 字段在 CSV 编码中显示为“字段”标题。
													Field int
													
													// Field appears as 'myName' header in CSV encoding and is an empty string	// 字段在 CSV 编码中显示为 'myName' 标题，并且是一个空字符串
													// if Field is 0.															// 如果 Field 为 0。
													Field int `csv:"myName,omitempty"`
													
													// Field appears as 'Field' header in CSV encoding and is an empty string	// Field 在 CSV 编码中显示为 'Field' 头并且是一个空字符串
													// if Field is 0.															// 如果 Field 为 0。
													Field int `csv:",omitempty"`
													
													// Encode ignores this field.												// 编码忽略这个字段。
													Field int `csv:"-"`
													
													// Encode treats this field exactly as if it was an embedded field and adds	// Encode 将此字段完全视为嵌入字段并添加
													// "my_prefix_" to each field's name.										// “my_prefix_”到每个字段的名称。
													Field Struct `csv:"my_prefix_,inline"`
													
													// Encode treats this field exactly as if it was an embedded field.			// Encode 将此字段完全视为嵌入字段。
													Field Struct `csv:",inline"`
													// 具有非空前缀的内联标签的字段不能是循环结构。将此类值传递给 Encode 将导致无限循环。
													// 编码不刷新数据。如果使用的 Writer 支持，则调用者负责调用 Flush()。
// Example (All) 		Example (Inline) 		Example (Streaming)
3.func (e *Encoder) EncodeHeader(v any) error		// EncodeHeader 将提供的结构值的 CSV 标头写入输出流。提供的参数 v 必须是结构值。
													// 如果在它之前调用了 EncodeHeader，则第一个 Encode 方法调用将不会写入标头。在数据集可能为空但需要标头的情况下可以调用此方法。
													// EncodeHeader 类似于 Header 函数，但它与 Encoder 一起工作并直接写入输出流。查看标头文档以了解确切的标头编码规则。
// Example
4.func (enc *Encoder) SetHeader(header []string)	// SetHeader 覆盖提供的数据类型的默认标头。字段按照提供的标头的顺序进行编码。如果标题中指定的列在提供的类型中不存在，它将被编码为空列。不属于提供的标头的字段将被忽略。如果提供的标头包含重复的列名，编码器无法保证正确的顺序。
													// 必须在 EncodeHeader 和/或 Encode 之前调用 SetHeader 才能生效。
5.func (enc *Encoder) WithMarshalers(m *Marshalers)	// WithMarshalers 为编码器设置提供的编组器。
													// WithMarshalers 基于 encoding/json 提案：https://github.com/golang/go/issues/5901。

type InvalidDecodeError struct {					// InvalidDecodeError 描述传递给 Decode 的无效参数。 （Decode 的参数必须是非 nil 结构指针）
	Type reflect.Type								// func (e *InvalidDecodeError) Error() string
}

type InvalidEncodeError struct {					// 当提供的值无效时，Encode 会返回 InvalidEncodeError。
	Type reflect.Type								// func (e *InvalidEncodeError) Error() string
}

type InvalidMarshalError struct {					// 当提供的值无效时，Marshal 会返回 InvalidMarshalError。
	Type reflect.Type								// func (e *InvalidMarshalError) Error() string
}

type InvalidUnmarshalError struct {					// InvalidUnmarshalError 描述传递给 Unmarshal 的无效参数。 （Unmarshal 的参数必须是结构指针的非零切片）
	Type reflect.Type								// func (e *InvalidUnmarshalError) Error() string
}

type Marshaler interface {							// Marshaler 是由可以将自身编组为有效字符串的类型实现的接口。
	MarshalCSV() ([]byte, error)
}

type MarshalerError struct {						// 当 MarshalCSV 或 MarshalText 返回错误时，编码器返回 MarshalerError。
	Type          reflect.Type						// func (e *MarshalerError) Error() string
	MarshalerType string							// func (e *MarshalerError) Unwrap() error
	Err           error								// Unwrap 为 Go1.13+ 中的错误包实现了 Unwrap 接口。
}

type Marshalers struct {							// 编组器存储自定义解组函数。编组器是不可变的。
	// contains filtered or unexported fields		// 封送拆收器基于编码/json 提案：https://github.com/golang/go/issues/5901。
}
1.func MarshalFunc[T any](f func(T) ([]byte, error)) *Marshalers	// MarshalFunc 将提供的函数存储在 Marshalers 中并返回它。
																	// T 必须是具体类型，例如 Foo 或 *Foo，或至少具有一种方法的接口。
																	// 在编码过程中，字段首先与具体类型匹配。如果未找到匹配项，则编码器会查找字段是否按注册顺序实现了任何已注册的接口。
																	// 如果 T 是空接口，则 UnmarshalFunc 会崩溃。
2.func NewMarshalers(ms ...*Marshalers) *Marshalers					// NewMarshalers 将提供的 Marshalers 合并为一个并返回它。如果封送拆收器包含重复的函数签名，则第一个提供的将获胜。

type MissingColumnsError struct {					// 仅当 DisallowMissingColumns 选项设置为 true 时，解码器才会返回 MissingColumnsError。它包含所有缺失列的列表。
	Columns []string								// func (e *MissingColumnsError) Error() string
}

type Reader interface {								// Reader 提供读取单个 CSV 记录的接口。
	Read() ([]string, error)						// 如果没有剩余数据可供读取，则 Read 返回 (nil, io.EOF)。
}													// 它由 csv.Reader 实现。

type UnmarshalTypeError struct {					// UnmarshalTypeError 描述了一个不适合特定 Go 类型值的字符串值。
	Value string       // string value
	Type  reflect.Type // type of Go value it could not be assigned to		// 无法分配给它的 Go 值的类型
}
func (e *UnmarshalTypeError) Error() string

type Unmarshaler interface {						// Unmarshaler 是由类型实现的接口，可以解组单个记录的字段描述。
	UnmarshalCSV([]byte) error
}

type Unmarshalers struct {							// Unmarshalers 存储自定义解组函数。解组器是不可变的。
	// contains filtered or unexported fields		// 解组器基于编码/json 提案：https://github.com/golang/go/issues/5901。
}
1.func NewUnmarshalers(us ...*Unmarshalers) *Unmarshalers			// NewUnmarshalers 将提供的 Unmarshalers 合并为一个并返回它。如果 Unmarshalers 包含重复的函数签名，则先提供的函数签名获胜。
2.func UnmarshalFunc[T any](f func([]byte, T) error) *Unmarshalers	// UnmarshalFunc 将提供的函数存储在 Unmarshaler 中并返回它。
																	// 类型参数 T 必须是具体类型，例如 *time.Time，或至少具有一个方法的接口。
																	// 在解码期间，字段首先与具体类型匹配。如果未找到匹配项，则解码器会查找字段是否按注册顺序实现了任何已注册的接口。
																	// 如果 T 是空接口，则 UnmarshalFunc 会崩溃。

type UnsupportedTypeError struct {					// 尝试对不受支持的类型的值进行编码或解码时，会返回 UnsupportedTypeError。
	Type reflect.Type								// func (e *UnsupportedTypeError) Error() string
}

type Writer interface {								// Writer 提供写入单个 CSV 记录的接口。
	Write([]string) error							// 它由 csv.Writer 实现。
}
