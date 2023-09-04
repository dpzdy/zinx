package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 只是负责测试datapack拆包封包
func TestDataPack(t *testing.T) {
	/*
		模拟的服务器
	*/
	//1创建socketTCP
	listenner, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server linten err ", err)
	}
	//创建一个go 负责从客户端处理业务
	go func() {
		//2从客户端读取数据 拆包处理.
		for {
			conn, err := listenner.Accept()
			if err != nil {
				fmt.Println("server accept error ", err)
				return
			}

			go func(conn net.Conn) {
				//处理客户端请求
				//拆包  两次读conn

				dp := NewDataPack()
				for {
					//第一次：把包的head读出来
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head err ", err)
						return
					}
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err ", err)
						break
					}
					if msgHead.GetMsgLen() > 0 {
						//msg 是有数据的 第二次读
						//第二次：根据head的datalen  在读取data的内容
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())

						//根据datalen的长度再次从io流中读取
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err ", err)
							return
						}

						//完整的一个消息读取完毕
						fmt.Println("----->Recv MsgId: ", msg.Id, ", datalen: ", msg.DataLen, " ,data: ", string(msg.Data))

					}

				}

			}(conn)
		}
	}()

	/*
		模拟客户端
	*/

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client dial err ", err)
		return
	}
	//创建一个封包对象dp
	dp := NewDataPack()
	//模拟粘包过程，封装两个msg一起发送
	//封装msg1包
	msg1 := &Message{
		Id:      1,
		DataLen: 4,
		Data:    []byte{'z', 'i', 'n', 'x'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err ", err)
		return
	}

	//封装msg2包
	msg2 := &Message{
		Id:      2,
		DataLen: 7,
		Data:    []byte{'n', 'i', 'h', 'a', 'o', 'z', 'y'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 err ", err)
		return
	}
	//将两个包粘在一起
	sendData1 = append(sendData1, sendData2...)
	//一次
	conn.Write(sendData1)

	//客户端zus
	select {}
}
