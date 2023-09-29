package handler

import (
	"github.com/go-redis/redis/v7"
	"github.com/sjmshsh/HopeIM/services/service/database"
	"gorm.io/gorm"
)

type ServiceHandler struct {
	BaseDb    *gorm.DB
	MessageDb *gorm.DB
	Cache     *redis.Client
	Idgen     *database.IDGenerator
}
