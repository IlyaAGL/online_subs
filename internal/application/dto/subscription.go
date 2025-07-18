package dto

type Subscription struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}
type SumSubscriptionsRequest struct {
	UserID      string `json:"user_id"`
	ServiceName string `json:"service_name"`
	StartPeriod string `json:"start_date"`
	EndPeriod   string `json:"end_date"`
}

type SumSubscriptionsResponse struct {
	Total int `json:"total"`
}
