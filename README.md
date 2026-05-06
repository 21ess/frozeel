##  介绍

冻鳗糕手(frozen eel bot)：一个基于 Agentic AI 的 im 动漫猜谜游戏robot

基本数据流 v1
![image-20260418175612137](https://txcould-image-1318385221.cos.ap-nanjing.myqcloud.com/image/image-20260418175612137.png)

## 目录结构
```text
.
├── .env.example    # 模板环境文件（存放 telegram token，llm api key 等）
├── adapter         # im 适配层 （适配不同的 im 软件）
├── agent           # agent 层  
├── cmd     
├── game            # 游戏逻辑层
├── go.mod
├── internal        # 内部包
├── LICENSE
├── prompt          # prompt 模板
├── README.md   
└── store           # 存储层
    ├── mongo
    └── store.go
```

每个包入口有同名 `.go` 文件，定义核心接口，实现放在次级目录比如 `store/mongo`

## 游戏状态机
```mermaid 
stateDiagram-v2
    direction TB

    [*] --> StateIdle: 初始化
    
    state StateIdle {
        [*] --> WaitingForStart
        WaitingForStart --> WaitingForStart: 接收非 EventStart 事件<br>返回错误响应
    }

    state StateConfiguring {
        [*] --> ConfigPhase
        ConfigPhase --> ConfigPhase: EventConfigure<br>更新配置
        ConfigPhase --> ConfigPhase: 接收其他事件<br>返回错误提示
    }

    state StatePlaying {
        [*] --> GameInProgress
        
        GameInProgress --> GameInProgress: EventGuess<br>处理猜测
        GameInProgress --> GameInProgress: EventHint<br>发送提示
        GameInProgress --> AnswerRevealed: 猜对 / EventGiveUp / EventEnd
    }

    state StateEnded {
        [*] --> GameOver
        GameOver --> GameOver: 忽略所有事件
    }

    StateIdle --> StatePlaying: EventStart<br>生成随机角色
    StateIdle --> StateConfiguring: EventConfigure(预留)
    
    StateConfiguring --> StatePlaying: EventStart<br>生成随机角色
    StateConfiguring --> StateEnded: EventEnd<br>取消游戏
    
    StatePlaying --> StateEnded: EventGuess(正确)<br>恭喜猜对
    StatePlaying --> StateEnded: EventGiveUp<br>公布答案
    StatePlaying --> StateEnded: EventEnd<br>结束游戏

    note right of StateIdle
        初始状态
        - 等待 /start 命令
        - 其他操作返回错误
    end note

    note right of StateConfiguring
        配置阶段（预留）
        - 可配置时间范围、目录等
        - 当前未实现
    end note

    note right of StatePlaying
        游戏进行中
        - 接收玩家猜测
        - 提供提示
        - 支持放弃和结束
    end note

    note right of StateEnded
        游戏结束
        - 显示答案
        - 清理资源
        - 不再响应事件
    end note

```