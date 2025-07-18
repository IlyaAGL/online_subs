package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/agl/online_subs/internal/application/dto"
	"github.com/agl/online_subs/internal/application/ports"
	"github.com/agl/online_subs/internal/errormsgs"
	"github.com/agl/online_subs/pkg/logger"
)

type SubsController struct {
	service ports.SubscriptionService
	port    string
}

func NewSubsController(service ports.SubscriptionService) *SubsController {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &SubsController{
		service: service,
		port:    port,
	}
}

// @title Online Subscriptions API
// @version 1.0
// @description REST API for managing user online subscriptions
// @host localhost:8080
// @BasePath /
func (sc *SubsController) StartServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /subscriptions", sc.CreateSubscription)
	mux.HandleFunc("GET /subscriptions/{userUUID}", sc.GetSubscriptionByUserUUID)
	mux.HandleFunc("POST /subscriptions/filter", sc.GetSubscriptionFiltered)
	mux.HandleFunc("PUT /subscriptions/{userUUID}", sc.UpdateSubscriptionByUserUUID)
	mux.HandleFunc("DELETE /subscriptions/{userUUID}", sc.DeleteSubscriptionByUserUUID)
	mux.HandleFunc("POST /subscriptions/sum", sc.SumSubscriptions)

	if err := http.ListenAndServe(":"+sc.port, mux); err != nil {
		logger.Log.Error("Failed to start server", "error", err)
	}
}

// @Summary Create subscription
// @Description Create a new subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body dto.Subscription true "Subscription info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions [post]
func (sc *SubsController) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var subDto dto.Subscription
	if err := json.NewDecoder(r.Body).Decode(&subDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	if err := sc.service.CreateSubscription(subDto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "created"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Get sum of subscriptions
// @Description Get total price of subscriptions for user/service/period
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param filter body dto.SumSubscriptionsRequest true "Filter parameters"
// @Success 200 {object} dto.SumSubscriptionsResponse
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions/sum [post]
func (sc *SubsController) SumSubscriptions(w http.ResponseWriter, r *http.Request) {
	var req dto.SumSubscriptionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	total, err := sc.service.SumSubscriptions(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.SumSubscriptionsResponse{Total: total}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Get subscription by user UUID
// @Description Get subscription for a user by UUID
// @Tags subscriptions
// @Produce json
// @Param userUUID path string true "User UUID"
// @Success 200 {object} dto.Subscription
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Subscription not found"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions/{userUUID} [get]
func (sc *SubsController) GetSubscriptionByUserUUID(w http.ResponseWriter, r *http.Request) {
	userUUID := r.PathValue("userUUID")
	if userUUID == "" {
		http.Error(w, "User UUID is required", http.StatusBadRequest)
		return
	}

	sub, err := sc.service.GetSubscriptionByUserUUID(userUUID)
	if errormsgs.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusNotFound)
		
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary List subscriptions by filter
// @Description Get subscriptions matching filter criteria
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param filter body dto.Subscription true "Filter parameters"
// @Success 200 {array} dto.Subscription
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 404 {object} map[string]string "No subscriptions found"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions/filter [post]
func (sc *SubsController) GetSubscriptionFiltered(w http.ResponseWriter, r *http.Request) {
	var subDto dto.Subscription
	if err := json.NewDecoder(r.Body).Decode(&subDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	subscriptions, err := sc.service.GetSubscriptionFiltered(subDto)
	if errormsgs.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(subscriptions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

// @Summary Update subscription by user UUID
// @Description Update an existing subscription by user UUID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param userUUID path string true "User UUID"
// @Param subscription body dto.UpdateSubscription true "Update data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions/{userUUID} [put]
func (sc *SubsController) UpdateSubscriptionByUserUUID(w http.ResponseWriter, r *http.Request) {
	userUUID := r.PathValue("userUUID")
	if userUUID == "" {
		http.Error(w, "User UUID is required", http.StatusBadRequest)
		return
	}

	var subDto dto.UpdateSubscription
	if err := json.NewDecoder(r.Body).Decode(&subDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := sc.service.UpdateSubscriptionByUserUUID(subDto, userUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Delete subscription by user UUID
// @Description Delete an existing subscription by user UUID
// @Tags subscriptions
// @Produce json
// @Param userUUID path string true "User UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string "Invalid UUID"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /subscriptions/{userUUID} [delete]
func (sc *SubsController) DeleteSubscriptionByUserUUID(w http.ResponseWriter, r *http.Request) {
	userUUID := r.PathValue("userUUID")
	if userUUID == "" {
		http.Error(w, "User UUID is required", http.StatusBadRequest)
		return
	}

	if err := sc.service.DeleteSubscriptionByUserUUID(userUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
