package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/agl/online_subs/internal/domain/entities"
	"github.com/agl/online_subs/internal/errormsgs"
	"github.com/agl/online_subs/pkg/logger"
)

type SubsRepo struct {
	db      *sql.DB
	builder squirrel.StatementBuilderType
}

func NewSubsRepo(db *sql.DB) *SubsRepo {
	return &SubsRepo{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (sr *SubsRepo) CreateSubscription(sub entities.Subscription) error {
	logger.Log.Info("Repo: CreateSubscription called", "user_id", sub.UserID, "service_name", sub.ServiceName)

	query, args, err := sr.builder.
		Insert("Subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		ToSql()

	if err != nil {
		logger.Log.Error("Repo: Failed to build insert query", "error", err)

		return fmt.Errorf("failed to build query: %w", err)
	}

	tx, err := sr.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to start transaction", "error", err)

		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Log.Error("Repo: Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	_, err = tx.Exec(query, args...)
	if err != nil {
		logger.Log.Error("Repo: Failed to execute insert", "error", err)

		return fmt.Errorf("failed to create subscription: %w", err)
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Error("Repo: Failed to commit transaction", "error", err)

		return err
	}

	logger.Log.Info("Repo: Subscription created successfully", "user_id", sub.UserID, "service_name", sub.ServiceName)

	return nil
}

func (sr *SubsRepo) GetSubscriptionByUserUUID(userUUID string) (entities.Subscription, error) {
	logger.Log.Info("Repo: GetSubscriptionByUserUUID called", "user_id", userUUID)

	query, args, err := sr.builder.
		Select("service_name", "price", "user_id", "start_date", "end_date").
		From("Subscriptions").
		Where("user_id = ?", userUUID).
		ToSql()

	if err != nil {
		logger.Log.Error("Repo: Failed to build select query", "error", err)

		return entities.Subscription{}, fmt.Errorf("couldn't make the query: %w", err)
	}

	logger.Log.Info("Repo: Executing query for GetSubscriptionByUserUUID", "query", query, "args", args)

	row := sr.db.QueryRow(query, args...)

	var sub entities.Subscription
	err = row.Scan(&sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("Repo: Subscription not found", "user_id", userUUID)

			return entities.Subscription{}, errormsgs.NotFound
		}

		logger.Log.Error("Repo: Failed to scan subscription", "error", err)

		return entities.Subscription{}, fmt.Errorf("couldn't extract the entity: %w", err)
	}

	logger.Log.Info("Repo: Subscription fetched successfully", "user_id", userUUID, "service_name", sub.ServiceName)

	return sub, nil
}

func (sr *SubsRepo) GetSubscriptionFiltered(subscription entities.Subscription) ([]entities.Subscription, error) {
	logger.Log.Info("Repo: GetSubscriptionFiltered called", "user_id", subscription.UserID)

	logger.Log.Info("Repo: Building query for GetSubscriptionFiltered", "user_id", subscription.UserID, "service_name", subscription.ServiceName, "price", subscription.Price)

	builder := sr.builder.Select("service_name", "price", "user_id", "start_date", "end_date").From("Subscriptions")

	if subscription.UserID != "" {
		logger.Log.Info("Repo: Filtering by user_id", "user_id", subscription.UserID)

		builder = builder.Where("user_id = ?", subscription.UserID)
	}

	if subscription.Price != 0 {
		logger.Log.Info("Repo: Filtering by price >=", "price", subscription.Price)

		builder = builder.Where("price >= ?", subscription.Price)
	}

	if subscription.ServiceName != "" {
		logger.Log.Info("Repo: Filtering by service_name", "service_name", subscription.ServiceName)

		builder = builder.Where("service_name = ?", subscription.ServiceName)
	}

	if !subscription.StartDate.IsZero() {
		logger.Log.Info("Repo: Filtering by start_date >=", "start_date", subscription.StartDate)

		builder = builder.Where("start_date >= ?", subscription.StartDate)
	}

	if !subscription.EndDate.IsZero() {
		logger.Log.Info("Repo: Filtering by end_date <=", "end_date", subscription.EndDate)

		builder = builder.Where("end_date <= ?", subscription.EndDate)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Log.Error("Repo: Failed to build select query", "error", err)

		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	logger.Log.Info("Repo: Executing filtered select query", "query", query, "args", args)

	rows, err := sr.db.Query(query, args...)
	if err != nil {
		logger.Log.Error("Repo: Failed to execute select query", "error", err)

		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	defer rows.Close()

	subscriptions := make([]entities.Subscription, 0)

	for rows.Next() {
		var sub entities.Subscription
		err := rows.Scan(&sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			logger.Log.Error("Repo: Failed to scan row", "error", err)

			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		logger.Log.Info("Repo: Row scanned", "user_id", sub.UserID, "service_name", sub.ServiceName)

		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error("Repo: Error in rows iteration", "error", err)

		return nil, fmt.Errorf("error in rows iteration: %w", err)
	}

	if len(subscriptions) == 0 {
		logger.Log.Error("Repo: Subscriptions not found", "filter", subscription)

		return nil, errormsgs.NotFound
	}

	logger.Log.Info("Repo: Subscriptions fetched successfully", "count", len(subscriptions))

	return subscriptions, nil
}

func (sr *SubsRepo) UpdateSubscriptionByUserUUID(subscription entities.Subscription) error {
	logger.Log.Info("Repo: UpdateSubscriptionByUserUUID called", "user_id", subscription.UserID)

	builder := sr.builder.Update("Subscriptions")

	fieldsToUpdate := false

	if subscription.ServiceName != "" {
		builder = builder.Set("service_name", subscription.ServiceName)
		fieldsToUpdate = true
	}

	if subscription.Price != 0 {
		builder = builder.Set("price", subscription.Price)
		fieldsToUpdate = true
	}

	if !subscription.StartDate.IsZero() {
		builder = builder.Set("start_date", subscription.StartDate)
		fieldsToUpdate = true
	}

	if !subscription.EndDate.IsZero() {
		builder = builder.Set("end_date", *subscription.EndDate)
		fieldsToUpdate = true
	}

	if !fieldsToUpdate {
		logger.Log.Info("Repo: No fields to update for user", "user_id", subscription.UserID)

		return nil
	}

	builder = builder.Where("user_id = ?", subscription.UserID)
	query, args, err := builder.ToSql()
	if err != nil {
		logger.Log.Error("Repo: Failed to build update query", "error", err)

		return fmt.Errorf("failed to build update query: %w", err)
	}

	tx, err := sr.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to start transaction", "error", err)

		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Log.Error("Repo: Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	result, err := tx.Exec(query, args...)
	if err != nil {
		logger.Log.Error("Repo: Failed to execute update", "error", err)

		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Repo: Failed to get affected rows", "error", err)
		
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		logger.Log.Error("Repo: No subscription found for update", "user_id", subscription.UserID)

		return fmt.Errorf("no subscription found for user %s", subscription.UserID)
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Error("Repo: Failed to commit transaction", "error", err)

		return err
	}

	logger.Log.Info("Repo: Subscription updated successfully", "user_id", subscription.UserID)
	
	return nil
}

func (sr *SubsRepo) DeleteSubscriptionByUserUUID(userUUID string) error {
	logger.Log.Info("Repo: DeleteSubscriptionByUserUUID called", "user_id", userUUID)

	query, args, err := sr.builder.
		Delete("Subscriptions").
		Where("user_id = ?", userUUID).
		ToSql()

	if err != nil {
		logger.Log.Error("Repo: Failed to build delete query", "error", err)

		return fmt.Errorf("failed to build delete query: %w", err)
	}

	tx, err := sr.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to start transaction", "error", err)

		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Log.Error("Repo: Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	result, err := tx.Exec(query, args...)
	if err != nil {
		logger.Log.Error("Repo: Failed to execute delete", "error", err)

		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Repo: Failed to get affected rows", "error", err)

		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		logger.Log.Error("Repo: No subscription found for delete", "user_id", userUUID)

		return fmt.Errorf("no subscription found for user %s", userUUID)
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Error("Repo: Failed to commit transaction", "error", err)

		return err
	}

	logger.Log.Info("Repo: Subscription deleted successfully", "user_id", userUUID)

	return nil
}

func (sr *SubsRepo) SumSubscriptions(userID, serviceName string, startPeriod, endPeriod *time.Time) (int, error) {
	logger.Log.Info("Repo: SumSubscriptions called", "user_id", userID, "service_name", serviceName)

	builder := sr.builder.Select("COALESCE(SUM(price), 0)").From("Subscriptions")

	if userID != "" {
		builder = builder.Where("user_id = ?", userID)
	}

	if serviceName != "" {
		builder = builder.Where("service_name = ?", serviceName)
	}

	if startPeriod != nil {
		builder = builder.Where("start_date >= ?", *startPeriod)
	}

	if endPeriod != nil {
		builder = builder.Where("end_date <= ?", *endPeriod)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Log.Error("Repo: Failed to build sum query", "error", err)

		return 0, fmt.Errorf("failed to build sum query: %w", err)
	}

	var sum int

	err = sr.db.QueryRow(query, args...).Scan(&sum)
	if err != nil {
		logger.Log.Error("Repo: Failed to execute sum query", "error", err)

		return 0, fmt.Errorf("failed to execute sum query: %w", err)
	}

	logger.Log.Info("Repo: SumSubscriptions completed", "user_id", userID, "service_name", serviceName, "total", sum)

	return sum, nil
}
