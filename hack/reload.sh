#!/bin/bash

echo "ready to reload exporter"
curl -X POST http://127.0.0.1:8012/manager/reload
echo ""


