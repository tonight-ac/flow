package flow

import "context"

const DefaultWorkersNum = 20
const DefaultIDCounter = 1

type WorkFunc func(context.Context, *Worker) error
