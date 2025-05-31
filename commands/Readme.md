### commands

That helper to implement some kind of event system, when we have some functions, which we should call by specific event, but we are not able to directly call such functions.
We run them in few gorutines, and all data pass through context.

For example:
```go
// We subscribe on event "test", and call `Set` function, which we take by closure
GetEventManager().Subscribe("test", commands.EventHandlerFunc(func(ctx context.Context) {
    scene, ok := ctx.Value(ctxSceneName).(Scene)
    if !ok {
        return
    }
    err := m.Set(scene)
    if err != nil {
        fastlog.WLogger().Errorw("Error setting new scene", "err", err)
    }
}))

// Then somewhere in code, we can call, and ctx should have expected `ctxSceneName`
GetEventManager().Call(ctx, "test")
```