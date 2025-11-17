#! /bin/bash

sh build_web.sh
cd dist
python3 -m http.server 8000