@echo off
setlocal
cd /d %~dp0

echo [1/3] 构建前端...
cd web
call npm run build || (echo 前端构建失败 & exit /b 1)
cd ..

echo [2/3] 拷贝前端产物到 server/dist...
if exist server\dist rmdir /s /q server\dist
mkdir server\dist
xcopy web\dist server\dist /e /i /q >nul || (echo 拷贝失败 & exit /b 1)

echo [3/3] 构建后端（嵌入前端）...
cd server
set CGO_ENABLED=0
go build -ldflags "-s -w" -o 简历制作平台.exe . || (echo 后端构建失败 & exit /b 1)
cd ..

echo.
echo 构建完成: server\简历制作平台.exe
echo 双击运行即可，数据保存在 exe 同目录的 data 文件夹
