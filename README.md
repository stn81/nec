# nec redis proxy

## build
./scripts/build dev/prod/prod_realtime

## start
./outputs/bin/nec start

## tool cli example
VALUE如果为@开头，则从文件中读取value内容
```sh
# redis-cli set key value
./outputs/bin/nec cli set KEY VALUE
# redis-cli set key seconds value
./outputs/bin/nec cli setex KEY SECONDS VALUE
# redis-cli hset key field value
./outputs/bin/nec cli hset KEY FIELD VALUE
# redis-cli get key
./outputs/bin/nec cli get KEY
# redis-cli hget key field
./outputs/bin/nec cli hget KEY FIELD
```
## tool fetch example
```sh
./outputs/bin/nec fetch -p PARTITION -o OFFSET -d DUMP_PATH
```

## tool offset example
```sh
# show all partition commit offset for consumer group
./outputs/bin/nec offset
# show specific partition commit offset
./outputs/bin/nec offset -p PARTITION
# reset partition offset
./outputs/bin/nec offset -a set -p PARTITION -o OFFSET
```
