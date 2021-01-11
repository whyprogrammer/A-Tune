#!/usr/bin/python3
# -*- coding: utf-8 -*-
# Copyright (c) 2020 Huawei Technologies Co., Ltd.
# A-Tune is licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2020-12-17

"""
Routers for /v1/UI/tuning url.
"""

import logging
import json
from flask import abort
from flask_restful import Resource

from analysis.engine.parser import UI_TUNING_GET_PARSER
from analysis.engine.database import trigger_tuning

LOGGER = logging.getLogger(__name__)
CORS = [('Access-Control-Allow-Origin', '*')]


class UiTuning(Resource):
    """restful api for web ui tuning page"""

    def get(self, cmd):
        """restful api get"""
        if not cmd:
            abort(404, 'does not get command')

        args = UI_TUNING_GET_PARSER.parse_args()
        if cmd == 'initialPage':
            uid = args.get('uid')
            status = args.get('status')
            res = trigger_tuning.get_tuning_list(int(uid), status)
            return json.dumps({'message': res}), 200, CORS

        if cmd == 'rename':
            name = args.get('name')
            new_name = args.get('newName')
            res, reason = trigger_tuning.rename_tuning(name, new_name)
            return json.dumps({'rename': res, 'reason': reason}), 200, CORS

        if cmd == 'chooseFile':
            name = args.get('name')
            res = trigger_tuning.tuning_exist(name)
            return json.dumps({'isExist': res}), 200, CORS

        if cmd == 'initialChart':
            name = args.get('name')
            status = args.get('status')
            if not trigger_tuning.tuning_exist(name):
                return json.dumps({'isExist': False}), 200, CORS
            response_obj = trigger_tuning.get_tuning_info(status, name)
            return json.dumps(response_obj), 200, CORS

        if cmd == 'compareWith':
            name = args.get('name')
            line = args.get('line')
            response_obj = trigger_tuning.get_tuning_data('finished', name, line)
            return json.dumps(response_obj), 200, CORS

        if cmd == 'getTuningData':
            name = args.get('name')
            status = args.get('status')
            line = args.get('line')
            response_obj = trigger_tuning.get_tuning_data(status, name, line)
            return json.dumps(response_obj), 200, CORS

        if cmd == 'getTuningStatus':
            name = args.get('name')
            res = trigger_tuning.get_tuning_status(name)
            return json.dumps({'status': res}), 200, CORS
        return '', 200, CORS