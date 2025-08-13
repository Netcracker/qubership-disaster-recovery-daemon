package v1

import (
	"encoding/json"
	"fmt"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/internal/usecase"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	verbose = GetEnv("DEBUG", "false")
)

func NewServerHandler(authenticator Authenticator) *ServerHandler {
	router := mux.NewRouter()
	return &ServerHandler{
		router:        router,
		authenticator: authenticator,
	}
}

type ServerHandler struct {
	router        *mux.Router
	authenticator Authenticator
}

func (sh *ServerHandler) NewHealthRoute(useCase usecase.ReadMode) {
	sh.router.Handle("/health", http.HandlerFunc(getHealth(useCase))).Methods(http.MethodGet)
}

func (sh *ServerHandler) NewHealthzRoute(useCase usecase.Health) {
	sh.router.Handle("/healthz", http.HandlerFunc(sh.authenticationWrapper(getClusterHealthStatus(useCase)))).Methods(http.MethodGet)
}

func (sh *ServerHandler) NewReadModeRoute(useCase usecase.ReadMode) {
	sh.router.Handle("/sitemanager", http.HandlerFunc(sh.authenticationWrapper(getModeAndStatus(useCase)))).Methods(http.MethodGet)
}

func (sh *ServerHandler) NewUpdateModeRoute(useCase usecase.SetMode) {
	sh.router.Handle("/sitemanager", http.HandlerFunc(sh.authenticationWrapper(setMode(useCase)))).Methods(http.MethodPost)
}

func (sh *ServerHandler) BuildHandler() http.Handler {
	return JsonContentType(handlers.CompressHandler(sh.router))
}

func JsonContentType(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}

func (sh *ServerHandler) authenticationWrapper(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter,
	r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if authenticated, token := sh.authenticator.CheckAuth(r); !authenticated {
			log.Println("Unauthorized request.")
			if token == "" {
				w.Header().Set("WWW-Authenticate", "Bearer")
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Access denied"))
			if err != nil {
				log.Printf("Can not write authentication response due to: %v", err)
			}
			return
		}
		f(w, r)
	}
}

func getHealth(useCase usecase.ReadMode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := useCase.GetModeAndStatus()
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, err)
		} else {
			sendSuccessfulResponse(w, state)
		}
	}
}

func getClusterHealthStatus(useCase usecase.Health) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("New request for disaster recovery service health has been received.")
		healthState, err := useCase.GetHealth()
		if err != nil {
			log.Printf("Can not get the service health. Error is [%v]", err)
			sendFailedHealthResponse(w)
			return
		}
		log.Printf("The disaster recovery health state is [%v]", healthState)
		sendSuccessfulResponse(w, healthState)
	}
}

func getModeAndStatus(useCase usecase.ReadMode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("New request for disaster recovery status has been received.")
		state, err := useCase.GetModeAndStatus()
		if err != nil {
			comment := fmt.Sprintf("Can not get a disaster recovery state. Error is [%v]", err)
			log.Println(comment)
			sendFailedSwitchoverResponse(w, "", comment)
			return
		}
		log.Printf("The disaster recovery status is [%v]", state)
		sendSuccessfulResponse(w, state)
	}
}

func setMode(useCase usecase.SetMode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data entity.RequestData
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			comment := fmt.Sprintf("Reading data from request failed. Error is [%v]", err)
			log.Println(comment)
			sendFailedSwitchoverResponse(w, "", comment)
			return
		}
		err = json.Unmarshal(requestBody, &data)
		if err != nil {
			comment := fmt.Sprintf("Unmarshalling data from request failed. Error is [%v]", err)
			log.Println(comment)
			sendFailedSwitchoverResponse(w, "", comment)
			return
		}
		log.Printf("New request for disaster recovery mode changing has been received. The request body is [%v]", data)

		switchoverState, err := useCase.SetDrMode(data)
		if err != nil {
			log.Printf("can not set disaster recovery mode. Error is [%v]", err)
			sendFailedSwitchoverResponse(w, switchoverState.Mode, switchoverState.Comment)
			return
		}
		sendSuccessfulResponse(w, switchoverState)
	}
}

func sendSuccessfulResponse(w http.ResponseWriter, response interface{}) {
	sendResponse(w, http.StatusOK, response)
}

func sendFailedHealthResponse(w http.ResponseWriter) {
	response := entity.HealthResponse{
		Status: entity.DOWN,
	}
	sendResponse(w, http.StatusInternalServerError, response)
}

func sendFailedSwitchoverResponse(w http.ResponseWriter, mode string, comment string) {
	response := entity.SwitchoverState{
		Mode:    mode,
		Status:  entity.FAILED,
		Comment: comment,
	}
	sendResponse(w, http.StatusInternalServerError, response)
}

func sendResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.WriteHeader(statusCode)
	responseBody, _ := json.Marshal(response)
	if verbose == "true" {
		fmt.Printf("Response body: %s\n", responseBody)

	}
	_, _ = w.Write(responseBody)
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
