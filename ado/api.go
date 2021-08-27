package ado

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// API provides access to the RestApi functions and uses the Service interface for interacting with Azure DevOps
type API struct {
	adoService Service
	logger Logging
}

func (api *API) decodeSearchCriteria(w http.ResponseWriter, body io.ReadCloser) (criteria *SearchCriteria) {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&criteria)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return nil
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return nil
	}
	return criteria
}

func (api *API) postCacheHandler(client redis.Cmdable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		org := r.Header.Get("Org")
		if org == "" {
			http.Error(w, "Org header is required", http.StatusBadRequest)
			return
		}
		personalAccessToken := r.Header.Get("PAT")
		if personalAccessToken == "" {
			http.Error(w, "PAT header is required", http.StatusBadRequest)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		criteria := api.decodeSearchCriteria(w, r.Body)
		if criteria == nil {
			return
		}

		redisKey := fmt.Sprintf("%s%s%s%s", org, criteria.ProjectNamePattern, criteria.FileNamePattern, criteria.ContentPattern)

		val := api.getContentFromRedis(client, redisKey)
		if val == "" {
			msg := fmt.Sprintf("Cache miss for %s", redisKey)
			api.logger.LogInfo(msg)
			log.Println(msg)
			response, e := api.getContentFromAdo(org, personalAccessToken, criteria)
			if e != nil {
				api.logger.LogError(e)

				if e.Error() == "unable to connect to azure devops" {
					http.Error(w, e.Error(), http.StatusServiceUnavailable)
					return
				}

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			err := client.Set(redisKey, *response, time.Hour*24).Err()
			if err != nil {
				api.logger.LogError(err)
				log.Println(err)
			}
			if api.processResponse(w, *response) {
				return
			}
		} else {
			msg := fmt.Sprintf("Cache Hit for %s", redisKey)
			api.logger.LogInfo(msg)
			log.Println(msg)
			if api.processResponse(w, []byte(val)) {
				return
			}
		}
	}
}

func (api *API) getContentFromRedis(client redis.Cmdable, redisKey string) string {
	val, err := client.Get(redisKey).Result()
	if err == redis.Nil {
		return ""
	}
	if err != nil {
		api.logger.LogError(err)
		log.Printf("unable to connect to redis: %s", err)
		return ""
	}
	return val
}

func (api *API) processResponse(w http.ResponseWriter, val []byte) bool {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(val)
	if err != nil {
		api.logger.LogError(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return true
	}
	return false
}

func (api *API) getContentFromAdo(org, personalAccessToken string, criteria *SearchCriteria) (*[]byte, error) {
	organizationURL := fmt.Sprintf("https://dev.azure.com/%s", org)

	err := api.adoService.CreateConnection(organizationURL, personalAccessToken)
	if err != nil {
		log.Printf("AzureDevOpsService Failure")
		return nil, err
	}

	scanProjects := ScanProjects{
		adoService: api.adoService,
		criteria:   criteria,
		logger: api.logger,
	}

	results, err := scanProjects.Scan()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	result, err := json.Marshal(results)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &result, nil
}

func (api *API) healthHander(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// InitializeServer wires everything up to run the RestApi server
func InitializeServer() *http.Server {
	var (
		api = API{
			adoService: new(AzureDevOpsService),
			logger: new(AppInsightsLogger),
		}
		host = getEnv("REDIS_HOST", "localhost")
		port = getEnv("REDIS_PORT", ":6380")
		password = getEnv("REDIS_PASSWORD", "")
	)

	client := redis.NewClient(&redis.Options{
		Addr:        host + port,
		Password:    password,
		DB:          0,
		TLSConfig: 	 &tls.Config{},
	})

	r := mux.NewRouter()
	r.HandleFunc("/", api.postCacheHandler(client)).Methods(http.MethodPost)
	r.HandleFunc("/health", api.healthHander).Methods(http.MethodGet)

	srv := &http.Server{
		Handler: r,
		Addr: ":8080",
		ReadTimeout: 60 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	// Configure Logging
	LogFileLocation := os.Getenv("LOG_FILE_LOCATION")
	if LogFileLocation != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   LogFileLocation,
			MaxSize:    500,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		})
	}
	return srv
}