偶数是服务器发出，奇数是客户端发出的命令
命令      描述
byte(0)   服务器告诉客户端从哪开始传
byte(1)   客户端在正式传文件数据的时候必须带上这个命令，
byte(2)   服务器告诉客户端圆满完成
byte(3)   客户端告知服务器已传完
byte(4)   服务器告诉客户端流重复了
byte(8)   服务器告诉客户端最后结尾时长度不对
byte(6)   服务器告诉客户端文件送去切的时候出错

fid   := r.FormValue("fid")//唯一标识视频的id
sid:=r.FormValue("sid")	//用户登录的那个sid
examtype := r.FormValue("examtype")  //考试类型
area := r.FormValue("area")//地区
format :=r.FormValue("format") //视频格式
fnickname :=r.FormValue("fnickname") //视频的名称