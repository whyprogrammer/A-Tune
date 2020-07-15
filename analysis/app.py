#!/usr/bin/python3
# -*- coding: utf-8 -*-
# Copyright (c) 2019 Huawei Technologies Co., Ltd.
# A-Tune is licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2019-10-29

"""
Application initialization, including log configuration, restful api registration.
"""

import os
import ssl
import sys
import logging
from configparser import ConfigParser
from logging.handlers import SysLogHandler
from flask import Flask
from flask_restful import Api

sys.path.insert(0, os.path.dirname(__file__) + "/../")
from analysis.resources import configurator, monitor, optimizer, collector, \
    classification, profile, train

LOG = logging.getLogger('werkzeug')
LOG.setLevel(logging.ERROR)

APP = Flask(__name__)
API = Api(APP)


def config_log(level):
    """app config log"""
    logging_format = logging.Formatter('atuned: %(asctime)s [%(levelname)s] '
                                       '%(name)s[line:%(lineno)d] : %(message)s')
    syslog_handler = SysLogHandler(address="/dev/log", facility=SysLogHandler.LOG_LOCAL0)
    syslog_handler.setFormatter(logging_format)

    root_logger = logging.getLogger()
    root_logger.setLevel(level)
    root_logger.addHandler(syslog_handler)


def main(filename, port):
    """app main function"""
    if not os.path.exists(filename):
        return
    config = ConfigParser()
    config.read(filename)

    level = logging.getLevelName(config.get("log", "level").upper())
    config_log(level)
    APP.config.update(SESSION_COOKIE_SECURE=True,
                      SESSION_COOKIE_HTTPONLY=True,
                      SESSION_COOKIE_SAMESITE='Lax')

    if port == 'rest':
        API.add_resource(configurator.Configurator, '/v1/setting', '/setting')
        API.add_resource(monitor.Monitor, '/v1/monitor', '/monitor')
        API.add_resource(collector.Collector, '/v1/collector', '/v1/collector')
        API.add_resource(profile.Profile, '/v1/profile', '/v1/profile')
        API.add_resource(train.Training, '/v1/training', '/v1/training')
    else:
        API.add_resource(optimizer.Optimizer, '/v1/optimizer', '/v1/optimizer/<string:task_id>')
        API.add_resource(classification.Classification, '/v1/classification', '/v1/classification')

    if config.has_option("server", "tls") and config.get("server", "tls") == "true":
        cert_file = config.get("server", "tlshttpcertfile")
        key_file = config.get("server", "tlshttpkeyfile")
        ca_file = config.get("server", "tlshttpcacertfile")
        context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
        context.load_cert_chain(certfile=cert_file, keyfile=key_file)
        context.load_verify_locations(ca_file)
        context.verify_mode = ssl.CERT_REQUIRED
        if port == 'rest':
            APP.run(host=config.get("server", "rest_host"), port=config.get("server", "rest_port"),
                    ssl_context=context)
        else:
            APP.run(host=config.get("server", "engine_host"), port=config.get("server", "engine_port"),
                    ssl_context=context)
    else:
        if port == 'rest':
            APP.run(host=config.get("server", "rest_host"), port=config.get("server", "rest_port"))
        else:
            APP.run(host=config.get("server", "engine_host"), port=config.get("server", "engine_port"))


if __name__ == '__main__':
    if len(sys.argv) != 3:
        sys.exit(-1)
    main(sys.argv[1], sys.argv[2])
