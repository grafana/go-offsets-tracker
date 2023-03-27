package offsets

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFieldOffset(t *testing.T) {
	dataFile := `{
	"data" : {
		"struct_1" : { 
			"field_1" : {
				"offsets": [
					{ "offset": 1187, "since": "1.18.7" },
					{ "offset": 1190, "since": "1.19.0" }
				]
			}
		}
	}
}`
	tracker, err := Read(bytes.NewBufferString(dataFile))
	require.NoError(t, err)

	offset, ok := tracker.Find("struct_1", "field_1", "1.19.7")
	assert.True(t, ok)
	assert.Equal(t, 1190, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.19.0")
	assert.True(t, ok)
	assert.Equal(t, 1190, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.18.9")
	assert.True(t, ok)
	assert.Equal(t, 1187, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.17.9")
	assert.Falsef(t, ok, "found: %d", int(offset))
}
