package interfaces

import "user-ws-api/models"

type OrderSubmitter interface {
	Submit(order models.Order)
}
