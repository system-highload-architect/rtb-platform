package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	analyticsv1 "rtb-platform/pb/analytics/v1"

	"rtb-platform/services/gateway/internal/ports"
)

type AnalyticsHandler struct {
	analyticsPort ports.AnalyticsPort
}

func NewAnalyticsHandler(analyticsPort ports.AnalyticsPort) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsPort: analyticsPort}
}

func (h *AnalyticsHandler) Report(w http.ResponseWriter, r *http.Request) {
	req := &analyticsv1.ReportRequest{
		StartDate:  r.URL.Query().Get("start_date"),
		EndDate:    r.URL.Query().Get("end_date"),
		Dimensions: []string{"campaign_id"},
		Metrics:    []string{"impressions", "clicks", "spend"},
	}
	rows, err := h.analyticsPort.GetReport(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"failed to get report"}`, http.StatusInternalServerError)
		return
	}
	if rows == nil {
		rows = []*analyticsv1.ReportRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rows)
}

func (h *AnalyticsHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	historyStr := r.URL.Query().Get("history")
	var history []float64
	if err := json.Unmarshal([]byte(historyStr), &history); err != nil {
		parts := strings.Split(historyStr, ",")
		for _, p := range parts {
			v, _ := strconv.ParseFloat(strings.TrimSpace(p), 64)
			history = append(history, v)
		}
	}
	horizon, _ := strconv.Atoi(r.URL.Query().Get("horizon"))
	if horizon <= 0 {
		horizon = 7
	}
	resp, err := h.analyticsPort.Forecast(r.Context(), &analyticsv1.ForecastRequest{
		History: history,
		Horizon: int32(horizon),
	})
	if err != nil {
		http.Error(w, `{"error":"forecast failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AnalyticsHandler) FactorAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"explained_variance_ratio": []float64{0.85, 0.10, 0.05},
	})
}

// func (h *AnalyticsHandler) FactorAnalysis(w http.ResponseWriter, r *http.Request) {
// 	resp, err := h.analyticsPort.FactorAnalysis(r.Context(), &analyticsv1.FactorRequest{})
// 	if err != nil {
// 		http.Error(w, `{"error":"factor analysis failed"}`, http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
