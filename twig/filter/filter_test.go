package filter

import (
	"strings"
	"testing"
	"time"

	"github.com/tyler-sommer/stick"
)

func TestFilters(t *testing.T) {
	newBatchFunc := func(in stick.Value, args ...stick.Value) func() stick.Value {
		return func() stick.Value {
			batched := filterBatch(nil, in, args...)
			res := ""
			stick.Iterate(batched, func(k, v stick.Value, l stick.Loop) (bool, error) {
				stick.Iterate(v, func(k, v stick.Value, l stick.Loop) (bool, error) {
					res += stick.CoerceString(v) + "."
					return false, nil
				})
				res += "."
				return false, nil
			})
			return res
		}
	}

	tz, err := time.LoadLocation("Australia/Perth")
	if nil != err {
		t.Error(err)
	}
	testDate := time.Date(1980, 5, 31, 22, 01, 0, 0, tz)
	testDate2 := time.Date(2018, 2, 3, 2, 1, 44, 123456000, tz)

	tests := []struct {
		name     string
		actual   func() stick.Value
		expected stick.Value
	}{
		{"default nil", func() stick.Value { return filterDefault(nil, nil, "person") }, "person"},
		{"default empty string", func() stick.Value { return filterDefault(nil, "", "person") }, "person"},
		{"default not empty", func() stick.Value { return filterDefault(nil, "user", "person") }, "user"},
		{"abs positive", func() stick.Value { return filterAbs(nil, 5.1) }, 5.1},
		{"abs negative", func() stick.Value { return filterAbs(nil, -42) }, 42.0 /* note: coerced to float */},
		{"abs invalid", func() stick.Value { return filterAbs(nil, "invalid") }, 0.0},
		{"len string", func() stick.Value { return filterLength(nil, "hello") }, 5},
		{"len nil", func() stick.Value { return filterLength(nil, nil) }, 0},
		{"len slice", func() stick.Value { return filterLength(nil, []string{"h", "e"}) }, 2},
		{"capitalize", func() stick.Value { return filterCapitalize(nil, "word") }, "Word"},
		{"lower", func() stick.Value { return filterLower(nil, "HELLO, WORLD!") }, "hello, world!"},
		{"title", func() stick.Value { return filterTitle(nil, "hello, world!") }, "Hello, World!"},
		{"trim", func() stick.Value { return filterTrim(nil, " Hello   ") }, "Hello"},
		{"upper", func() stick.Value { return filterUpper(nil, "hello, world!") }, "HELLO, WORLD!"},
		{"batch underfull with fill", newBatchFunc([]int{1, 2, 3, 4, 5, 6, 7, 8}, 3, "No Item"), "1.2.3..4.5.6..7.8.No Item.."},
		{"batch underfull without fill", newBatchFunc([]int{1, 2, 3, 4, 5}, 3), "1.2.3..4.5.."},
		{"batch full", newBatchFunc([]int{1, 2, 3, 4}, 2), "1.2..3.4.."},
		{"batch empty", newBatchFunc([]int{}, 10), ""},
		{"batch nil", newBatchFunc(nil, 10), ""},
		{"first array", func() stick.Value { return filterFirst(nil, []string{"1", "2", "3", "4"}) }, "1"},
		{"first string", func() stick.Value { return filterFirst(nil, "1234") }, "1"},
		{"first string utf8", func() stick.Value { return filterFirst(nil, "東京") }, "東"},
		{"date c", func() stick.Value { return filterDate(nil, testDate, "c") }, "1980-05-31T22:01:00+08:00"},
		{"date r", func() stick.Value { return filterDate(nil, testDate, "r") }, "Sat, 31 May 1980 22:01:00 +0800"},
		{"date test", func() stick.Value { return filterDate(nil, testDate2, "d D j l F m M n Y y a A g G h H i s O P T") }, "03 Sat 3 Saturday February 02 Feb 2 2018 18 am AM 2 02 02 02 01 44 +0800 +08:00 AWST"},
		{"date u", func() stick.Value { return filterDate(nil, testDate2, "s.u") }, "44.123456"},
		{"join", func() stick.Value { return filterJoin(nil, []string{"a", "b", "c"}, "-") }, "a-b-c"},
		{"merge", func() stick.Value {
			return stickSliceToString(filterMerge(nil, []string{"a", "b"}, []string{"c", "d"}))
		}, "a.b.c.d"},

		{
			"json encode",
			func() stick.Value {
				return filterJSONEncode(nil, map[string]interface{}{"a": 1, "b": true, "c": 3.14, "d": "a string", "e": []string{"one", "two"}, "f": map[string]interface{}{"alpha": "foo", "beta": nil}})
			},
			`{"a":1,"b":true,"c":3.14,"d":"a string","e":["one","two"],"f":{"alpha":"foo","beta":null}}`,
		},

		{"keys array", func() stick.Value { return stickSliceToString(filterKeys(nil, []string{"a", "b", "c"})) }, `0.1.2`},
		{"keys map", func() stick.Value {
			return stickSliceToString(filterKeys(nil, map[string]string{"a": "1", "b": "2", "c": "3"}))
		}, `a.b.c`},

		{"last array", func() stick.Value { return filterLast(nil, []string{"1", "2", "3", "4"}) }, "4"},
		{"last string", func() stick.Value { return filterLast(nil, "1234") }, "4"},
		{"last string utf8", func() stick.Value { return filterLast(nil, "東京") }, "京"},

		{"nl2br", func() stick.Value { return filterNL2BR(nil, "a\nb\nc").(stick.SafeValue).Value() }, "a<br />b<br />c"},

		{"raw", func() stick.Value { return filterRaw(nil, "<script></script>").(stick.SafeValue).Value() }, "<script></script>"},

		{
			"replace",
			func() stick.Value {
				return filterReplace(nil, "I like %this% and %that%.", map[string]string{"%this%": "foo", "%that%": "bar"})
			},
			"I like foo and bar.",
		},

		{"reverse array", func() stick.Value { return stickSliceToString(filterReverse(nil, []string{"1", "2", "3", "4"})) }, "4.3.2.1"},
		{"reverse string", func() stick.Value { return filterReverse(nil, "1234") }, "4321"},
		{"reverse string utf8", func() stick.Value { return filterReverse(nil, "東京") }, "京東"},

		{"round common down", func() stick.Value { return filterRound(nil, 3.4) }, 3.0},
		{"round common up", func() stick.Value { return filterRound(nil, 3.6) }, 4.0},
		{"round common half", func() stick.Value { return filterRound(nil, 3.5) }, 4.0},
		{"round common down 2 digits", func() stick.Value { return filterRound(nil, 3.114, 2) }, 3.11},
		{"round common up 2 digits", func() stick.Value { return filterRound(nil, 3.116, 2) }, 3.12},
		{"round common half 2 digits", func() stick.Value { return filterRound(nil, 3.115, 2) }, 3.12},
		{"round ceil", func() stick.Value { return filterRound(nil, 3.123, 0, "ceil") }, 4.0},
		{"round ceil 2 digits", func() stick.Value { return filterRound(nil, 3.123, 2, "ceil") }, 3.13},
		{"round floor", func() stick.Value { return filterRound(nil, 3.123, 0, "floor") }, 3.0},
		{"round floor 2 digits", func() stick.Value { return filterRound(nil, 3.123, 2, "floor") }, 3.12},

		{"slice array 2..", func() stick.Value { return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d", "e"}, 2)) }, "c.d.e"},
		{"slice array 2..2", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d", "e"}, 2, 2))
		}, "c.d"},
		{"slice array -3..2", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d", "e"}, -3, 2))
		}, "c.d"},
		{"slice array -3..-1", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d", "e"}, -3, -1))
		}, "c.d"},
		{"slice string 2..", func() stick.Value { return filterSlice(nil, "abcde", 2) }, "cde"},
		{"slice string 2..2", func() stick.Value { return filterSlice(nil, "abcde", 2, 2) }, "cd"},
		{"slice string -3..2", func() stick.Value { return filterSlice(nil, "abcde", -3, 2) }, "cd"},
		{"slice string -3..-1", func() stick.Value { return filterSlice(nil, "abcde", -3, -1) }, "cd"},

		{"split string on comma", func() stick.Value { return stickSliceToString(filterSplit(nil, "a,b,c,d,e", ",")) }, "a.b.c.d.e"},
		{"split string on comma limited", func() stick.Value { return stickSliceToString(filterSplit(nil, "a,b,c,d,e", ",", 3)) }, "a.b.c"},
		{"split string on comma neg limited", func() stick.Value { return stickSliceToString(filterSplit(nil, "a,b,c,d,e", ",", -3)) }, "a.b"},
		{"split string on comma neg limited big", func() stick.Value { return stickSliceToString(filterSplit(nil, "a,b,c,d,e", ",", -10)) }, ""},
		{"split string on empty", func() stick.Value { return stickSliceToString(filterSplit(nil, "abcde", "")) }, "a.b.c.d.e"},
		{"split string on empty", func() stick.Value { return stickSliceToString(filterSplit(nil, "abcde", "", 2)) }, "ab.cd.e"},
		{"split string on empty", func() stick.Value { return stickSliceToString(filterSplit(nil, "abcde", "", -2)) }, "a.b.c.d.e"},
	}
	for _, test := range tests {
		res := test.actual()
		if res != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, res)
		}
	}
}

func stickSliceToString(value stick.Value) (output string) {
	var slice []string
	stick.Iterate(value, func(k, v stick.Value, l stick.Loop) (bool, error) {
		slice = append(slice, stick.CoerceString(v))
		return false, nil
	})

	return strings.Join(slice, ".")
}
