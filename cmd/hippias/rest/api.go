// This file defines a small REST server for Anthem to access.

package rest

import (
	"net/http"

	"github.com/ChorusOne/Hippias/cmd/hippias/rest/endpoints"
	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"

	// Riddle me this, what should we call our middleware to disambiguate them
	// from chi's built in middleware.
	riddleware "github.com/ChorusOne/Hippias/cmd/hippias/rest/middleware"
)

// StartAPI creates an HTTP server with a REST API for Anthem to consume. This
// function blocks and so should be spawned as a goroutine.
func StartAPI(config *types.Config, state types.State) {
	// Prepare Endpoints and Chi Router
	r := chi.NewRouter()

	// Main Routing Table
	r.Get("/", IndexResponder(state))
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Use(riddleware.Paginate)
		r.Get("/account", endpoints.AccountList(state))
		r.Get("/account/describe", endpoints.AccountListDescribed(state))
		r.Get("/account/{accountID}", endpoints.Account(state))
		r.Get("/account/{accountID}/history", endpoints.AccountHistory(state))
		r.Get("/account/{accountID}/events", endpoints.EventList(state))
		r.Get("/account/{accountID}/transactions", endpoints.TransactionList(state))
		r.Get("/event", endpoints.EventList(state))
		r.Get("/transaction", endpoints.TransactionList(state))
	})

	// Expose Documentation
	r.Get("/api", Documentation(r))

	// Block & Serve
	http.ListenAndServe(":"+config.ListenPort, r)
}

// -----------------------------------------------------------------------------

// Handler is a small type that helps to make returning closures in the
// functions below less verbose.
type Handler = func(w http.ResponseWriter, r *http.Request)

// IndexResponder acts as a health check.
func IndexResponder(_ types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	}
}

// Documentation produces markdown output describing the passed router.
func Documentation(r *chi.Mux) Handler {
	documentation := docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
		ProjectPath: "Hippidaedyvitruvius",
		Intro:       "Currently exposed API endpoints for Oasis extraction.",
	})

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			"<meta charset='utf-8'>\n" +
				documentation +
				"\n<!-- Markdeep: --><style class=\"fallback\">body{visibility:hidden;white-space:pre;font-family:monospace}</style><script src=\"markdeep.min.js\" charset=\"utf-8\"></script><script src=\"https://casual-effects.com/markdeep/latest/markdeep.min.js\" charset=\"utf-8\"></script><script>window.alreadyProcessedMarkdeep||(document.body.style.visibility=\"visible\")</script>",
		))
	}
}
