# alog

[![Github-Actions](https://github.com/Saodd/alog/workflows/Go/badge.svg?branch=main)](https://saodd.github.io/alog/coverage.html)

An Error-Tracker with Logger in Golang. 给你的Go补充一套错误追踪机制。

## 背景

### 1. 它可以做什么？

帮你将函数抛出的异常“串连”起来，添加代码定位、上下文变量等功能。经过简单配置，即可直通 [sentry](https://sentry.io) .

### 2. 它的应用场景是什么？

最典型的场景是 Web应用，特别是 [gin框架](https://github.com/gin-gonic/gin) 所构建的应用。

当然，如你也可以在非Web上下文环境中使用它。

### 3. 它为什么叫这个名字？

因为在预想场景中，本库的出现频率可能仅次于`if err!=nil`，所以本库（以及其中的部分函数、类型）起名时以尽可能「简单、独特、好按」为原则。

## 用法

### 1. 普通地捕获错误

你应当在每一级处理错误时都加入`alog.CE`这个函数：

```go
func SomeFunction(ctx context.Context) error {
	if err != nil {
		alog.CE(ctx, err, alog.V{"你需要记录的变量": "你需要记录的值"})  // 这里！
		return err
	}
	return nil
}
```

### 2. 重要：ctx里要有Tracker对象

请使用我包装好的方法：

```go
func SomeFunction() error {
	ctx, cancel := WithTracker(context.Background())
    defer cancel()   // 一定要cancel！要养成习惯！

    if err != nil {
		alog.CE(ctx, err, alog.V{"你需要记录的变量": "你需要记录的值"})
		return err
	}
	return nil
}
```

当然，`context`里没有`Tracker`也行，那么错误会直接打在日志上。

### 3. 捷径：直接装入gin中间件

```go
func main() {
	g := gin.New()
	g.Use(alog.GinWithLogger(), alog.GinWithTracker(), alog.GinWithRecover())  // 注意顺序！
}
```

### 4. 可选：初始化！

```go
func main() {
    alog.InitAlog("v1.0.0", "https://...", "812793r713452d")  // 记得初始化！填入Sentry相关参数。
}
```

如果没有初始化Sentry相关配置，则仅仅在本地打印出错误日志。一般在开发阶段这样做。

### 5. 在非gin上下文中使用

```go
func anythingElse() {
    ctx, cancel := alog.WithTracker(context.Background())  // 仿context包风格
    defer cancel()
    defer alog.CERecover(ctx, V{"data": "你想要追踪的变量"})  // 会帮你recover
    panic(errors.New("我是一个异常！"))
}
```

> TIPS: 在panic情况下，只最多追踪6层调用栈。

### 6. 更多

请参考测试用例，或者直接读源码，并不难。
