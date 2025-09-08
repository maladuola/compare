# Mogost 工具集

专业的文件处理和比较工具套件，为 Mogost 团队开发的一系列实用工具。

## 功能特性

### 工具1：文件比较工具
- 上传两个文件进行内容比较
- 支持类似 Beyond Compare 的左右对齐显示
- 红色字体高亮显示差异部分
- 支持逐行比较和差异分析

### 工具2：CSV 文件查看器
- 上传 CSV 文件并查看其内容
- 支持表格形式展示
- 自动识别数据类型和列结构
- 提供文件统计信息

### 工具3：压缩文件交易比较
- 上传 ZIP 压缩文件，自动解压
- 分析目录结构（支持 ABC、ABD、ABE 等目录）
- 自动识别交易文件（babyy-risk-{id}.txt 和 candyy-risk-{id}.txt）
- 比较同一交易ID对应的两个文件的差异
- 支持多目录批量比较

## 技术栈

- **后端**: Go + Gin框架
- **前端**: HTML5 + CSS3 + JavaScript
- **文件处理**: 支持多种文件格式
- **差异算法**: 基于 go-diff 库实现

## 快速开始

### 1. 环境要求
- Go 1.21 或更高版本
- 现代浏览器（支持HTML5）

### 2. 构建和运行

```bash
# 下载依赖
go get -v

# 构建程序
go build -v -o tools

# 运行程序
./tools
```

或者使用构建脚本：

```bash
# 给脚本执行权限
chmod +x build.sh

# 运行构建脚本
./build.sh

# 运行程序
./tools
```

### 3. 访问工具

在浏览器中访问：http://localhost:8080

## 使用说明

### 文件比较工具
1. 点击"文件比较工具"卡片
2. 上传两个要比较的文件
3. 系统会自动进行内容比较
4. 查看左右对齐的差异显示

### CSV 文件查看器
1. 点击"CSV 文件查看器"卡片
2. 上传 CSV 文件
3. 查看表格形式的数据展示
4. 获取文件统计信息

### 压缩文件交易比较
1. 点击"压缩文件交易比较"卡片
2. 上传 ZIP 压缩文件
3. 系统自动解压并分析目录结构
4. 查看交易文件的差异比较结果

## 项目结构

```
mogost-tools/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块文件
├── build.sh               # 构建脚本
├── README.md              # 项目说明
├── tools/                 # 工具包
│   ├── file_compare.go    # 文件比较工具
│   ├── csv_viewer.go      # CSV查看器
│   └── archive_compare.go # 压缩文件比较工具
├── templates/             # 前端模板
│   └── index.html         # 主页面
├── uploads/               # 上传文件目录
│   ├── file-compare/      # 文件比较上传目录
│   ├── csv/               # CSV文件上传目录
│   └── archive-compare/   # 压缩文件上传目录
└── temp/                  # 临时文件目录
```

## API 接口

### 文件比较工具
- `POST /api/file-compare/upload` - 上传文件
- `GET /api/file-compare/compare` - 比较文件

### CSV 查看器
- `POST /api/csv/upload` - 上传CSV文件
- `GET /api/csv/view` - 查看CSV内容

### 压缩文件比较
- `POST /api/archive-compare/upload` - 上传压缩文件
- `GET /api/archive-compare/compare` - 比较交易文件

## 开发说明

### 添加新工具
1. 在 `tools/` 目录下创建新的工具文件
2. 实现相应的处理函数
3. 在 `main.go` 中注册路由
4. 在前端添加相应的界面

### 自定义样式
- 修改 `templates/index.html` 中的 CSS 样式
- 支持响应式设计和现代UI风格

## 注意事项

1. 上传的文件会保存在 `uploads/` 目录下，请定期清理
2. 压缩文件解压后会占用临时空间，处理完成后会自动清理
3. 建议在生产环境中配置适当的文件大小限制
4. 确保服务器有足够的磁盘空间处理大文件

## 许可证

本项目为 Mogost 团队内部使用工具，版权所有。

## 联系方式

如有问题或建议，请联系 Mogost 团队开发组。
