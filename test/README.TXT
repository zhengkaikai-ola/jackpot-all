测试服务器地址:172.19.186.37:9800
更改下注等级接口:
    uid :1
    appid:1
    level:2
    使用URL:http://172.19.186.37:9800/changeSpinLevel?uid=1&appid=1&level=2
    成功：状态码200，Body无数据
    失败：状态码500，Body记录错误信息
开奖接口:
    uid :1
    appid:1
    使用URL:http://172.19.186.37:9800/spin?uid=1&appid=1
    成功：状态码200，Body为FinalResult对应的JSON字符串
    失败：状态码500，Body记录错误信息字符串
查询配置接口：
    使用URL:http://172.19.186.37:9800/config
    成功：状态码200，Body为JackpotConfig对应的JSON字符串
    失败：状态码500，Body记录错误信息字符串
查询广播接口：
    使用URL:http://172.19.186.37:9800/broadcast
    成功：状态码200，Body为JackpotBroadcast对应的JSON字符串
    失败：状态码500，Body记录错误信息字符串
加钱测试接
    uid :1
    appid:1
    使用URL:http://172.19.186.37:9800/addMoney?uid=1&appid=1
    成功：状态码200，Body为JackpotBroadcast对应的JSON字符串
    失败：状态码500，Body记录错误信息字符串口：