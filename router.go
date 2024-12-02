package botmother

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/telebot.v3"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfContext = reflect.TypeOf((*telebot.Context)(nil)).Elem()
)

func (bm *BotMother) registerCallbackHandler(name string, method *reflect.Method) {
	bm.callbackMap.Store(name, method)
}

func (bm *BotMother) findCallHandler(name string) (*reflect.Method, bool) {
	value, ok := bm.callbackMap.Load(name)
	if !ok {
		return nil, false
	}
	return value.(*reflect.Method), true
}

func (bm *BotMother) registerMessageHandler(name string, method *reflect.Method) {
	bm.messageMap.Store(name, method)
}

func (bm *BotMother) findMessageHandler(name string) (*reflect.Method, bool) {
	value, ok := bm.messageMap.Load(name)
	if !ok {
		return nil, false
	}
	return value.(*reflect.Method), true
}

func (bm *BotMother) InitRouter(router Router) {
	isHandler := func(method reflect.Method) bool {
		mt := method.Type
		if mt.NumIn() != 2 || mt.NumOut() != 1 {
			return false
		}
		if t := mt.In(1); t != typeOfContext || mt.Out(0) != typeOfError {
			return false
		}
		return true
	}

	bm.router = router

	t := reflect.TypeOf(router)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		name := method.Name
		if strings.HasPrefix(name, "OnBtn") {
			if isHandler(method) {
				bm.registerCallbackHandler(strings.TrimPrefix(name, "On"), &method)
			}
		} else if strings.HasPrefix(name, "OnCommand") {
			if isHandler(method) {
				bm.registerMessageHandler("/"+strings.ToLower(strings.TrimPrefix(name, "OnCommand")), &method)
			}
		}
	}
}

func (bm *BotMother) handleMessage(ctx telebot.Context) error {
	defer func() {
		if r := recover(); r != nil {
			bm.Logger.Error("Recovered", fmt.Sprintf("%v", r))
		}
	}()

	m := ctx.Message()
	ss := strings.Split(m.Text, " ")

	group := m.FromGroup() || m.FromChannel()

	handler, ok := bm.findMessageHandler(ss[0])
	if ok {
		if group {
			return nil
		}
		args := []reflect.Value{reflect.ValueOf(bm.router), reflect.ValueOf(ctx)}
		ret := handler.Func.Call(args)
		if i := ret[0].Interface(); i != nil {
			err := i.(error)
			if !strings.HasPrefix(err.Error(), "[bot]") {
				bm.Logger.Error(err)
			}
		}
	} else {
		if err := bm.router.DefaultMessageHandler(ctx); err != nil {
			if !strings.HasPrefix(err.Error(), "[bot]") {
				bm.Logger.Error(err)
			}
		}
	}

	return nil
}

func (bm *BotMother) handleCallback(ctx telebot.Context) error {
	defer func() {
		if r := recover(); r != nil {
			bm.Logger.Error("Recovered", fmt.Sprintf("%v", r))
		}
	}()

	c := ctx.Callback()
	data := string([]byte(c.Data)[1:])
	ss := strings.Split(data, "|")
	if len(ss) == 0 {
		return nil
	}
	c.Unique = ss[0]
	c.Data = strings.Join(ss[1:], "|")

	if ctx.Message().FromGroup() || ctx.Message().FromChannel() {
		return nil
	}

	handler, ok := bm.findCallHandler(c.Unique)
	if ok {
		args := []reflect.Value{reflect.ValueOf(bm.router), reflect.ValueOf(ctx)}
		ret := handler.Func.Call(args)
		if i := ret[0].Interface(); i != nil {
			err := i.(error)
			if !strings.HasPrefix(err.Error(), "[bot]") {
				bm.Logger.Error(err)
			}
		}
	}

	return nil
}
