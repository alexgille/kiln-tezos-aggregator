package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kiln-tezos-delegation/repository"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Controller interface {
	GetDelegations(context.Context, int) ([]repository.Delegation, error)
}

// GetDelegationHandler handles GET requests to fetch delegations, possibly filtered
// for a given year. The year query parameter must be in YYYY format.
// Responds with a specific HTTP status if method or year query parameter are invalid.
func GetDelegationHandler(ctrl Controller) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		if request.Method != http.MethodGet {
			resp.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		yearParam := YearNotSpecified
		if val := request.URL.Query().Get("year"); val != "" {
			_, err := fmt.Sscanf(val, "%4d", &yearParam)
			if err != nil || len(val) != 4 {
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		type Delegation struct {
			Timestamp string `json:"timestamp"`
			Amount    string `json:"amount"`
			Delegator string `json:"delegator"`
			Level     string `json:"level"`
		}

		type Response struct {
			Data []Delegation `json:"data"`
		}

		dlgs, err := ctrl.GetDelegations(request.Context(), yearParam)

		data := make([]Delegation, len(dlgs))
		for i := range dlgs {
			data[i].Timestamp = dlgs[i].BlockTimestamp.UTC().Format(time.RFC3339)
			data[i].Amount = strconv.Itoa(int(dlgs[i].Amount))
			data[i].Delegator = dlgs[i].Sender
			data[i].Level = strconv.Itoa(int(dlgs[i].Level))
		}

		switch {
		case errors.Is(err, nil):
			resp.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(resp).Encode(Response{
				Data: data,
			})
		default:
			resp.WriteHeader(http.StatusInternalServerError)
			log.Default().Println("api handler error: ", err.Error())
		}
	})
}
