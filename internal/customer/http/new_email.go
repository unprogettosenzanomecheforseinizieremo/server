package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/metabs/server/customer"
	"github.com/metabs/server/email"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"net/http"
)

type newEmailReq struct {
	Email customer.Email `json:"email"`
}

func (r *newEmailReq) UnmarshalJSON(data []byte) error {
	type clone newEmailReq
	var req clone
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}

	var err error
	if r.Email, err = customer.NewEmail(req.Email.String()); err != nil {
		return err
	}

	return nil
}

func newEmail(repo customer.Repo, sender *email.Sender, log *zap.SugaredLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), "new email")
		defer span.End()

		logger := log.With("trace_id", span.SpanContext().TraceID.String(), "action", "new email")

		id, err := customer.NewID(chi.URLParam(r, "id"))
		if err != nil {
			logger.With("error", err).Info("could not create id")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger = logger.With("id", id)

		var rb newEmailReq
		switch err := json.NewDecoder(r.Body).Decode(&rb); {
		case errors.Is(err, customer.ErrInvalidEmail):
			w.WriteHeader(http.StatusBadRequest)
			if _, err2 := w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`,err.Error()))); err2 != nil {
				logger.With("error", err, "error_2", err2).Error("could not write response")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			logger.With("error", err, "email", rb.Email).Info("bad request")
			return

		case err != nil:
			logger.With("error", err).Error("could not decode request")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch _, err := repo.GetByEmail(ctx, rb.Email); {
		case err != nil && !errors.Is(err, customer.ErrNotFound):
			logger.With("error", err).Error("could not get customer")
			w.WriteHeader(http.StatusInternalServerError)
			return

		case err == nil:
			logger.With("error", err).Error("customer already exists")
			w.WriteHeader(http.StatusConflict)
			return
		}

		logger = logger.With("id", id)
		c, err := repo.Get(ctx, id)
		if err != nil {
			logger.With("error", err).Warn("could not find customer")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		c.GenerateChangeEmailHash()

		if err := repo.Add(ctx, c); err != nil {
			logger.With("error", err).Error("could not add customer")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: when not a MVP use a queue
		if err := sender.SendChangeEmail(ctx, c.Email.String(), c.ID.String(), c.ChangeEmailHash); err != nil {
			logger.With("error", err).Error("could not send email")
		}

		if err := json.NewEncoder(w).Encode(c); err != nil {
			logger.With("error", err).Error("could not write response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
