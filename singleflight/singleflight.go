package singleflight

import "sync"

type call struct {
	sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	sync.Mutex
	dict map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.Lock()
	defer g.Unlock()

	if g.dict == nil {
		g.dict = make(map[string]*call)
	}

	if callGot, ok := g.dict[key]; ok {
		callGot.Wait()
		return callGot.val, callGot.err
	}

	callToDo := new(call)
	callToDo.Add(1)
	g.dict[key] = callToDo
	callToDo.val, callToDo.err = fn()
	callToDo.Done()
	delete(g.dict, key)

	return callToDo.val, callToDo.err
}
