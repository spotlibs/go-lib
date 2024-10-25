package ctx

import (
	"context"

	"github.com/goravel/framework/contracts/queue"
)

// NewFromWorker return new context that has given signature and request id as
// the metadata inside the context.
func NewFromWorker(sig, reqId string) context.Context {
	var meta Metadata
	meta.SignaturePath = sig
	meta.ReqId = reqId
	return context.WithValue(context.Background(), contextKey, meta)
}

// ToQueue append request id from given context to queue.Arg.
func ToQueue(c context.Context) (out []queue.Arg) {
	out = append(out, queue.Arg{Type: "string", Value: Get(c).ReqId})
	return
}

// NewFromQueue capture request/task id from given queue value in job, also
// set given signature as signature path in metadata context then return the
// context.
//
// This function expect the request id is in the first argument, and can be
// safely used if sending the job to queue using ToQueue.
func NewFromQueue(sig string, q ...any) context.Context {
	var meta Metadata
	if len(q) > 0 {
		if v, ok := q[0].(string); ok {
			meta.ReqId = v
		}
	}
	meta.SignaturePath = sig
	return context.WithValue(context.Background(), contextKey, meta)
}
