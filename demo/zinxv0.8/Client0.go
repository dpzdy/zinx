package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/zinx/znet"
)

/*
模拟客户端
*/
func main() {

	fmt.Println("client0 start")
	time.Sleep(time.Second)

	//链接远程服务器，得到一个conn链接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err,exit!")
		return
	}

	for {
		//发送封包的msg
		dp := znet.NewDataPack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(0, []byte("Zinx client0 Test Message")))
		if err != nil {
			fmt.Println("Pack error ", err)
		}
		if _, err := conn.Write(binaryMsg); err != nil {
			fmt.Println("write error ", err)
			return
		}

		//服务器勇敢回复一个msg数据
		//读取流中的head部分  得到id 和 datalen
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head err ", err)
			break
		}
		//将二进制的head拆包到nsg中
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack error ", err)
			break
		}
		if msgHead.GetMsgLen() > 0 {
			//再根据Datelen进行第二次读取，将data读出
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data error ", err)
				return
			}
			fmt.Println("------> Recv Server Msg : Id =  ", msg.Id, " data = ", string(msg.GetData()))
		}

		//cpu阻塞
		time.Sleep(time.Second)
	}
}
