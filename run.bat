@echo off
set filename=client/DNS_Cache.info

for /F "tokens=*" %%A in  (%filename%) do  (
   start cmd /k "go run server/server.go %%A %firstToken%"
)
@echo on
