package ports

import "github.com/agl/online_subs/internal/application/dto"

type SubscriptionService interface {
	CreateSubscription(subscripption dto.Subscription) error
	GetSubscriptionByUserUUID(userUUID string) (dto.Subscription, error)
	GetSubscriptionFiltered(subscription dto.Subscription) ([]dto.Subscription, error)
	UpdateSubscriptionByUserUUID(subscripption dto.UpdateSubscription, userUUID string) error
	DeleteSubscriptionByUserUUID(userUUID string) error
	SumSubscriptions(req dto.SumSubscriptionsRequest) (int, error)
}
