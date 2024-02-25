package main

import chai "github.com/mhamedGd/chai"

var SplashSceen chai.Scene

var sprite_render_system chai.SpriteRenderOriginSystem

var float32_animation_system chai.TweenAnimatorSystemFloat32
var splash_animation_sync_system SplashAnimationSyncSystem

func SplashScreenStart() {
	SplashSceen.Background = chai.NewRGBA8(0.0, 0.0, 0.0, 1.0)
	chai.ScaleView(1)
	sprite_render_system.Scale = 1 / 20.0
	sprite_render_system.Sprites = &chai.Sprites
	SplashSceen.NewRenderSystem(&sprite_render_system)

	SplashSceen.NewUpdateSystem(&float32_animation_system)
	SplashSceen.NewUpdateSystem(&splash_animation_sync_system)

	logo := chai.LoadPng("Assets/Chai_Logo.png")

	SplashSceen.NewEntity(chai.Vector2fZero, chai.NewVector2f(12.0, 8.0), 0.0)
	SplashSceen.WriteComponentToLastEntity(chai.SpriteComponent{Texture: logo, Tint: chai.WHITE})

	splash_anim := chai.NewAnimationComponentFloat32()
	splash_anim.NewTweenAnimationFloat32("Fade", false)
	splash_anim.RegisterKeyframe("Fade", 0.0, 0.0)
	splash_anim.RegisterKeyframe("Fade", 3.0, 1.0)
	splash_anim.RegisterKeyframe("Fade", 6.0, 0.0)

	splash_anim.Play("Fade")

	SplashSceen.WriteComponentToLastEntity(splash_anim)

}

type SplashAnimationSyncSystem struct {
	chai.EcsSystemImpl
}

func (sa *SplashAnimationSyncSystem) Update(dt float32) {
	chai.EachEntity(chai.AnimationComponent[float32]{}, func(entity *chai.EcsEntity, a interface{}) {
		anim := a.(chai.AnimationComponent[float32])
		var spriteC chai.SpriteComponent
		chai.ReadComponent(sa.GetEcsEngine(), entity, &spriteC)
		spriteC.Tint.SetColorAFloat32(anim.GetCurrentValue("Fade"))
		if anim.HasFinished("Fade") {
			chai.ChangeScene(&SceneOne)
		}
		chai.WriteComponent(sa.GetEcsEngine(), entity, spriteC)
	})
}
