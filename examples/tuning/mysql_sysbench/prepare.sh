#!/bin/sh
# Copyright (c) lingff(ling@stu.pku.edu.cn),
# School of Software & Microelectronics, Peking University.
#
# A-Tune is licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
#
# Create: 2021-05-24

path=$(
  cd "$(dirname "$0")"
  pwd
)

echo "initializing MySQL..."
rm -f /etc/my.cnf
cp my.cnf /etc
service mysql stop
taskset -c 0,1 service mysql start
/usr/local/mysql/bin/mysql -uroot -p123456 -e "DROP DATABASE IF EXISTS sbtest;"
/usr/local/mysql/bin/mysql -uroot -p123456 -e "CREATE DATABASE sbtest;"

echo "checking sysbench..."
sysbench --version
if [ $? -ne 0 ]; then   
    echo "sysbench FAILED";   
    exit 1;   
fi

echo "update the client and server yaml files"
sed -i "s#sh .*/mysql_sysbench_benchmark.sh#sh $path/mysql_sysbench_benchmark.sh#g" $path/mysql_sysbench_client.yaml
sed -i "s#cat .*/sysbench_oltp_read_write.log#cat $path/sysbench_oltp_read_write.log#g" $path/mysql_sysbench_client.yaml

echo "copy the server yaml file to /etc/atuned/tuning/"
cp $path/mysql_sysbench_server.yaml /etc/atuned/tuning/
