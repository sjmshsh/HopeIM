package HopeIM

import (
	"errors"
	"fmt"
	"github.com/sjmshsh/HopeIM/wire/pkt"
	"sync"
)

var ErrSessionLost = errors.New("err:session lost")

type Router struct {
	handlers *FuncTree
	pool     sync.Pool
}

func NewRouter() *Router {
	r := &Router{
		handlers: NewTree(),
	}
	r.pool.New = func() interface{} {
		return BuildContext()
	}
	return r
}

func (s *Router) Handle(commond string, handlers ...HandlerFunc) {
	s.handlers.Add(commond, handlers...)
}

func (s *Router) Serve(packet *pkt.LogicPkt, dispather Dispather, cache SessionStorage, session Session) error {
	if dispather == nil {
		return fmt.Errorf("dispather is nil")
	}
	if cache == nil {
		return fmt.Errorf("cache is nil")
	}
	ctx := s.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.request = packet
	ctx.Dispather = dispather
	ctx.SessionStorage = cache
	ctx.session = session

	s.serveContext(ctx)
	s.pool.Put(ctx)
	return nil
}

func (s *Router) serveContext(ctx *ContextImpl) {
	chain, ok := s.handlers.Get(ctx.Header().Command)
	if !ok {
		ctx.handlers = []HandlerFunc{handleNoFound}
		ctx.Next()
		return
	}

	ctx.handlers = chain
	ctx.Next()
}

func handleNoFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResp{Message: "NotImplemented"})
}

type FuncTree struct {
	nodes map[string]HandlersChain
}

func NewTree() *FuncTree {
	return &FuncTree{
		nodes: make(map[string]HandlersChain, 10),
	}
}

func (t *FuncTree) Add(path string, handlers ...HandlerFunc) {
	t.nodes[path] = append(t.nodes[path], handlers...)
}

func (t *FuncTree) Get(path string) (HandlersChain, bool) {
	f, ok := t.nodes[path]
	return f, ok
}
