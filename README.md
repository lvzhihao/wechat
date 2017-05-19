# wechat
微信公众平台SDK

* Required Go1.8+

## install
```
go get github.com/lvzhihao/wechat
```

## build
support godep
```
godep go build
```
other
```
go get ./...
go build
```
Tips:
```
go get github.com/tools/godep
```

## import 
```
touch wechat.json
```
edit wechat.json
```
[
    {
        "name": "公众号1",
        "appid": "wx09xxxxxxxx",
        "appsecret": "81axxxxxxxx",
        "receive_token": "eyNxxxxxxxx",
        "callback_url": "demo.com"
    },
    {
        "name": "公众号w",
        "appid": "wx09xxxxxxxx",
        "appsecret": "81axxxxxxxx",
        "receive_token": "eyNxxxxxxxx",
        "callback_url": "demo.com"
    }
]
```
run import command
```
./wechat mgr import --mongo mongodb://127.0.0.0/wechat
```
default usage
```
manager wechat config

Usage:
  wechat mgr [flags]

Flags:
  -h, --help           help for mgr
      --mongo string   [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options] (default "mongodb://127.0.0.1/wechat")
```
## list key config
```
./wechat mgr list
+----------------+--------------------+----------+----------------------------------+
|      NAME      |       APPID        |   KEY    |              SECRET              |
+----------------+--------------------+----------+----------------------------------+
| xxxxxxxxxxxx01 | wx0xxxxxxxxxxxxxxx | xxxxxxxx | xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+----------------+--------------------+----------+----------------------------------+
| xxxxxxxxxxxx02 | wx1xxxxxxxxxxxxxxx | xxxxxxxx | xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+----------------+--------------------+----------+----------------------------------+
```

## run server
```
./wechat server --addr 127.0.0.1:9099 --mongo mongodb://127.0.0.1/wechat
```
recommend support https
```
./wechat server --addr 127.0.0.1:9099 --mongo mongodb://127.0.0.1/wechat --cert ssl/server.cert --key ssl/server.key
```
or nginx proxy support ssl
```
proxy_set_header Host $host;
proxy_set_header X-Forwarded-Proto https;
proxy_pass http://127.0.0.1:9099
```
default usage
```
Auto Refresh Access Token. For example:

wechat server --app_id=xxxx --app_secret=xxxx

Usage:
  wechat server [flags]

Flags:
      --addr string           代理监听地址 (default "127.0.0.1:9099")
      --cert string           ssl证书
      --consul string         consul api
      --consul_token string   consul acl token
      --debug                 display debug log
  -h, --help                  help for server
      --key string            ssl证书
      --mongo string          [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options] (default "mongodb://127.0.0.1/wechat")
```

## usage
fetch wechat callbackips
```
curl http://127.0.0.1:9099/cgi-bin/getcallbackip?key=KEY&secret=SECRET
{"ip_list":["101.226.62.77","101.226.62.78","101.226.62.79","101.226.62.80","101.226.62.81","101.226.62.82","101.226.62.83","101.226.62.84","101.226.62.85","101.226.62.86","101.226.103.59","101.226.103.60","101.226.103.61","101.226.103.62","101.226.103.63","101.226.103.69","101.226.103.70","101.226.103.71","101.226.103.72","101.226.103.73","140.207.54.73","140.207.54.74","140.207.54.75","140.207.54.76","140.207.54.77","140.207.54.78","140.207.54.79","140.207.54.80","182.254.11.203","182.254.11.202","182.254.11.201","182.254.11.200","182.254.11.199","182.254.11.198","59.37.97.100","59.37.97.101","59.37.97.102","59.37.97.103","59.37.97.104","59.37.97.105","59.37.97.106","59.37.97.107","59.37.97.108","59.37.97.109","59.37.97.110","59.37.97.111","59.37.97.112","59.37.97.113","59.37.97.114","59.37.97.115","59.37.97.116","59.37.97.117","59.37.97.118","112.90.78.158","112.90.78.159","112.90.78.160","112.90.78.161","112.90.78.162","112.90.78.163","112.90.78.164","112.90.78.165","112.90.78.166","112.90.78.167","140.207.54.19","140.207.54.76","140.207.54.77","140.207.54.78","140.207.54.79","140.207.54.80","180.163.15.149","180.163.15.151","180.163.15.152","180.163.15.153","180.163.15.154","180.163.15.155","180.163.15.156","180.163.15.157","180.163.15.158","180.163.15.159","180.163.15.160","180.163.15.161","180.163.15.162","180.163.15.163","180.163.15.164","180.163.15.165","180.163.15.166","180.163.15.167","180.163.15.168","180.163.15.169","180.163.15.170","101.226.103.0\/25","101.226.233.128\/25","58.247.206.128\/25","182.254.86.128\/25","103.7.30.21","103.7.30.64\/26","58.251.80.32\/27","183.3.234.32\/27","121.51.130.64\/27"]}
```
wechat oauth
```
$code = $request->input("code");
if($code != "") {
    $client = new Client(['timeout'=> 2.0]);
    $time = time();
    $res = $client->request('GET', 'http://127.0.0.1:9099/connect/oauth2/access_token', [
        'query' => [
            'key' => KEY,
            'secret' => SECRET,
            'code' => $code,
            'time' => $time
        ]
    ]);
    if($res->getStatusCode() == "200") {
        $data = $res->getBody();
    }
    return $data;
} else {
    return redirect("https://available.domain/connect/oauth2/authorize?key=KEY&secret=SECRET&scope=snsapi_userinfo&redirect_uri=".urlencode(url()->current()));
}
```

## feature
* wechat api proxy
* wechat oauth proxy
* auto refresh wechat token
* auto refresh user token
* auto receive message 
* support consul

## todo list
* support receive message queue
* support receive message plugin
* supoort acl
