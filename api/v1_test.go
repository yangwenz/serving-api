package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	mockplatform "github.com/yangwenz/model-serving/platform/mock"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPredictV1(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"model_name":    "test_model",
				"model_version": "v1",
				"inputs": map[string][][]float32{
					"instances": {
						[]float32{6.8, 2.8, 4.8, 1.4},
						[]float32{6.0, 3.4, 4.5, 1.6},
					},
				},
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			platform := mockplatform.NewMockPlatform(ctrl)
			server := newTestServer(t, platform)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/v1/predict", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
