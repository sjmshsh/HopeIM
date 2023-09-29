package storage

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/wire/pkt"
	"time"
)

const (
	LocationExpired = time.Hour * 48
)

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) HopeIM.SessionStorage {
	return &RedisStorage{
		cli: cli,
	}
}

func (r *RedisStorage) Add(sesssion *pkt.Session) error {
	// save Hope.Location
	loc := HopeIM.Location{
		ChannelId: sesssion.ChannelId,
		GateId:    sesssion.GateId,
	}
	locKey := KeyLocation(sesssion.Account, "")
	err := r.cli.Set(locKey, loc.Bytes(), LocationExpired).Err()
	if err != nil {
		return err
	}

	// save session
	snKey := KeySession(sesssion.ChannelId)
	buf, _ := proto.Marshal(sesssion)
	err = r.cli.Set(snKey, buf, LocationExpired).Err()
	if err != nil {
		return err
	}
	return nil
}

// Delete a session
func (r *RedisStorage) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	err := r.cli.Del(locKey).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(channelId)
	err = r.cli.Del(snKey).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get get session by
// channelId == sessionId
func (r *RedisStorage) Get(ChannelId string) (*pkt.Session, error) {
	snKey := KeySession(ChannelId)
	bts, err := r.cli.Get(snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, HopeIM.ErrSessionNil
		}
		return nil, err
	}
	var session pkt.Session
	_ = proto.Unmarshal(bts, &session)
	return &session, err
}

func (r *RedisStorage) GetLocations(accounts ...string) ([]*HopeIM.Location, error) {
	keys := KeyLocations(accounts...)
	list, err := r.cli.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}
	var result = make([]*HopeIM.Location, 0)
	for _, l := range list {
		if l == nil {
			continue
		}
		var loc HopeIM.Location
		_ = loc.Unmarshal([]byte(l.(string)))
		result = append(result, &loc)
	}
	if len(result) == 0 {
		return nil, HopeIM.ErrSessionNil
	}
	return result, nil
}

func (r *RedisStorage) GetLocation(account string, device string) (*HopeIM.Location, error) {
	key := KeyLocation(account, device)
	bts, err := r.cli.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, HopeIM.ErrSessionNil
		}
		return nil, err
	}
	var loc HopeIM.Location
	_ = loc.Unmarshal(bts)
	return &loc, nil
}

func KeySession(channel string) string {
	return fmt.Sprintf("login:sn:%s", channel)
}

func KeyLocation(account, device string) string {
	if device == "" {
		return fmt.Sprintf("login:loc:%s", account)
	}
	return fmt.Sprintf("login:loc:%s:%s", account, device)
}

func KeyLocations(accounts ...string) []string {
	arr := make([]string, len(accounts))
	for i, account := range accounts {
		arr[i] = KeyLocation(account, "")
	}
	return arr
}
