package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractExpectedSequence(t *testing.T) {
	tests := []struct {
		name          string
		errorMsg      string
		expectedSeq   uint64
		expectedFound bool
	}{
		{
			name:          "valid error message with expected sequence",
			errorMsg:      "error while simulating tx: rpc error: code = Unknown desc = account sequence mismatch, expected 1471614, got 1471613: incorrect account sequence",
			expectedSeq:   1471614,
			expectedFound: true,
		},
		{
			name:          "error message from log example",
			errorMsg:      "account sequence mismatch, expected 1471614, got 1471613: incorrect account sequence [cosmos/cosmos-sdk@v0.45.17/x/auth/ante/sigverify.go:264] With gas wanted: '0' and gas used: '5043313'",
			expectedSeq:   1471614,
			expectedFound: true,
		},
		{
			name:          "error message with different sequence numbers",
			errorMsg:      "account sequence mismatch, expected 999999, got 999998",
			expectedSeq:   999999,
			expectedFound: true,
		},
		{
			name:          "error message with zero sequence",
			errorMsg:      "account sequence mismatch, expected 0, got 1",
			expectedSeq:   0,
			expectedFound: true,
		},
		{
			name:          "error message with large sequence number",
			errorMsg:      "account sequence mismatch, expected 18446744073709551615, got 18446744073709551614",
			expectedSeq:   18446744073709551615,
			expectedFound: true,
		},
		{
			name:          "error message without expected keyword",
			errorMsg:      "account sequence mismatch, got 1471613: incorrect account sequence",
			expectedSeq:   0,
			expectedFound: false,
		},
		{
			name:          "error message with account sequence mismatch but no numbers",
			errorMsg:      "account sequence mismatch: incorrect account sequence",
			expectedSeq:   0,
			expectedFound: false,
		},
		{
			name:          "empty error message",
			errorMsg:      "",
			expectedSeq:   0,
			expectedFound: false,
		},
		{
			name:          "error message with unrelated expected keyword",
			errorMsg:      "something went wrong, expected something else",
			expectedSeq:   0,
			expectedFound: false,
		},
		{
			name:          "error message with expected but no number",
			errorMsg:      "account sequence mismatch, expected abc, got 1471613",
			expectedSeq:   0,
			expectedFound: false,
		},
		{
			name:          "error message with multiple expected keywords",
			errorMsg:      "account sequence mismatch, expected 1471614, got 1471613, but expected something else",
			expectedSeq:   1471614,
			expectedFound: true,
		},
		{
			name:          "error message with sequence in different format (commas)",
			errorMsg:      "account sequence mismatch, expected 1,471,614, got 1,471,613",
			expectedSeq:   1, // Regex will match the first digit sequence "1"
			expectedFound: true,
		},
		{
			name:          "error message with negative number (should not match)",
			errorMsg:      "account sequence mismatch, expected -1471614, got 1471613",
			expectedSeq:   0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq, found := extractExpectedSequence(tt.errorMsg)
			require.Equal(t, tt.expectedFound, found, "found flag should match")
			if tt.expectedFound {
				assert.Equal(t, tt.expectedSeq, seq, "sequence number should match")
			} else {
				assert.Equal(t, uint64(0), seq, "sequence should be 0 when not found")
			}
		})
	}
}

func TestExtractExpectedSequence_EdgeCases(t *testing.T) {
	t.Run("very long error message", func(t *testing.T) {
		longMsg := "error while simulating tx: rpc error: code = Unknown desc = account sequence mismatch, expected 1234567890, got 1234567889: incorrect account sequence [cosmos/cosmos-sdk@v0.45.17/x/auth/ante/sigverify.go:264] With gas wanted: '0' and gas used: '5043313' and some other very long text that goes on and on"
		seq, found := extractExpectedSequence(longMsg)
		require.True(t, found)
		assert.Equal(t, uint64(1234567890), seq)
	})

	t.Run("error message with whitespace", func(t *testing.T) {
		msg := "account sequence mismatch, expected  1471614  , got 1471613"
		seq, found := extractExpectedSequence(msg)
		require.True(t, found)
		assert.Equal(t, uint64(1471614), seq)
	})

	t.Run("error message with newlines", func(t *testing.T) {
		msg := "account sequence mismatch,\nexpected 1471614,\ngot 1471613"
		seq, found := extractExpectedSequence(msg)
		require.True(t, found)
		assert.Equal(t, uint64(1471614), seq)
	})
}
