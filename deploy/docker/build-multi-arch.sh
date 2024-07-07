#!/bin/bash
#
# Copyright 2024 Daniel C. Brotsky. All rights reserved.
# All the copyrighted work in this repository is licensed under the
# open source MIT License, reproduced in the LICENSE file.
#
docker build --platform linux/amd64,linux/arm64 -t clickonetwo/adobe_usage_tracker:latest .
