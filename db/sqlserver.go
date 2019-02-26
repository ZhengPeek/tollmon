package db

import (
	_ "github.com/denisenkom/go-mssqldb"
	"database/sql"
	"tollsys/tollmon/g"
	"tollsys/tollmon/datastruct"
	"strings"
	"strconv"
	"tollsys/tookit/database"
	"errors"
)
//Client sqlserver客户端实例
var Client database.ISqlClient

func init() {
	Client = database.GetSQLInstance()
}
//InitDB 获取Client实例接口，单例
func InitDB() {
	err := Client.Create(g.Config().DB.Host, g.Config().DB.DbName, g.Config().DB.User, g.Config().DB.Pwd)
	if err != nil {
		g.LogError(err.Error())
	}
	g.LogInfo("Init SQL Server DB OK")
}
//QueryResultNodeList实现QueryResult接口，该结构返回可Node列表
type QueryResultNodeList struct {
	nodeList []datastruct.Node
}
func (q *QueryResultNodeList)GC(){
	q.nodeList = nil
}
//GetNodeList 由QueryResultNodeLis实例调用，获取NodeList
func (q *QueryResultNodeList) GetNodeList()([]datastruct.Node,error){
	if len(q.nodeList) == 0{
		g.LogError("NodeList is blank")
		return nil,errors.New("NodeList is blank")
	}
	return q.nodeList, nil
}
//OnResult由QueryResultNodeLis实例调用，实现tollsys/toolkit/database/IQueryResult接口
func (q *QueryResultNodeList) OnResult(r *sql.Rows)(err error) {
	var node datastruct.Node
	for r.Next() {
		var nodeIp []byte
		err = r.Scan(&node.NodeID, &node.NodeName, &nodeIp,&node.TranMode)
		if err != nil {
			g.LogError("sql convert err:", err.Error())
		}
		var t []string
		for _, b := range nodeIp {
			i := int(b)
			t = append(t, strconv.Itoa(i))
		}
		node.NodeType,_ = strconv.Atoi(node.NodeID[len(node.NodeID)-1:len(node.NodeID)])
		node.NodeIP = strings.Join(t, ".")
		t = nil
		q.nodeList = append(q.nodeList, node)
	}
	return err
}
