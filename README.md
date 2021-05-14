# go-aspect

## 简介

go-aspect是一个为golang提供切面编程可能性的工具，可以使用该工具，替换原有的go build来进行编译，将预先配置好的切面编织到目标代码中

## 效果图

![效果图](https://cdn.justice-love.com/image/png/go-aspect.png)

## 安装方式

### 源码安装

1. 下载源码到本地
2. 切换到源码目录，运行 make install

## 使用简介

编织是基于目标代码根目录下的`aspect.point`文件进行，文件样例如下：
```bigquery

import "context"
import "fmt"
import "time"

point after(test.*X.Inject(c Context)) {
	fmt.Println("456")
	fmt.Println("789")
}

point before(test.Do(c Context)) {
	{{c}} = context.WithValue({{c}}, "date", time.Now())
}

```

文件分成两部分，
1. 添加import依赖
2. 对应的point切入点信息，下文是对其的解释

```bigquery
//固定值point    织入方式  对应的函数  方法参数
point           before(test.Do(c Context)) {
//  织入的代码
    {{c}} = context.WithValue({{c}}, "date", time.Now())
}
```

1. 织入方式：目前支持 before，after， defer三种方式
2. 对应函数：包名.[方法的接收者].函数名，方法接收者如果没有可以不填
    * 如果织入的代码中使用到了接收者，接收者的参数为去掉*前缀的变量，如test.*X.Inject中，使用X，并且需要使用{{X}}占位符表明接收者
3. 方法参数：需要按照代织入的方法参数顺序进行编写，现支持的参数类型，struct（使用对应的类型，不需要带前缀），map，slice，array，func，interface。
    * `point after(test.*X.Inject(c Context, m map, s slice, i interface, f func)) {`
    * 如果织入的代码中有使用到对应的参数，请使用{{c}}这样的占位符表明

### 编译

1. xgc build：编译代码，和go build一样，直接执行编译后的文件即可
2. xgc debug：生成织入后的代码，用以问题排查，代码存放在$USER_HOME/.xgc下
