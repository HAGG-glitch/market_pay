package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func performRequest(h gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	h(c)
	return w
}

func TestSuccess(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		Success(c, map[string]string{"key": "val"}, "ok")
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "ok", body["message"])
	assert.Equal(t, "val", body["data"].(map[string]interface{})["key"])
}

func TestCreated(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		Created(c, "new-resource", "created")
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "created", body["message"])
	assert.Equal(t, "new-resource", body["data"])
}

func TestPaginated(t *testing.T) {
	items := []int{1, 2, 3}
	w := performRequest(func(c *gin.Context) {
		Paginated(c, items, 30, 1, 10, "list")
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "list", body["message"])
	assert.EqualValues(t, 3, body["total_pages"])
	assert.EqualValues(t, 1, body["page"])
	assert.EqualValues(t, 10, body["limit"])
}

func TestPaginated_PartialPage(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	w := performRequest(func(c *gin.Context) {
		Paginated(c, items, 25, 3, 10, "page 3")
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.EqualValues(t, 3, body["total_pages"])
}

func TestError(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		Error(c, http.StatusBadRequest, "bad input")
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "bad input", body["message"])
}

func TestErrorFromAppError(t *testing.T) {
	appErr := apperrors.ErrUnauthorized("not allowed")
	w := performRequest(func(c *gin.Context) {
		ErrorFromAppError(c, appErr)
	})
	assert.Equal(t, appErr.Status, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, appErr.Message, body["message"])
}

func TestErrorFromAppError_WrapsNonAppError(t *testing.T) {
	w := performRequest(func(c *gin.Context) {
		ErrorFromAppError(c, http.ErrAbortHandler)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Contains(t, body["message"], "Something went wrong on our end")
}
