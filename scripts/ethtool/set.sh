#!/bin/sh
# Copyright (c) 2019 Huawei Technologies Co., Ltd.
# A-Tune is licensed under the Mulan PSL v1.
# You can use this software according to the terms and conditions of the Mulan PSL v1.
# You may obtain a copy of Mulan PSL v1 at:
#     http://license.coscl.org.cn/MulanPSL
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v1 for more details.
# Create: 2019-10-29

RUN_HOME=$(
  cd "$(dirname "$0")"
  pwd
)
source "$RUN_HOME"/common.sh

command -v ethtool >/dev/null 2>&1
ret=$?
[ $ret -ne 0 ] && echo "\033[31m command ethtool is not exist \033[31m" && exit 1

value=$(get_ethtool_value "$@")
if [[ "$value" == $(eval echo "$*") ]]; then
  echo "no need to set"
  exit 0
fi

ethtool "$@"

