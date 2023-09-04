package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zinx/zinx/utils"
	"zinx/zinx/ziface"
)

// 封包  拆包的具体模块
type DataPack struct{}

// 拆包封包实例的初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头的长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//Datalen uint32 + ID uint32  8字节
	return 8
}

// 封包方法
// |data|msgID|data|
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放byte字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//将dataLen写进databuff中  大端、小端格式
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}
	//将MsgID写进databuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}
	//将Data写进databuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil

}

// 拆包方法(将包的head信息读出来  之后根据信息的长度在进行一次毒)
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	//创建一个从输入二进制数据的ioreader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head信息，得到dataLen和MsgID
	msg := &Message{}

	//读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	//读MsgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	//判断datalen允许的最大包长度
	if utils.GlobalObject.MaxPackageSize > 0 && msg.DataLen > utils.GlobalObject.MaxPackageSize {
		return nil, errors.New("too large msg data")
	}
	return msg, nil
}
