package ports

import (
	"time"

	"github.com/agl/online_subs/internal/domain/entities"
)

type SubscriptionRepo interface {
	CreateSubscription(subscription entities.Subscription) error
	GetSubscriptionByUserUUID(userUUID string) (entities.Subscription, error)
	GetSubscriptionFiltered(subscription entities.Subscription) ([]entities.Subscription, error)
	UpdateSubscriptionByUserUUID(subscription entities.Subscription) error
	DeleteSubscriptionByUserUUID(userUUID string) error
	SumSubscriptions(userID, serviceName string, startPeriod *time.Time, endPeriod *time.Time) (int, error)
}
