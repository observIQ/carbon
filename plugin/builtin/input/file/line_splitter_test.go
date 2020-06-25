package file

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenizerTestCase struct {
	Name              string
	Pattern           string
	Raw               []byte
	ExpectedTokenized []string
	ExpectedError     error
}

func (tc tokenizerTestCase) RunFunc(splitFunc bufio.SplitFunc) func(t *testing.T) {
	return func(t *testing.T) {
		scanner := bufio.NewScanner(bytes.NewReader(tc.Raw))
		scanner.Split(splitFunc)
		tokenized := make([]string, 0)
		for {
			ok := scanner.Scan()
			if !ok {
				assert.Equal(t, tc.ExpectedError, scanner.Err())
				break
			}
			tokenized = append(tokenized, scanner.Text())
		}

		assert.Equal(t, tc.ExpectedTokenized, tokenized)
	}
}

func TestLineStartSplitFunc(t *testing.T) {
	testCases := []tokenizerTestCase{
		{
			Name:    "OneLogSimple",
			Pattern: `LOGSTART \d+ `,
			Raw:     []byte(`LOGSTART 123 log1`),
			ExpectedTokenized: []string{
				`LOGSTART 123 log1`,
			},
		},
		{
			Name:    "TwoLogsSimple",
			Pattern: `LOGSTART \d+ `,
			Raw:     []byte(`LOGSTART 123 log1 LOGSTART 234 log2`),
			ExpectedTokenized: []string{
				`LOGSTART 123 log1 `,
				`LOGSTART 234 log2`,
			},
		},
		{
			Name:              "NoMatches",
			Pattern:           `LOGSTART \d+ `,
			Raw:               []byte(`file that has no matches in it`),
			ExpectedTokenized: []string{},
		},
		{
			Name:    "PrecedingNonMatches",
			Pattern: `LOGSTART \d+ `,
			Raw:     []byte(`part that doesn't match LOGSTART 123 part that matches`),
			ExpectedTokenized: []string{
				`part that doesn't match `,
				`LOGSTART 123 part that matches`,
			},
		},
		{
			Name:    "HugeLog100",
			Pattern: `LOGSTART \d+ `,
			Raw: func() []byte {
				newRaw := []byte(`LOGSTART 123 `)
				newRaw = append(newRaw, generatedByteSliceOfLength(100)...)
				newRaw = append(newRaw, []byte(`LOGSTART 234 endlog`)...)
				return newRaw
			}(),
			ExpectedTokenized: []string{
				`LOGSTART 123 ` + string(generatedByteSliceOfLength(100)),
				`LOGSTART 234 endlog`,
			},
		},
		{
			Name:    "HugeLog10000",
			Pattern: `LOGSTART \d+ `,
			Raw: func() []byte {
				newRaw := []byte(`LOGSTART 123 `)
				newRaw = append(newRaw, generatedByteSliceOfLength(10000)...)
				newRaw = append(newRaw, []byte(`LOGSTART 234 endlog`)...)
				return newRaw
			}(),
			ExpectedTokenized: []string{
				`LOGSTART 123 ` + string(generatedByteSliceOfLength(10000)),
				`LOGSTART 234 endlog`,
			},
		},
		{
			Name:    "ErrTooLong",
			Pattern: `LOGSTART \d+ `,
			Raw: func() []byte {
				newRaw := []byte(`LOGSTART 123 `)
				newRaw = append(newRaw, generatedByteSliceOfLength(1000000)...)
				newRaw = append(newRaw, []byte(`LOGSTART 234 endlog`)...)
				return newRaw
			}(),
			ExpectedError:     errors.New("bufio.Scanner: token too long"),
			ExpectedTokenized: []string{},
		},
	}

	for _, tc := range testCases {
		re := regexp.MustCompile(tc.Pattern)
		splitFunc := NewLineStartSplitFunc(re)
		t.Run(tc.Name, tc.RunFunc(splitFunc))
	}

	t.Run("FirstMatchHitsEndOfBuffer", func(t *testing.T) {
		splitFunc := NewLineStartSplitFunc(regexp.MustCompile("LOGSTART"))
		data := []byte(`LOGSTART`)

		t.Run("NotAtEOF", func(t *testing.T) {
			advance, token, err := splitFunc(data[:], false)
			require.NoError(t, err)
			require.Equal(t, 0, advance)
			require.Nil(t, token)
		})

		t.Run("AtEOF", func(t *testing.T) {
			advance, token, err := splitFunc(data[:], true)
			require.NoError(t, err)
			require.Equal(t, len(data), advance)
			require.Equal(t, data, token)
		})
	})
}

func TestLineEndSplitFunc(t *testing.T) {
	testCases := []tokenizerTestCase{
		{
			Name:    "OneLogSimple",
			Pattern: `LOGEND \d+`,
			Raw:     []byte(`my log LOGEND 123`),
			ExpectedTokenized: []string{
				`my log LOGEND 123`,
			},
		},
		{
			Name:    "TwoLogsSimple",
			Pattern: `LOGEND \d+`,
			Raw:     []byte(`log1 LOGEND 123log2 LOGEND 234`),
			ExpectedTokenized: []string{
				`log1 LOGEND 123`,
				`log2 LOGEND 234`,
			},
		},
		{
			Name:              "NoMatches",
			Pattern:           `LOGEND \d+`,
			Raw:               []byte(`file that has no matches in it`),
			ExpectedTokenized: []string{},
		},
		{
			Name:    "NonMatchesAfter",
			Pattern: `LOGEND \d+`,
			Raw:     []byte(`part that matches LOGEND 123 part that doesn't match`),
			ExpectedTokenized: []string{
				`part that matches LOGEND 123`,
			},
		},
		{
			Name:    "HugeLog100",
			Pattern: `LOGEND \d+`,
			Raw: func() []byte {
				newRaw := generatedByteSliceOfLength(100)
				newRaw = append(newRaw, []byte(`LOGEND 123`)...)
				return newRaw
			}(),
			ExpectedTokenized: []string{
				string(generatedByteSliceOfLength(100)) + `LOGEND 123`,
			},
		},
		{
			Name:    "HugeLog10000",
			Pattern: `LOGEND \d+`,
			Raw: func() []byte {
				newRaw := generatedByteSliceOfLength(10000)
				newRaw = append(newRaw, []byte(`LOGEND 123`)...)
				return newRaw
			}(),
			ExpectedTokenized: []string{
				string(generatedByteSliceOfLength(10000)) + `LOGEND 123`,
			},
		},
		{
			Name:    "HugeLog1000000",
			Pattern: `LOGEND \d+`,
			Raw: func() []byte {
				newRaw := generatedByteSliceOfLength(1000000)
				newRaw = append(newRaw, []byte(`LOGEND 123`)...)
				return newRaw
			}(),
			ExpectedTokenized: []string{},
			ExpectedError:     errors.New("bufio.Scanner: token too long"),
		},
	}

	for _, tc := range testCases {
		re := regexp.MustCompile(tc.Pattern)
		splitFunc := NewLineEndSplitFunc(re)
		t.Run(tc.Name, tc.RunFunc(splitFunc))
	}
}

func generatedByteSliceOfLength(length int) []byte {
	chars := []byte(`abcdefghijklmnopqrstuvwxyz`)
	newSlice := make([]byte, length)
	for i := 0; i < length; i++ {
		newSlice[i] = chars[i%len(chars)]
	}
	return newSlice
}
