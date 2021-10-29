# -
我们在数据库操作的时候，比如dao层中遇到一个sql.ErrNoRows 的时候，是否应该Wrap这个error，抛给上层。为什么，应该怎么做请写出代码？

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
// 完整代码main.go
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

