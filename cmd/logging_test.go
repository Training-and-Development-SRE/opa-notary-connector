package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryLogging(t *testing.T) {
	t.Parallel()
	r := setupServer()
	r.GET("/panic", func(c *gin.Context) {
		panic("testing recovery logger")
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	resp, err := http.Get(fmt.Sprintf("%s/panic", ts.URL))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

}
