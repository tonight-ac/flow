package flow

import "context"

func newWorker(id int32, manager *Manager, workFunc WorkFunc) *Worker {
	return &Worker{
		id:       id,
		master:   manager,
		workFunc: workFunc,
	}
}

type Worker struct {
	id       int32
	master   *Manager
	workFunc WorkFunc
	values   []interface{}
}

func (w *Worker) AddValue(in interface{}) {
	w.values = append(w.values, in)
}

func (w *Worker) GetValue(index int32) interface{} {
	if int(index) >= len(w.values) {
		return nil
	}
	return w.values[index]
}

// wait 是否能开始
func (w *Worker) wait() (err error) {
	for _, u := range w.master.board[w.id] {
		if u != nil {
			err = <-u.result
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// hustle 工作函数
func (w *Worker) hustle(ctx context.Context) (err error) {
	// 等待所有前置工作完成
	err = w.wait()
	if err != nil { // 前置工作已经产生了错误，没有必要继续运行
		return err
	}

	err = w.workFunc(ctx, w)
	if err != nil {
		return err
	}

	return nil
}
