package parser

import (
	"context"
	"testing"
	"time"

	"github.com/observiq/carbon/entry"
	"github.com/observiq/carbon/internal/testutil"
	"github.com/observiq/carbon/plugin"
	"github.com/observiq/carbon/plugin/helper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyslogParser(t *testing.T) {
	basicConfig := func() *SyslogParserConfig {
		return &SyslogParserConfig{
			ParserConfig: helper.ParserConfig{
				TransformerConfig: helper.TransformerConfig{
					BasicConfig: helper.BasicConfig{
						PluginID:   "test_plugin_id",
						PluginType: "syslog_parser",
					},
					WriterConfig: helper.WriterConfig{
						OutputIDs: []string{"output1"},
					},
				},
				ParseFrom: entry.NewRecordField(),
				ParseTo:   entry.NewRecordField(),
			},
		}
	}

	cases := []struct {
		name              string
		config            *SyslogParserConfig
		inputRecord       interface{}
		expectedTimestamp time.Time
		expectedRecord    interface{}
	}{
		{
			"RFC3164",
			func() *SyslogParserConfig {
				cfg := basicConfig()
				cfg.Protocol = "rfc3164"
				return cfg
			}(),
			"<34>Jan 12 06:30:00 1.2.3.4 apache_server: test message",
			time.Date(time.Now().Year(), 1, 12, 6, 30, 0, 0, time.UTC),
			map[string]interface{}{
				"appname":  "apache_server",
				"facility": 4,
				"hostname": "1.2.3.4",
				"message":  "test message",
				"priority": 34,
				"severity": 2,
			},
		},
		{
			"RFC3164Bytes",
			func() *SyslogParserConfig {
				cfg := basicConfig()
				cfg.Protocol = "rfc3164"
				return cfg
			}(),
			[]byte("<34>Jan 12 06:30:00 1.2.3.4 apache_server: test message"),
			time.Date(time.Now().Year(), 1, 12, 6, 30, 0, 0, time.UTC),
			map[string]interface{}{
				"appname":  "apache_server",
				"facility": 4,
				"hostname": "1.2.3.4",
				"message":  "test message",
				"priority": 34,
				"severity": 2,
			},
		},
		{
			"RFC5424",
			func() *SyslogParserConfig {
				cfg := basicConfig()
				cfg.Protocol = "rfc5424"
				return cfg
			}(),
			`<86>1 2015-08-05T21:58:59.693Z 192.168.2.132 SecureAuth0 23108 ID52020 [SecureAuth@27389 UserHostAddress="192.168.2.132" Realm="SecureAuth0" UserID="Tester2" PEN="27389"] Found the user for retrieving user's profile`,
			time.Date(2015, 8, 5, 21, 58, 59, 693000000, time.UTC),
			map[string]interface{}{
				"appname":  "SecureAuth0",
				"facility": 10,
				"hostname": "192.168.2.132",
				"message":  "Found the user for retrieving user's profile",
				"msg_id":   "ID52020",
				"priority": 86,
				"proc_id":  "23108",
				"severity": 6,
				"structured_data": map[string]map[string]string{
					"SecureAuth@27389": {
						"PEN":             "27389",
						"Realm":           "SecureAuth0",
						"UserHostAddress": "192.168.2.132",
						"UserID":          "Tester2",
					},
				},
				"version": 1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buildContext := testutil.NewBuildContext(t)
			newPlugin, err := tc.config.Build(buildContext)
			require.NoError(t, err)
			syslogParser := newPlugin.(*SyslogParser)

			mockOutput := testutil.NewMockPlugin("output1")
			entryChan := make(chan *entry.Entry, 1)
			mockOutput.On("Process", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				entryChan <- args.Get(1).(*entry.Entry)
			}).Return(nil)

			err = syslogParser.SetOutputs([]plugin.Plugin{mockOutput})
			require.NoError(t, err)

			newEntry := entry.New()
			newEntry.Record = tc.inputRecord
			err = syslogParser.Process(context.Background(), newEntry)
			require.NoError(t, err)

			select {
			case e := <-entryChan:
				require.Equal(t, e.Record, tc.expectedRecord)
				require.Equal(t, tc.expectedTimestamp, e.Timestamp)
			case <-time.After(time.Second):
				require.FailNow(t, "Timed out waiting for entry to be processed")
			}
		})
	}
}
