package main

import (
	chai "github.com/mhamedGd/chai"
)

var game chai.App
var backgroundDims chai.Vector2f = chai.Vector2f{X: 800 / WORLD_SCALING, Y: 600 / WORLD_SCALING}

const WORLD_SCALING = 10.0

func main() {
	game = chai.App{
		Title:  "Test",
		Width:  800,
		Height: 600,
		OnStart: func() {
			SplashSceen = chai.NewScene()
			SceneOne = chai.NewScene()
			SplashSceen.OnSceneStart = SplashScreenStart
			SceneOne.OnSceneStart = OnSceneOneStart
			chai.ChangeScene(&SplashSceen)

			chai.ScaleView(WORLD_SCALING)
		},
		OnUpdate: func(f float32) {},
		OnDraw: func() {
			if chai.GetNumberOfFingersTouching() > 0 {
				chai.Shapes.DrawCircle(chai.GetMouseWorldPosition(), 0.25, chai.NewRGBA8(0, 0, 0, 255))
			}
		},
		OnEvent: func(ae *chai.AppEvent) {},
	}

	chai.Run(&game)
}
