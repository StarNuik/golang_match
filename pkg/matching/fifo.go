package matching

// import (
// 	"slices"

// 	"github.com/starnuik/golang_match/pkg/model"
// 	"github.com/starnuik/golang_match/pkg/schema"
// )

// func NewFifo(cfg KernelConfig) Kernel {
// 	return &fifoKernel{
// 		cfg: cfg,
// 	}
// }

// type fifoKernel struct {
// 	cfg KernelConfig
// }

// func (k *fifoKernel) Match(queue []*model.QueuedUser) []schema.MatchResponse {
// 	blockSize := k.cfg.MatchSize
// 	blocksCount := len(queue) / blockSize

// 	if blocksCount == 0 {
// 		return nil
// 	}

// 	slices.SortFunc(queue, func(l *model.QueuedUser, r *model.QueuedUser) int {
// 		return l.QueuedAt.Compare(r.QueuedAt)
// 	})

// 	out := []schema.MatchResponse{}
// 	for i := 0; i < blocksCount; i++ {
// 		sub := queue[i*blockSize : (i+1)*blockSize]
// 		resp := fillResponse(sub)
// 		out = append(out, resp)
// 	}

// 	return out
// }
