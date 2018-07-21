# 编译

````
sh build.sh
````
# 参数说明

````
sub_cmd: dump|import|clear

  -P string
        redis key prefix
  -a string
        redis passpord
  -d int
        redis db
  -f string
        导入或导出数据的文件名
  -h string
        redis host (default "127.0.0.1")
  -p int
        redis port (default 6379)

````
# 导出

````
./bin/redis_dump_mac dump -h 127.0.0.1 -p 6379 -a 'redis123123'  -P 'pc_v1*' -f out1.json

````

# 导入

````
./bin/redis_dump_mac import -h 127.0.0.1 -p 6379 -a 'redis123123'   -f out.json

````


# 删除
````
./bin/redis_dump_mac clear -h 127.0.0.1 -p 6379 -a 'redis123123'  -P 'pc_v1*'

````