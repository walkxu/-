package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"strings"
)

/**
*	使用Wrap抛给上层必要原因：
*	1、防止获取数据信息的指针为空被业务层使用，引起空指针异常或其他错误；
*	2、业务层能够获取数据库查询错误的根本原因，可以将错误原因映射到用户层
*	使用Wrap抛给上层的其他原因：
*	1、记录sql查询语句的日志
*	2、保存错误根因
*	3、业务在调用第三库或标准库时，应该使用errors.Wrap 保存堆栈信息及错误详细详细，方便业务层问题定位
*	4、根据查询的结果进行业务处理，如果查询结果为零不进行相关业务处理，忽略或进行其他操作
 */

//数据库配置
const (
	userName = "root"
	password = "201142163"
	ip = "10.110.16.115"
	port = "3306"
	dbName = "test1"
)

//Db数据库连接池
var DB *sql.DB

// user表结构
type User struct {
	Name 	string
	Age 	int
	Id 		int

}

// 数据库操作类型
type HANDTYPE int
const (
	INIT		HANDTYPE  = 	iota
	INSERT
	SELECTBYNAME
	TRUNCATE
)

// 初始化数据库
func InitDB() (dbErr error) {
	dbPath := strings.Join([]string{userName,":",password,"@tcp(",ip,":",port,")/",dbName,"?charset=utf8"},"")

	DB,dbErr = sql.Open("mysql",dbPath)
	if dbErr != nil {
		return errors.Wrapf(dbErr,"main:InitDB open database fail:",dbPath)
	}
	DB.SetConnMaxLifetime(100)
	DB.SetConnMaxIdleTime(10)

	if dbErr = DB.Ping();dbErr != nil {
		fmt.Println("sss")
		return errors.Wrapf(dbErr,"main:InitDB connect database fail:%s",dbPath)
	}
	return nil
}

// 插入数据
func InsertUser(user User)(error) {
	tx, err := DB.Begin()
	if err != nil{
		return errors.Wrap(err,"main:InsertUser db begin failed")
	}

	stmt,err := tx.Prepare("insert into myuser(`name`,`age`) VALUES (?,?)")
	if err != nil {
		return errors.Wrap(err,"main:InsertUser Sql Prepare failed")
	}

	//将参数传递到sql语句中并且执行
	if _, errExec := stmt.Exec(user.Name, user.Age); errExec != nil {
		return errors.Wrap(errExec,"main:InsertUser Exec failed")
	}

	tx.Commit()
	return nil
}

// 删除数据
func DeleteUser(user User) error {
	//开启事务
	tx, err := DB.Begin()
	if err != nil{
		return errors.Wrap(err,"main:DeleteUser db begin failed")
	}
	//准备sql语句
	stmt, err := tx.Prepare("DELETE FROM myuser WHERE id = ?")
	if err != nil{
		return errors.Wrap(err,"main:DeleteUser Sql Prepare failed")
	}
	//设置参数以及执行sql语句
	if _, errExec := stmt.Exec(user.Id); errExec != nil {
		return errors.Wrap(errExec,"main:DeleteUser Exec failed")
	}

	//提交事务
	tx.Commit()
	return nil
}

// 清理数据表
func TruncateUser() error {
	//开启事务
	tx, err := DB.Begin()
	if err != nil{
		return errors.Wrap(err,"main:TruncateUser db begin failed")
	}
	//准备sql语句
	stmt, err := tx.Prepare("TRUNCATE table myuser")
	if err != nil{
		return errors.Wrap(err,"main:TruncateUser Sql Prepare failed")
	}
	//设置参数以及执行sql语句
	if _, errExec := stmt.Exec(); errExec != nil {
		return errors.Wrap(errExec,"main:TruncateUser Exec failed")
	}
	//提交事务
	tx.Commit()
	return nil
}

// 根据用户名查询
func QueryUserByName(name string) (interface{},error) {
	var user User
	sql := "SELECT * FROM myuser WHERE name = ? "
	err := DB.QueryRow(sql, name).Scan(&user.Id, &user.Age, &user.Name)
	if err != nil{
		return nil,errors.Wrapf(err,"main:QueryUserByName Scan failed:%s,%s",sql,name)
	}
	return &user,nil
}

// 数据库处理函数
func HandLeDbFunc(dType HANDTYPE,inInfo User) (outInfo interface{},dbErr error) {
	switch dType {
	case INIT:
		dbErr = InitDB()
	case INSERT:
		dbErr = InsertUser(inInfo)
	case TRUNCATE:
		dbErr = TruncateUser()
	case SELECTBYNAME:
		outInfo,dbErr = QueryUserByName(inInfo.Name)
	default:
		dbErr = errors.Errorf("main:HandLeDbFunc unknow db type:%d",dType)
	}
	return outInfo,dbErr
}

// 数据中间处理
func HandleDataFunc(inName string,queryName string) (interface{},error) {
	var outUser interface{}
	var err error = nil
	for i := HANDTYPE(0); i <= TRUNCATE; i++ {
		if i == SELECTBYNAME {
			outUser,err = HandLeDbFunc(i,User{queryName,22,0})
		} else {
			_,err = HandLeDbFunc(i,User{inName,22,0})
		}

		if err != nil {
			return nil,errors.WithMessagef(err,"main:HandleDataFunc handletype:%d",i)
		}
	}
	return outUser,nil
}

// 打印错误详细信息
func PrintErr(err error) {
	if err != nil {
		fmt.Printf("original error:%+v\n",errors.Cause(err))
		fmt.Printf("stack trace:\n%+v\n",err)
		if errors.Cause(err) == sql.ErrNoRows {
			doSomething()
		}
	}
}

// sql.ErrNoRows 错误做的一些业务
func doSomething() {
	fmt.Println("doSomething	1")
	fmt.Println("doSomething	2")
	fmt.Println("doSomething	3")
	fmt.Println("doSomething......")
}

// 主函数
func main()  {
	// 正常查询
	info, err := HandleDataFunc("walk","walk")
	if err != nil {
		PrintErr(err)
		return
	}
	if user,ok := info.(*User);ok {
		fmt.Printf("name:%s,age:%d,id:%d",user.Name,user.Age,user.Id)
	}

	// 异常查询
	info1, err1 := HandleDataFunc("walk","walk1")
	if err1 != nil {
		PrintErr(err1)
		return
	}
	if user,ok := info1.(*User);ok {
		fmt.Printf("name:%s,age:%d,id:%d",user.Name,user.Age,user.Id)
	}
	return
}