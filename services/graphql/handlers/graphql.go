package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
)

type GraphQLHandler struct {
	Schema *graphql.Schema
}

type requestBody struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func NewGraphQLHandler(schema *graphql.Schema) *GraphQLHandler {
	return &GraphQLHandler{
		Schema: schema,
	}
}

// ServeHTTP handles incoming POST GraphQL execution requests.
func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Support simple GET query execution for URL-based debugging
		if r.Method == http.MethodGet {
			query := r.URL.Query().Get("query")
			if query == "" {
				h.respondJSON(w, http.StatusBadRequest, map[string]string{
					"error": "Missing Query",
					"message": "GraphQL query parameter is required for GET requests",
				})
				return
			}
			result := graphql.Do(graphql.Params{
				Schema:        *h.Schema,
				RequestString: query,
			})
			h.respondJSON(w, http.StatusOK, result)
			return
		}

		w.Header().Set("Allow", "GET, POST")
		h.respondJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method Not Allowed",
			"message": "GraphQL endpoints only support GET and POST methods",
		})
		return
	}

	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Malformed JSON Request",
			"message": err.Error(),
		})
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         *h.Schema,
		RequestString:  body.Query,
		VariableValues: body.Variables,
		OperationName:  body.OperationName,
	})

	h.respondJSON(w, http.StatusOK, result)
}

func (h *GraphQLHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
