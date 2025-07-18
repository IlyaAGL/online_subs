package service

import (
	"errors"
	"time"

	"github.com/agl/online_subs/internal/application/dto"
	"github.com/agl/online_subs/internal/application/ports"
	"github.com/agl/online_subs/internal/domain/entities"
	"github.com/agl/online_subs/pkg/logger"
)

type SubscriptionService struct {
	repo ports.SubscriptionRepo
}

func NewSubsService(repo ports.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) CreateSubscription(subDto dto.Subscription) error {
	logger.Log.Info("CreateSubscription called", "user_id", subDto.UserID, "service_name", subDto.ServiceName)

	startDateParsed, err := time.Parse("01-2006", subDto.StartDate)
	if err != nil {
		logger.Log.Error("Failed to parse start date", "error", err)

		return err
	}

	subEntity := entities.Subscription{
		ServiceName: subDto.ServiceName,
		Price:       subDto.Price,
		UserID:      subDto.UserID,
		StartDate:   startDateParsed,
	}

	if subDto.EndDate != "" {
		endDateParsed, err := time.Parse("01-2006", subDto.EndDate)
		if err != nil {
			logger.Log.Error("Failed to parse end date", "error", err)

			return err
		}

		if endDateParsed.Before(startDateParsed) {
			logger.Log.Error("End date cannot be before start date", "start_date", subDto.StartDate, "end_date", subDto.EndDate)

			return errors.New("invalid data: end date cannot be before start date")
		}

		subEntity.EndDate = &endDateParsed
	}

	err = s.repo.CreateSubscription(subEntity)
	if err != nil {
		logger.Log.Error("Failed to create subscription", "error", err)

		return err
	}

	logger.Log.Info("Subscription created successfully", "user_id", subDto.UserID, "service_name", subDto.ServiceName)

	return nil
}

func (s *SubscriptionService) GetSubscriptionByUserUUID(userUUID string) (dto.Subscription, error) {
	logger.Log.Info("GetSubscriptionByUserUUID called", "user_id", userUUID)

	subEntity, err := s.repo.GetSubscriptionByUserUUID(userUUID)
	if err != nil {
		logger.Log.Error("Failed to get subscription", "error", err)
		return dto.Subscription{}, err
	}

	logger.Log.Info("Mapping entity to DTO", "service_name", subEntity.ServiceName, "price", subEntity.Price)

	startDateFormatted := subEntity.StartDate.Format("01-2006")

	subDTO := dto.Subscription{
		ServiceName: subEntity.ServiceName,
		Price:       subEntity.Price,
		UserID:      subEntity.UserID,
		StartDate:   startDateFormatted,
	}

	if subEntity.EndDate != nil {
		endDateFormatted := subEntity.EndDate.Format("01-2006")
		subDTO.EndDate = endDateFormatted
	}

	logger.Log.Info("Subscription fetched successfully", "user_id", userUUID, "service_name", subDTO.ServiceName)

	return subDTO, nil
}

func (s *SubscriptionService) GetSubscriptionFiltered(subDTO dto.Subscription) ([]dto.Subscription, error) {
	logger.Log.Info("GetSubscriptionFiltered called", "user_id", subDTO.UserID, "service_name", subDTO.ServiceName)

	var startDateParsed, endDateParsed time.Time

	if subDTO.StartDate != "" {
		logger.Log.Info("Parsing start date", "start_date", subDTO.StartDate)

		parsed, err := time.Parse("01-2006", subDTO.StartDate)
		if err != nil {
			logger.Log.Error("Failed to parse start date", "error", err)
			return nil, err
		}

		startDateParsed = parsed
	}

	if subDTO.EndDate != "" {
		logger.Log.Info("Parsing end date", "end_date", subDTO.EndDate)

		parsed, err := time.Parse("01-2006", subDTO.EndDate)
		if err != nil {
			logger.Log.Error("Failed to parse end date", "error", err)
			return nil, err
		}

		endDateParsed = parsed
	}

	if !startDateParsed.IsZero() && !endDateParsed.IsZero() && endDateParsed.Before(startDateParsed) {
		logger.Log.Error("End date cannot be before start date", "start_date", subDTO.StartDate, "end_date", subDTO.EndDate)

		return nil, errors.New("invalid data: end date cannot be before start date")
	}

	logger.Log.Info("Building filter entity", "user_id", subDTO.UserID, "service_name", subDTO.ServiceName, "price", subDTO.Price)
	
	subEntity := entities.Subscription{
		UserID:      subDTO.UserID,
		Price:       subDTO.Price,
		ServiceName: subDTO.ServiceName,
		StartDate:   startDateParsed,
		EndDate:     &endDateParsed,
	}

	subscriptions, err := s.repo.GetSubscriptionFiltered(subEntity)
	if err != nil {
		logger.Log.Error("Failed to get filtered subscriptions", "error", err)
		return nil, err
	}

	logger.Log.Info("Mapping filtered subscriptions to DTO", "count", len(subscriptions))

	result := make([]dto.Subscription, 0)

	for _, sub := range subscriptions {
		startDateFormatted := sub.StartDate.Format("01-2006")
		endDateFormatted := ""
		if sub.EndDate != nil {
			endDateFormatted = sub.EndDate.Format("01-2006")
		}

		result = append(result, dto.Subscription{
			ServiceName: sub.ServiceName,
			Price:       sub.Price,
			UserID:      sub.UserID,
			StartDate:   startDateFormatted,
			EndDate:     endDateFormatted,
		})
	}

	logger.Log.Info("Filtered subscriptions fetched successfully", "result_count", len(result))

	return result, nil
}

func (s *SubscriptionService) UpdateSubscriptionByUserUUID(subDTO dto.UpdateSubscription, userUUID string) error {
	logger.Log.Info("UpdateSubscriptionByUserUUID called", "user_id", userUUID)

	subEntity := entities.Subscription{
		ServiceName: subDTO.ServiceName,
		Price:       subDTO.Price,
		UserID:      userUUID,
	}

	if subDTO.StartDate != "" {
		startDateParsed, err := time.Parse("01-2006", subDTO.StartDate)
		if err != nil {
			logger.Log.Error("Failed to parse start date", "error", err)

			return err
		}

		subEntity.StartDate = startDateParsed
	}

	if subDTO.EndDate != "" {
		endDateParsed, err := time.Parse("01-2006", subDTO.EndDate)
		if err != nil {
			logger.Log.Error("Failed to parse end date", "error", err)

			return err
		}

		subEntity.EndDate = &endDateParsed
	}

	if subEntity.EndDate != nil && subEntity.StartDate.After(*subEntity.EndDate) {
		logger.Log.Error("End date cannot be before start date", "start_date", subDTO.StartDate, "end_date", subDTO.EndDate)

		return errors.New("invalid data: end date cannot be before start date")
	}

	err := s.repo.UpdateSubscriptionByUserUUID(subEntity)
	if err != nil {
		logger.Log.Error("Failed to update subscription", "error", err)

		return err
	}

	logger.Log.Info("Subscription updated successfully", "user_id", userUUID)

	return nil
}

func (s *SubscriptionService) DeleteSubscriptionByUserUUID(userUUID string) error {
	logger.Log.Info("DeleteSubscriptionByUserUUID called", "user_id", userUUID)

	err := s.repo.DeleteSubscriptionByUserUUID(userUUID)
	if err != nil {
		logger.Log.Error("Failed to delete subscription", "error", err)

		return err
	}

	logger.Log.Info("Subscription deleted successfully", "user_id", userUUID)

	return nil
}

func (s *SubscriptionService) SumSubscriptions(req dto.SumSubscriptionsRequest) (int, error) {
	logger.Log.Info("SumSubscriptions called", "user_id", req.UserID, "service_name", req.ServiceName)

	var startDate, endDate *time.Time

	if req.StartPeriod != "" {
		t, err := time.Parse("01-2006", req.StartPeriod)
		if err != nil {
			logger.Log.Error("Failed to parse start period", "error", err)

			return 0, err
		}

		startDate = &t
	}

	if req.EndPeriod != "" {
		t, err := time.Parse("01-2006", req.EndPeriod)
		if err != nil {
			logger.Log.Error("Failed to parse end period", "error", err)

			return 0, err
		}
		endDate = &t
	}

	total, err := s.repo.SumSubscriptions(req.UserID, req.ServiceName, startDate, endDate)
	if err != nil {
		logger.Log.Error("Failed to sum subscriptions", "error", err)

		return 0, err
	}

	logger.Log.Info("SumSubscriptions completed", "user_id", req.UserID, "service_name", req.ServiceName, "total", total)

	return total, nil
}
