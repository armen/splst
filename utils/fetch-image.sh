#!/bin/bash

xvfb-run --auto-servernum --server-args="-screen 0, 1024x768x24" cutycapt --min-width=1024 --url="$1" --out="$2"
