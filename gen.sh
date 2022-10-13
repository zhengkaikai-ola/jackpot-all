#!/bin/sh

if [ "$1" ]
then 
	# 生成 proto
	./gf gen pbentity -t $1
	
	# 进入 proto 目录
	cd proto

	# 生成 pb 程序
	protoc --go_out=. entity_$1.proto

	# 返回
	cd ..

	# 注入tag
	protoc-go-inject-tag -input=app/pb/entity_$1.pb.go

	# 生成dao

	if [ "$2" ]
	then
		./gf gen dao -g $2 -t $1
	else
		./gf gen dao -g default -t $1
	fi
else
	echo "Wrong parameter. It should be like this, ./gen.sh table_name database_link_name"
fi
