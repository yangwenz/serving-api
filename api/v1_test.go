package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	mockplatform "github.com/yangwenz/model-serving/platform/mock"
	mockwk "github.com/yangwenz/model-serving/worker/mock"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPredictV1(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(platform *mockplatform.MockPlatform)
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
			buildStubs: func(platform *mockplatform.MockPlatform) {
				platform.EXPECT().
					Predict(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, nil)
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
			distributor := mockwk.NewMockTaskDistributor(ctrl)
			webhook := mockplatform.NewMockWebhook(ctrl)
			tc.buildStubs(platform)

			server := newTestServer(t, platform, distributor, webhook)
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

func TestAsyncPredictV1(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(distributor *mockwk.MockTaskDistributor, webhook *mockplatform.MockWebhook)
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
			buildStubs: func(distributor *mockwk.MockTaskDistributor, webhook *mockplatform.MockWebhook) {
				distributor.EXPECT().
					DistributeTaskRunPrediction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				webhook.EXPECT().CreateNewTask(gomock.Eq("test_model"), gomock.Eq("v1")).
					Times(1).
					Return("success", nil)
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
			distributor := mockwk.NewMockTaskDistributor(ctrl)
			webhook := mockplatform.NewMockWebhook(ctrl)
			tc.buildStubs(distributor, webhook)

			server := newTestServer(t, platform, distributor, webhook)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/async/v1/predict", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
