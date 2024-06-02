package api

import (
	"bytes"
	"encoding/json"
	mockapi "github.com/HyperGAI/serving-api/api/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTasks(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(webhook *mockapi.MockWebhook)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"ids": []string{
					"1234", "5678",
				},
			},
			buildStubs: func(webhook *mockapi.MockWebhook) {
				webhook.EXPECT().
					GetTaskInfo(gomock.Any()).
					Times(2).
					Return(map[string]string{"outputs": "test"}, nil)
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

			webhook := mockapi.NewMockWebhook(ctrl)
			tc.buildStubs(webhook)

			server := newTestServer(t, webhook)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/task/batch", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("UID", "12345")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
