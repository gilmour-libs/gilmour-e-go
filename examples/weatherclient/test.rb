batch := engine.NewPipe(
	engine.NewRequestComposition("weather.fetch"),
	engine.NewRequestComposition("weather.group"),
	engine.NewParallel(
		engine.NewPipe(
			engine.NewFuncComposition(monthLambda("jan")),
			engine.NewParallel(
				engine.NewRequestComposition("weather.min"),
				engine.NewRequestComposition("weather.max"),
			),
		),
		engine.NewPipe(
			engine.NewFuncComposition(monthLambda("feb")),
			engine.NewParallel(
				engine.NewRequestComposition("weather.min"),
				engine.NewRequestComposition("weather.max"),
			),
		),
	),
)

data := G.NewMessage()
data.SetData("pune")

var out []*G.Message

(<-batch.Execute(data)).GetData(&out)
