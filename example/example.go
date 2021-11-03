package main

import (
	"context"
	"fmt"
	"github.com/tonight-ac/flow"
)

func main() {
	ctx := context.Background()

	manager := flow.NewManager(ctx)

	// 添加A任务
	aid, _ := manager.AddWorker(ctx, func(ctx context.Context, worker *flow.Worker) error {
		// 进行相关操作
		s := "我是A任务"
		fmt.Println(s)
		worker.AddValue(s)
		return nil
	})

	// 添加B任务，依赖A任务完成
	bid, _ := manager.AddWorker(ctx, func(ctx context.Context, worker *flow.Worker) error {
		s := manager.GetWorkerValue(aid, 0).(string) // A任务的结果数据
		s = "我是B任务，我收到指令：" + s
		fmt.Println(s)
		worker.AddValue(s)
		return nil
	}, aid)

	// 添加C任务，依赖A任务和B任务完成
	_, _ = manager.AddWorker(ctx, func(ctx context.Context, worker *flow.Worker) error {
		s1 := manager.GetWorkerValue(aid, 0).(string) // A任务的结果数据
		s2 := manager.GetWorkerValue(bid, 0).(string) // A任务的结果数据
		s := "我是C任务，我收到指令：" + s1 + s2
		fmt.Println(s)
		worker.AddValue(s)
		return nil
	}, aid, bid)

	manager.Start(ctx)

	err := manager.Done(ctx)
	fmt.Println(err)
}
