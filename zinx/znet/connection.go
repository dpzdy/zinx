package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"zinx/zinx/utils"
	"zinx/zinx/ziface"
)

/*
链接模块
*/
type Connection struct {
	//当前Conn隶属与那个Server
	TCPServer ziface.IServer
	//当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接的ID
	ConnID uint32

	//当前的链接状态
	isClosed bool

	//告知当前链接已经退出/停止的channel(由Reader告知Writer退出)
	ExitChan chan bool

	//无缓冲的管道，用于读、写Goroutine之间的消息通信
	msgChan chan []byte

	//消息的管理MsgID,和对应的处理业务api
	MsgHandler ziface.IMsgHandle

	//链接属性集合
	property map[string]interface{}
	//保护链接属性的锁
	propertyLock sync.RWMutex
}

// 初始化链接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TCPServer:  server,
		Conn:       conn,
		ConnID:     connID,
		MsgHandler: msgHandler,
		isClosed:   false,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
		property:   make(map[string]interface{}),
	}
	//将Conn加入到ConnManger中
	c.TCPServer.GetConnMgr().Add(c)
	return c
}

// 读数据+封装request+router处理
func (c *Connection) StartReader() {
	fmt.Println("[Reader] Goroutine is running")
	defer fmt.Println("connID = ", c.ConnID, "[ Reader is exit],remote addr is ", c.RemoteAddr().String())
	defer c.Stop()

	for {
		//创建一个拆包解包对象
		dp := NewDataPack()
		//读取客户端的msg head 8个字节  二进制流
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("Read msg head error ", err)
			break
		}
		//拆包，得到msgid msgdatalen 放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("Unpack error ", err)
			break
		}
		//根据datalen 再次读取data 放在msg.data
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("Read msg data error ", err)
				break
			}
		}
		msg.SetData(data)
		//得到当前Conn数据的Request请求的数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		//从路由中，找到注册绑定的conn对应的router调用
		//根据绑定好的msgid 找到对应处理api业务的router调用
		if utils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			c.MsgHandler.DoMsgHandler(&req)
		}

		//从路由中，找到注册绑定的conn对应的router调用

	}
}

// 提供一个sendmsg方法  将发给客户端的数据  先封包 再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data进行封包 ，msgdatalen|msgid|data
	dp := NewDataPack()

	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack err msg id = ", msgId)
		return errors.New("Pack err msg")
	}
	//将数据发送给管道
	c.msgChan <- binaryMsg
	return nil
}

/*
写消息的Goroutine ,专门发送给客户端消息的模块
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer] Goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), " [conn Writer exit!]")

	//不断阻塞的等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data err ", err)
				return
			}
		case <-c.ExitChan:
			//代表reader已经退出，writer也要退出
			return

		}
	}
}

// 启动链接 让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start()..ConnID = ", c.ConnID)
	//启动从当前链接的度数据业务
	go c.StartReader()
	//启动从当前链接写数据的业务
	go c.StartWriter()
	//按照用户传递进来的  创建链接之后需要调用的处理业务，执行对应的hook
	c.TCPServer.CallOnConnStart(c)

}

// 停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()..ConnID = ", c.ConnID)
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	//调用开发者注册的  销毁链接之前  需要执行的函数
	c.TCPServer.CallOnConnStop(c)
	//关闭socket链接
	c.Conn.Close()
	//告知writer关闭
	c.ExitChan <- true
	//将当前链接从ConnMgr中摘除掉
	c.TCPServer.GetConnMgr().Remove(c)
	//回收资源
	close(c.ExitChan)
	close(c.msgChan)

}

// 获取当前链接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn

}

// 获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID

}

// 获取远程客户端的TCP状态 IP port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()

}

// 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("No peoperty FOUND")
	}
}

// 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
func matchMaxLengthSubString(input string) []string {
	// write code here
	lastOccurres := make(map[byte]int)
	start := 0
	maxLength := 0
	res := []string{}
	for i, ch := range []byte(input) {
		if lastI, ok := lastOccurres[ch]; ok && lastI >= start {
			start = lastI + i
		}
		if i-start+1 > maxLength {
			maxLength = i - start + 1
			res = []string{input[start : i+1]}
		} else if i-start+1 == maxLength {
			res = append(res, input[start:i+1])
		}
		lastOccurres[ch] = i
	}
	return res
}
func twoSum(nums []int, target int) [][]int {
	res := [][]int{}
	sort.Ints(nums)
	lo, hi := 0, len(nums)-1
	for lo < hi {
		asum := nums[lo] + nums[hi]
		left, right := nums[lo], nums[hi]
		if asum < target {
			for lo < hi && nums[lo] == left {
				lo++
			}
		} else if asum > target {
			for lo < hi && nums[hi] == right {
				hi--
			}
		} else {
			res = append(res, []int{left, right})
			for lo < hi && nums[lo] == left {
				lo++
			}
			for lo < hi && nums[hi] == right {
				hi--
			}
		}
	}
	//fmt.Println(res)
	return res
}

func threeSumt(nums []int, target int) [][]int {
	res := [][]int{}
	sort.Ints(nums)
	for i := 0; i < len(nums)-2; i++ {
		curList := twoSum(nums[i+1:], target-nums[i])
		if curList != nil {
			for _, arr := range curList {
				temp := []int{nums[i]}
				temp = append(temp, arr...)
				flag := true
				for j := 0; j < len(res); j++ {
					ans := 0
					for k := 0; k < len(res[j]); k++ {
						if temp[k] == res[j][k] {
							ans++
						}
					}
					if ans == len(res[j]) {
						flag = false
						break
					}
				}
				if flag {
					res = append(res, temp)
				}

			}
		}
	}
	return res
}
