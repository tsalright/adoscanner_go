package ado

import (
	mocks "adoscanner/mocks/ado"
	"bytes"
	"encoding/json"
	"github.com/alicebob/miniredis"
	"github.com/elliotchance/redismock"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestRedis returns a redis.Cmdable.
func newTestRedis() *redismock.ClientMock {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return redismock.NewNiceMock(client)
}

func Router(mockConnection *mocks.Service, mockClient redis.Cmdable, mockLogging *mocks.Logging) *mux.Router {
	api := API{
		adoService: mockConnection,
		logger: mockLogging,
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/Results", api.postCacheHandler(mockClient)).Methods("POST")
	return router
}

func TestPostMissingContentHeader(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	criteria := &SearchCriteria{
		ProjectNamePattern: "",
		FileNamePattern:    "",
		ContentPattern:     "",
	}

	jsonCriteria, _ := json.Marshal(criteria)
	req, err := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 415, rr.Code)
	assert.Equal(t, "Content-Type header is not application/json\n", rr.Body.String())
}

func TestPostWithContentHeaderMissingOrgHeader(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	criteria := &SearchCriteria{
		ProjectNamePattern: "",
		FileNamePattern:    "",
		ContentPattern:     "",
	}

	jsonCriteria, _ := json.Marshal(criteria)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Org header is required\n", rr.Body.String())
}

func TestPostWithContentAndOrgHeaderMissingPATHeader(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	criteria := &SearchCriteria{
		ProjectNamePattern: "",
		FileNamePattern:    "",
		ContentPattern:     "",
	}

	jsonCriteria, _ := json.Marshal(criteria)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "PAT header is required\n", rr.Body.String())
}

func TestPostReturnsBadRequestForBadlyFormedJSONWithLocation(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	jsonCriteria := []byte(`{"1":"1","2":"2","3":"3",}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body contains badly-formed JSON (at position 26)\n", rr.Body.String())
}

func TestPostReturnsBadRequestForBadlyFormedJson(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	jsonCriteria := []byte(`{"1":"1","2":"2","3":"3",`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body contains badly-formed JSON\n", rr.Body.String())
}

func TestPostReturnsBadRequestForInvalidValueForField(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	jsonCriteria := []byte(`{"ProjectNamePattern":"1","FileNamePattern":"2","ContentPattern":3}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body contains an invalid value for the \"ContentPattern\" field (at position 66)\n", rr.Body.String())
}

func TestPostReturnsBadRequestForExtraField(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	jsonCriteria := []byte(`{"ProjectNamePattern":"1","FileNamePattern":"2","ContentPattern":"3", "1":"1"}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body contains unknown field \"1\"\n", rr.Body.String())
}

func TestPostReturnsBadRequestForEmptyBody(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()
	jsonCriteria := []byte(``)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body must not be empty\n", rr.Body.String())
}

func TestPostReturnsRequestEntityTooLarge(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()

	var pnp bytes.Buffer
	var fnp bytes.Buffer
	var cp bytes.Buffer
	for i := 0; i < 100000; i++ {
		pnp.WriteString("ABC")
		fnp.WriteString("123")
		cp.WriteString("ABC123")
	}
	jsonCriteria := "{\"ProjectNamePattern\":\"" + pnp.String() + "\",\"FileNamePattern\":\"" + fnp.String() + "\",\"ContentPattern\":\"" + cp.String() + "\"}"
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer([]byte(jsonCriteria)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 413, rr.Code)
	assert.Equal(t, "Request body must not be larger than 1MB\n", rr.Body.String())
}

func TestPostReturnsBadRequestSingleJsonOnly(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockLogging := new(mocks.Logging)
	mockRedis := newTestRedis()

	jsonCriteria := []byte(`{"ProjectNamePattern":"1","FileNamePattern":"2","ContentPattern":"3"}{"ProjectNamePattern":"1","FileNamePattern":"2","ContentPattern":"3"}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "Request body must only contain a single JSON object\n", rr.Body.String())
}

func TestPostReturnsOKNoCache(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockConnection.On("CreateConnection", mock.Anything, mock.Anything).Return(nil)
	mockConnection.On("GetProjects", mock.Anything).Return(new(core.GetProjectsResponseValue), nil)

	mockLogging := new(mocks.Logging)
	mockLogging.On("LogInfo", mock.Anything)

	mockRedis := newTestRedis()
	mockRedis.On("Get", mock.Anything).Return(redis.NewStringResult("", redis.Nil))
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewStatusResult("OK", nil))

	jsonCriteria := []byte(`{"ProjectNamePattern":"11","FileNamePattern":"22","ContentPattern":"33"}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	mockConnection.AssertNumberOfCalls(t, "GetProjects", 1)
	mockConnection.AssertNumberOfCalls(t, "GetAdditionalProjects", 0)
	mockConnection.AssertNumberOfCalls(t, "GetRepositories", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItems", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItemContent", 0)
	mockConnection.AssertExpectations(t)
	mockLogging.AssertNumberOfCalls(t,"LogInfo", 1)
	mockLogging.AssertExpectations(t)
	mockRedis.AssertNumberOfCalls(t,"Get", 1)
	mockRedis.AssertNumberOfCalls(t, "Set", 1)
	mockRedis.AssertExpectations(t)
}

func TestPostReturnsOKWithInvalidRedisConfiguration(t *testing.T) {
	mockConnection := new(mocks.Service)
	mockConnection.On("CreateConnection", mock.Anything, mock.Anything).Return(nil)
	mockConnection.On("GetProjects", mock.Anything).Return(new(core.GetProjectsResponseValue), nil)

	mockLogging := new(mocks.Logging)
	mockLogging.On("LogError", mock.Anything)
	mockLogging.On("LogInfo", mock.Anything)

	r := redis.NewClient(&redis.Options{
		Addr:     "bogus:6380",
		Password: "nochance",
		DB:       0,
	})

	jsonCriteria := []byte(`{"ProjectNamePattern":"11","FileNamePattern":"22","ContentPattern":"33"}`)
	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, r, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	mockConnection.AssertNumberOfCalls(t, "GetProjects", 1)
	mockConnection.AssertNumberOfCalls(t, "GetAdditionalProjects", 0)
	mockConnection.AssertNumberOfCalls(t, "GetRepositories", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItems", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItemContent", 0)
	mockConnection.AssertExpectations(t)
	mockLogging.AssertNumberOfCalls(t, "LogInfo", 1)
	mockLogging.AssertNumberOfCalls(t, "LogError", 2)
	mockLogging.AssertExpectations(t)
}

func TestPostReturnsOKWithCache(t *testing.T) {
	mockConnection := new(mocks.Service)

	mockLogging := new(mocks.Logging)
	mockLogging.On("LogInfo", mock.Anything)

	jsonCriteria := []byte(`{"ProjectNamePattern":"11","FileNamePattern":"22","ContentPattern":"33"}`)
	mockRedis := newTestRedis()
	mockRedis.On("Get", mock.Anything).Return(redis.NewStringResult(`{"ProjectNamePattern":"11","FileNamePattern":"22","ContentPattern":"33"}`, nil))

	req, _ := http.NewRequest("POST", "/api/v1/Results", bytes.NewBuffer(jsonCriteria))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Org", "itsals")
	req.Header.Add("PAT", "123")
	rr := httptest.NewRecorder()
	Router(mockConnection, mockRedis, mockLogging).ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	mockConnection.AssertNumberOfCalls(t, "GetProjects", 0)
	mockConnection.AssertNumberOfCalls(t, "GetAdditionalProjects", 0)
	mockConnection.AssertNumberOfCalls(t, "GetRepositories", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItems", 0)
	mockConnection.AssertNumberOfCalls(t, "GetItemContent", 0)
	mockConnection.AssertExpectations(t)
	mockLogging.AssertNumberOfCalls(t, "LogInfo", 1)
	mockLogging.AssertExpectations(t)
	mockRedis.AssertNumberOfCalls(t,"Get", 1)
	mockRedis.AssertNumberOfCalls(t, "Set", 0)
	mockRedis.AssertExpectations(t)
}