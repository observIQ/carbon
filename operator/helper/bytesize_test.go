package helper

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type testCase struct {
	input       string
	expected    ByteSize
	expectError bool
}

var sharedTestCases = []testCase{
	{`1`, 1, false},
	{`3.3`, 3, false},
	{`0`, 0, false},
	{`10101010`, 10101010, false},
	{`0.01`, 0, false},
	{`"1"`, 1, false},
	{`"1kb"`, 1000, false},
	{`"1KB"`, 1000, false},
	{`"1kib"`, 1024, false},
	{`"1KiB"`, 1024, false},
	{`"1mb"`, 1000 * 1000, false},
	{`"1mib"`, 1024 * 1024, false},
	{`"1gb"`, 1000 * 1000 * 1000, false},
	{`"1gib"`, 1024 * 1024 * 1024, false},
	{`"1tb"`, 1000 * 1000 * 1000 * 1000, false},
	{`"1tib"`, 1024 * 1024 * 1024 * 1024, false},
	{`"1pB"`, 1000 * 1000 * 1000 * 1000 * 1000, false},
	{`"1pib"`, 1024 * 1024 * 1024 * 1024 * 1024, false},
	{`1e3`, 1000, false},
	{`"3ii3"`, 0, true},
	{`3ii3`, 0, true},
	{`--ii3`, 0, true},
	{`{"test":"val"}`, 0, true},
	{`1e3`, 1000, false},
}

func TestByteSizeUnmarshalJSON(t *testing.T) {
	for i, tc := range sharedTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var bs ByteSize
			err := json.Unmarshal([]byte(tc.input), &bs)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, bs)
		})
	}
}

func TestByteSizeUnmarshalYAML(t *testing.T) {
	additionalCases := []testCase{
		{`1kb`, 1000, false},
		{`1KB`, 1000, false},
		{`1kib`, 1024, false},
		{`1KiB`, 1024, false},
		{`1mb`, 1000 * 1000, false},
		{`1mib`, 1024 * 1024, false},
		{`1gb`, 1000 * 1000 * 1000, false},
		{`1gib`, 1024 * 1024 * 1024, false},
		{`1tb`, 1000 * 1000 * 1000 * 1000, false},
		{`1tib`, 1024 * 1024 * 1024 * 1024, false},
		{`1pB`, 1000 * 1000 * 1000 * 1000 * 1000, false},
		{`1pib`, 1024 * 1024 * 1024 * 1024 * 1024, false},
		{`test: val`, 0, true},
	}

	cases := append(sharedTestCases, additionalCases...)
	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var bs ByteSize
			err := yaml.Unmarshal([]byte(tc.input), &bs)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, bs)
		})
	}
}
