package testar

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type recordTB struct {
	record   string
	recorded bool
}

func (tb *recordTB) Helper() {}
func (tb *recordTB) Fatalf(format string, args ...any) {
	if tb.recorded {
		panic(fmt.Sprintf("tb.Fatalf(%q, %#v) called after a previous call", format, args))
	}

	tb.recorded = true
	tb.record = fmt.Sprintf(format, args...)
}

func parseJSON(hook func(v map[string]any)) func(r io.Reader) (map[string]any, error) {
	return func(r io.Reader) (map[string]any, error) {
		var data map[string]any
		err := json.NewDecoder(r).Decode(&data)
		if err != nil {
			return nil, err
		}

		if hook != nil {
			hook(data)
		}

		return data, nil
	}
}

func TestTest(t *testing.T) {
	tests := []struct {
		Name string
		Hook func(v map[string]any)
		Want string
	}{
		{
			Name: "nonexistant",
			Want: "failed to parse test vector: open testdata/nonexistant.txtar: no such file or directory",
		},
		{
			Name: "missing-sections",
			Want: "malformed test vector: got 2 sections, want 4",
		},
		{
			Name: "wrong-sections",
			Want: "malformed test vector: got section 2 with name \"JSON\", want \"XML\"",
		},
		{
			Name: "wrong-format",
			Want: "invalid test vector: got format \"encoding/xml\", want \"encoding/json\"",
		},
		{
			Name: "no-error",
			Want: "unexpected success while parsing error test vector, want an error we won't get",
		},
		{
			Name: "wrong-error",
			Want: "wrong error while parsing error test vector:\ngot:  invalid character '<' looking for beginning of value\nwant: some other error",
		},
		{
			Name: "correct-error",
			Want: "",
		},
		{
			Name: "parse-error",
			Want: "failed to parse test vector: invalid character '<' looking for beginning of value",
		},
		{
			Name: "encode-error",
			Hook: func(v map[string]any) {
				v["invalid"] = make(chan bool)
			},
			Want: "failed to encode test vector: json: unsupported type: chan bool",
		},
		{
			Name: "different",
			Want: `test vector gave different output:
 {
-	"foo": "bar"
+	"other": "different"
 }
`,
		},
		{
			Name: "mirror",
			Want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var tb recordTB
			Test(&tb, test.Name, parseJSON(test.Hook), "encoding/json")
			if test.Want != "" {
				if tb.record != test.Want {
					t.Fatalf("wrong error:\ngot:  %s\nwant: %s", tb.record, test.Want)
				}

				// All good.
				return
			}

			if tb.recorded {
				t.Fatalf("unexpected error: %s", tb.record)
			}
		})
	}
}
