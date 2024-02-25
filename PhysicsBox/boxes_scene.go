package main

import (
	"math/rand"

	chai "github.com/mhamedGd/chai"
)

var SceneOne chai.Scene

var rectDrawingSystem chai.FillRectRenderSystem = chai.FillRectRenderSystem{}

var dynamicBodiesSystem chai.DynamicBodyUpdateSystem = chai.DynamicBodyUpdateSystem{}
var dragToMouseSystem DragToMouseSystem = DragToMouseSystem{}

const BORDER_THICKNESS = 1.0

func OnSceneOneStart() {
	SceneOne.Background = chai.NewRGBA8(240, 240, 240, 255)
	rectDrawingSystem.Shapes = &chai.Shapes
	rectDrawingSystem.Shapes.LineWidth = 0.15

	SceneOne.NewRenderSystem(&rectDrawingSystem)

	SceneOne.NewUpdateSystem(&dynamicBodiesSystem)
	SceneOne.NewUpdateSystem(&dragToMouseSystem)

	for i := 0; i < 30; i++ {
		randCoords := chai.NewVector2f(((rand.Float32()-0.5)*2.0)*14, ((rand.Float32()-0.5)*2.0)*6)
		randDims := chai.NewVector2f(rand.Float32()*5, rand.Float32()*5)
		randDims.X = chai.ClampFloat32(randDims.X, 1.5, 10.0)
		randDims.Y = chai.ClampFloat32(randDims.X, 1.5, 10.0)
		CreateBox(&SceneOne, randCoords, randDims, 30.0, chai.NewRGBA8(70, 70, 70, 255))

	}
	// CreateBox(&SceneOne, chai.Vector2fZero, chai.NewVector2f(4.0, 2.5), 30.0, chai.NewRGBA8(70, 70, 70, 255))
	// CreateBox(&SceneOne, chai.Vector2fZero.AddXY(2.0, 0.0), chai.NewVector2f(4.0, 2.5), 30.0, chai.NewRGBA8(70, 70, 70, 255))
	// CreateBox(&SceneOne, chai.Vector2fZero.AddXY(-2.0, 0.0), chai.NewVector2f(4.0, 2.5), 30.0, chai.NewRGBA8(70, 70, 70, 255))
	// CreateBox(&SceneOne, chai.Vector2fZero.AddXY(0.0, 2.0), chai.NewVector2f(4.0, 2.5), 30.0, chai.NewRGBA8(70, 70, 70, 255))

	ent := SceneOne.NewEntity(chai.NewVector2f(0.0, -float32(game.Height)/(2.0*WORLD_SCALING)), chai.NewVector2f(float32(game.Width)/WORLD_SCALING, BORDER_THICKNESS), 0.0)
	SceneOne.WriteComponentToLastEntity(chai.FillRectRenderComponent{Tint: chai.NewRGBA8(20, 40, 70, 255)})
	SceneOne.WriteComponentToLastEntity(chai.NewStaticBody(ent, chai.Shape_RectCollider, ent.Dimensions, 5, chai.GetPhysicsWorld()))

	ent = SceneOne.NewEntity(chai.NewVector2f(0.0, float32(game.Height)/(2.0*WORLD_SCALING)), chai.NewVector2f(float32(game.Width)/WORLD_SCALING, BORDER_THICKNESS), 0.0)
	SceneOne.WriteComponentToLastEntity(chai.FillRectRenderComponent{Tint: chai.NewRGBA8(20, 40, 70, 255)})
	SceneOne.WriteComponentToLastEntity(chai.NewStaticBody(ent, chai.Shape_RectCollider, ent.Dimensions, 5, chai.GetPhysicsWorld()))

	ent = SceneOne.NewEntity(chai.NewVector2f(float32(game.Width)/(2.0*WORLD_SCALING), 0.0), chai.NewVector2f(BORDER_THICKNESS, float32(game.Height)/WORLD_SCALING), 0.0)
	SceneOne.WriteComponentToLastEntity(chai.FillRectRenderComponent{Tint: chai.NewRGBA8(20, 40, 70, 255)})
	SceneOne.WriteComponentToLastEntity(chai.NewStaticBody(ent, chai.Shape_RectCollider, ent.Dimensions, 5, chai.GetPhysicsWorld()))

	ent = SceneOne.NewEntity(chai.NewVector2f(-float32(game.Width)/(2.0*WORLD_SCALING), 0.0), chai.NewVector2f(BORDER_THICKNESS, float32(game.Height)/WORLD_SCALING), 0.0)
	SceneOne.WriteComponentToLastEntity(chai.FillRectRenderComponent{Tint: chai.NewRGBA8(20, 40, 70, 255)})
	SceneOne.WriteComponentToLastEntity(chai.NewStaticBody(ent, chai.Shape_RectCollider, ent.Dimensions, 5, chai.GetPhysicsWorld()))

}

func CreateBox(scene *chai.Scene, pos, dims chai.Vector2f, rot float32, tint chai.RGBA8) {
	ent := scene.NewEntity(pos, dims, rot)
	scene.WriteComponentToLastEntity(chai.FillRectRenderComponent{Tint: chai.GetRandomRGBA8()})
	scene.WriteComponentToLastEntity(chai.NewDynamicBody(ent, chai.Shape_RectCollider, ent.Dimensions, 1.0*dims.LengthSquared(), 0.35, 0.2, 1.0, chai.GetPhysicsWorld()))
	scene.WriteComponentToLastEntity(DragToMouseComponent{})
}

type DragToMouseComponent struct {
	chai.Component
	entityOffset chai.Vector2f
}

func (t *DragToMouseComponent) ComponentSet(val interface{}) { *t = val.(DragToMouseComponent) }

type DragToMouseSystem struct {
	chai.EcsSystemImpl
	justPressed bool

	isDragging  bool
	draggedBody *chai.PhysicsBody
}

func (mSys *DragToMouseSystem) Update(dt float32) {
	chai.EachEntity(DragToMouseComponent{}, func(entity *chai.EcsEntity, a interface{}) {
		dynamic := chai.DynamicBodyComponent{}
		chai.ReadComponent(mSys.GetEcsEngine(), entity, &dynamic)

		dragComp := a.(DragToMouseComponent)
		if chai.IsMousePressed(chai.LEFT_MOUSE_BUTTON) || chai.GetNumberOfFingersTouching() > 0 {
			if !mSys.justPressed {
				body, ok := chai.OverlapBox(chai.GetMouseWorldPosition().AddXY(-0.2, -0.2), chai.GetMouseWorldPosition().AddXY(0.2, 0.2))
				if !ok {
					return
				}

				if body != dynamic.GetPhyiscsBody() {
					return
				}
				dragComp.entityOffset = (entity.Pos.Subtract(chai.GetMouseWorldPosition()))
				mSys.justPressed = true
				mSys.isDragging = true
				mSys.draggedBody = dynamic.GetPhyiscsBody()
			}
			if mSys.isDragging && mSys.draggedBody == dynamic.GetPhyiscsBody() {
				targetPoint := chai.GetMouseWorldPosition().Add(dragComp.entityOffset)
				dynamic.ApplyForce(targetPoint.Subtract(entity.Pos).Scale(1150.0))
				chai.Shapes.DrawLine(entity.Pos, chai.GetMouseWorldPosition(), chai.NewRGBA8(0, 0, 0, 255))
			}
		} else {
			mSys.justPressed = false
			mSys.isDragging = false
			mSys.draggedBody = nil

		}

		chai.WriteComponent(mSys.GetEcsEngine(), entity, dragComp)

	})
}
