# issue2md

一个命令行和网页工具，用于将GitHub issue、discussion或pull request转换为Markdown格式文件。

>此仓库中的大部分内容是由人工智能生成的!

## 命令行模式

### 安装issue2md命令行工具

```
$ go install github.com/bigwhite/issue2md/cmd/issue2md@latest
```

### 将 Issue/Discussion/Pull Request 转换为Markdown

```
用法: issue2md [flags] url [markdown-file]
参数:
  url            要转换的GitHub issue、discussion或pull request的URL。
  markdown-file  (可选) 输出的Markdown文件。
标志:
  -enable-reactions
    	在输出中包含reactions。
  -enable-user-links
    	在输出中包含评论者的profile链接
```

## 网页模式

### 安装并运行issue2md web

#### 基于Docker镜像运行(推荐)

```
$docker run -d -p 8080:8080 bigwhite/issue2mdweb
```

#### 从源码构建安装

```
$ git clone https://github.com/bigwhite/issue2md.git
$ make web
$ ./issue2mdweb
服务器正在运行在 http://0.0.0.0:8080
```

### 将内容转换为 Markdown

在浏览器中打开 localhost:8080：

![](./screen-snapshot.png)

输入您想要转换的 issue、discussion 或 pull request 的URL，然后点击“Convert”按钮！