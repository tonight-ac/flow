package flow

import (
	"context"
	"sync"
)

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		iDCounter: DefaultIDCounter,
		workers:   make(map[int32]*Worker),
	}

	// 初始化图
	m.board = make([][]*WorkUnit, DefaultIDCounter+DefaultWorkersNum<<1)
	for i := range m.board {
		m.board[i] = make([]*WorkUnit, DefaultIDCounter+DefaultWorkersNum<<1)
	}

	// 默认开4倍错误channel空间
	m.errors = make(chan error, DefaultWorkersNum<<2)

	return m
}

type Manager struct {
	iDCounter int32
	workers   map[int32]*Worker
	board     [][]*WorkUnit // 初始化一个默认大小，比如100，IDCounter超了会自动扩容

	wg     sync.WaitGroup
	errors chan error
}

func (m *Manager) GetWorkerValue(id, index int32) interface{} {
	if _, ok := m.workers[id]; !ok {
		return nil
	}

	return m.workers[id].GetValue(index)
}

func (m *Manager) getAndIncrID() int32 {
	id := m.iDCounter
	m.iDCounter++
	return id
}

func (m *Manager) Start(ctx context.Context) {
	for _, w := range m.workers {
		w := w

		m.wg.Add(1)
		go func() {
			defer m.wg.Done()

			if w != nil {
				err := w.hustle(ctx)
				if err != nil {
					m.errors <- err
				}

				// 既然运行完成，更新棋盘
				m.refreshBoard(w.id, err)
			}
		}()
	}
}

func (m *Manager) refreshBoard(workId int32, result error) {
	for _, row := range m.board {
		if unit := row[workId]; unit != nil {
			unit.result <- result
		}
	}
}

func (m *Manager) closeBoard() {
	for _, row := range m.board {
		for _, unit := range row {
			if unit != nil {
				close(unit.result)
			}
		}
	}
}

func (m *Manager) Done(ctx context.Context) (err error) {
	defer close(m.errors) // 关闭errors
	defer m.closeBoard()  // 关闭棋盘上的channel

	m.wg.Wait()

	select {
	case err = <-m.errors:
		return err
	default:
		return nil
	}
}

func (m *Manager) AddWorker(ctx context.Context, workFunc WorkFunc, prevIDs ...int32) (id int32, err error) {
	id = m.getAndIncrID()

	// 新建Worker
	worker := newWorker(id, m, workFunc)

	// 添加worker进map
	m.workers[id] = worker

	// 如果棋盘大小不够，扩容操作
	if int(id) >= len(m.board) {
		m.resize()
	}

	for _, prevID := range prevIDs {
		m.board[id][prevID] = &WorkUnit{
			result: make(chan error, DefaultWorkersNum),
		}
	}

	return id, nil
}

// resize 库容Board为目前的2倍
func (m *Manager) resize() {
	// 记录目前节点数量
	size := len(m.board)

	// 扩容errors
	m.errors = make(chan error, size<<1)

	// 扩容board
	newBoard := make([][]*WorkUnit, size<<1)
	for i, row := range newBoard {
		row = make([]*WorkUnit, size<<1)
		if i < len(m.board) {
			for j, o := range m.board[i] {
				row[j] = o
			}
		}
	}

	m.board = newBoard
}

type WorkUnit struct {
	result chan error // 运行结果channel
}
