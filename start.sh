#!/bin/bash

for id in {0..15}; do
  val=$(printf "%02d" $id)
  /go/bin/service --bind=127.0.0.1:90$val --id=$id &
done

tail -f /dev/null
