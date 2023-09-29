package wire

import "time"

// AlgorithmHashSlots algorithm in routing
const (
	AlgorithmHashSlots = "hashslots"
)

// Command defined data type between client and server
const (
	// login
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	// chat
	CommandChatUserTalk  = "chat.user.talk"
	CommandChatGroupTalk = "chat.group.talk"
	CommandChatTalkAck   = "chat.talk.ack"

	// 离线
	CommandOfflineIndex   = "chat.offline.index"
	CommandOfflineContent = "chat.offline.content"

	// 群管理
	CommandGroupCreate  = "chat.group.create"
	CommandGroupJoin    = "chat.group.join"
	CommandGroupQuit    = "chat.group.quit"
	CommandGroupMembers = "chat.group.members"
	CommandGroupDetail  = "chat.group.detail"
)

const (
	// MetaDestServer 消息将要送达的网关的ServiceName
	MetaDestServer = "dest.server"

	// MetaDestChannels 消息将要送达的channels
	// 消息接收方，这是一个列表，也就是一个消息可以推送给多个用户
	// 由于没有设置多设备登录，因此一个用户就是一个channel
	MetaDestChannels = "dest.channels"
)

type Protocol string

const (
	ProtocolTCP       Protocol = "tcp"
	ProtocolWebsocket Protocol = "websocket"
)

// SNWGateway 定义统一的服务名
const (
	SNWGateway = "wgateway"
	SNTGateway = "tgateway"
	SNLogin    = "chat"
	SNChat     = "chat"
	SNService  = "service"
)

type ServiceID string

type SessionID string

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)

const (
	OfflineMessageExpiresIn = time.Hour * 24 * 30
	OfflineSyncIndexCount   = 3000
	OfflineMessageStoreDays = 30 //days
)

const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeVoice = 3
	MessageTypeVideo = 4
)
