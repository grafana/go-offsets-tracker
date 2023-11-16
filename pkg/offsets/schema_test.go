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

func TestGetFieldOffset_UnorthodoxVersions(t *testing.T) {
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

	// X:boringcrypto suffix will be removed, and we will assume 1.19.0
	offset, ok := tracker.Find("struct_1", "field_1", "1.19.0 X:boringcrypto")
	assert.True(t, ok)
	assert.Equal(t, 1190, int(offset))
	// -prerrelease suffix will be kept. This means that the provided version is < 1.19.0
	offset, ok = tracker.Find("struct_1", "field_1", "1.19.0-prerrelease")
	assert.True(t, ok)
	assert.Equal(t, 1187, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.20.0-prerrelease")
	assert.True(t, ok)
	assert.Equal(t, 1190, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.18.9.33")
	assert.True(t, ok)
	assert.Equal(t, 1187, int(offset))
	offset, ok = tracker.Find("struct_1", "field_1", "1.17.9#yahooooii")
	assert.Falsef(t, ok, "found: %d", int(offset))
}
