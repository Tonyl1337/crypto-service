package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {

	recorder := httptest.NewRecorder()

	type testResponse struct {
		Message string `json:"message"`
	}

	expected := testResponse{
		Message: "hello",
	}

	WriteJSON(
		recorder,
		http.StatusOK,
		expected,
	)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	require.Equal(
		t,
		"application/json",
		recorder.Header().Get("Content-Type"),
	)

	var actual testResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		expected,
		actual,
	)
}

func TestWriteError(t *testing.T) {

	recorder := httptest.NewRecorder()

	expectedErr := "database unavailable"

	WriteError(
		recorder,
		http.StatusInternalServerError,
		errors.New(expectedErr),
	)

	require.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	var actual Error

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		expectedErr,
		actual.Error,
	)
}
