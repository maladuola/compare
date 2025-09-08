#!/bin/bash

echo "Mogost 工具集构建脚本"
echo "======================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

echo "1. 下载依赖..."
go get -v

if [ $? -ne 0 ]; then
    echo "错误: 下载依赖失败"
    exit 1
fi

echo "2. 构建程序..."
go build -v -o tools

if [ $? -ne 0 ]; then
    echo "错误: 构建失败"
    exit 1
fi

echo "3. 创建必要目录..."
mkdir -p uploads/file-compare
mkdir -p uploads/csv
mkdir -p uploads/archive-compare
mkdir -p static
mkdir -p templates
mkdir -p temp

echo "4. 设置权限..."
chmod +x tools

echo "构建完成！"
echo ""
echo "使用方法："
echo "1. 运行程序: ./tools"
echo "2. 在浏览器中访问: http://localhost:8080"
echo ""
echo "工具说明："
echo "- 工具1: 文件比较工具 - 上传两个文件进行内容比较"
echo "- 工具2: CSV查看器 - 上传CSV文件查看内容"
echo "- 工具3: 压缩文件交易比较 - 上传ZIP文件，比较交易文件差异"
