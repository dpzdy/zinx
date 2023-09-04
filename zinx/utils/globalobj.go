package utils

import (
	"encoding/json"
	"os"
	"zinx/zinx/ziface"
)

/*
存储一切有关zinx框架的全局参数，供其他模块使用
一些参数是可以通过zinx.json有用户进行配置
*/

type GlobalObj struct {
	/*
		server
	*/
	TcpServer ziface.IServer //全局的server对象
	Host      string         //服务器主机监听的IP
	TcpPort   int            //服务器主机监听的端口号
	Name      string         //当前服务名称

	/*
		Zinx
	*/
	Version          string //zinx的版本号
	MaxConn          int    //最大连接数
	MaxPackageSize   uint32 //数据包最大值
	WorkerPoolSize   uint32 //当前业务工作worker池的数量
	MaxWorkerTaskLen uint32 //允许的最大worker池的数量
}

/*
定义一个全局的对外Globalobj
*/
var GlobalObject *GlobalObj

/*
从zinx.json去加载用于自定义的参数
*/
func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	//将json解析到struce中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

/*
提供init方法，初始化当前Clobalobj
*/
func init() {
	//如果配置问文件没有加载
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V0.4",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   10,   //worker工作池的个数
		MaxWorkerTaskLen: 1024, //每个worker对应的消息队列的任务的数量最大值
	}

	//尝试从conf/zinx.json去加载一些用户自定义的参数
	GlobalObject.Reload()
}
