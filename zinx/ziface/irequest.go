package ziface

/*
IRquest接口：把客户端请求的链接数据和请求的数据 保证到一个Request中
*/
type IRequest interface {
	//得到当前链接
	GetConnection() IConnection
	//得到请求的消息数据
	GetData() []byte

	GetMsgId() uint32
}
