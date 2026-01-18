package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Vector struct {
	X float64
	Y float64
}

// calculate the length of the vector using the pythagorean theorem
func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// returns a normalized version of the vector (length = 1)
func (v Vector) Normalize() Vector {
	length := v.Length()
	return Vector{v.X / length, v.Y / length}
}

// returns a scaled version of the vector
func (v Vector) Scale(xScale, yScale float64) Vector {
	return Vector{v.X * xScale, v.Y * yScale}
}

// returns a translated version of the vector
func (v Vector) Translate(dx, dy float64) Vector {
	return Vector{v.X + dx, v.Y + dy}
}

type SpaceObject struct {
	name           string
	mass           float64       // mass of the object in kg
	position       Vector        // position vector of the object in m
	scaledPosition Vector        // scaled position vector of the object in pixel
	velocity       Vector        // velocity vector of the object in m/s
	img            *ebiten.Image // object image
	pathImg        *ebiten.Image // image of the object path
	color          color.Color   // color of object and object path
}

func (so *SpaceObject) UpdateVelocity(force Vector) {
	// Update velocity in each direction using Newtons 2nd Law of motion:
	// F = ma -> a = F/m
	// with a = dv/dt (acceleration is the derivate of velocity)
	// we get dv = F/m * dt
	// this way we can apply dv by adding it to the current velocity
	so.velocity.X += (force.X / so.mass) * dt
	so.velocity.Y += (force.Y / so.mass) * dt
}

func (so *SpaceObject) UpdatePosition() {
	// Update position using v = dx/dt (velocity is the derivate of distance)
	// v = dx/dt -> dx = v*dt
	// this way we can apply dx by adding it to the current position
	so.position.X += so.velocity.X * dt
	so.position.Y += so.velocity.Y * dt
}

func (so *SpaceObject) UpdatePathImage() {

	// fill the path pixels in the assigned color
	newImg := ebiten.NewImage(1, 1)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(so.scaledPosition.X, so.scaledPosition.Y)
	newImg.Fill(so.color)
	so.pathImg.DrawImage(newImg, op)
}

func CreateRandomSpaceObject() *SpaceObject {

	names := []string{
		"Mercury",
		"Venus",
		"Earth",
		"Mars",
		"Jupiter",
		"Saturn",
		"Uranus",
		"Neptune",
		"Pluto",
	}

	name := names[rand.Intn(len(names))]

	// generate a random mass
	mass := rand.Float64() * 6.417e25

	fmt.Println(mass)

	// generate a random starting position in [-1e8*Scale, 1e8*Scale]
	position := Vector{
		(rand.Float64()*2*1e8 - 1e8) * XScale,
		(rand.Float64()*2*1e8 - 1e8) * YScale,
	}

	fmt.Println(position)

	// generate a random starting velocity
	velocity := Vector{
		rand.Float64()*7000 - 3500,
		rand.Float64()*7000 - 3500,
	}

	// generate a random color for imgage and path
	r := uint8(rand.Int())
	g := uint8(rand.Int())
	b := uint8(rand.Int())
	color := color.RGBA{r, g, b, 1}

	return &SpaceObject{
		name:     name,
		mass:     mass,
		position: position,
		velocity: velocity,
		img:      createEmptyColoredImage(2, 2, color),
		color:    color,
	}
}

var (
	mplusFaceSource *text.GoTextFaceSource
)

const (
	gravitation float64 = 6.67430e-11                    // Gravitational constant (m^3 kg^-1 s^-2)
	dt          float64 = 1.0 / 60.0 * 60 * 60 * 24 * 30 // time delta (1 sec / refreshrate * seconds * minutes * hours)
	XScale      float64 = 0.1e-6                         // x scaling to show the huge numbers on screen
	YScale      float64 = 0.1e-6                         // y scaling to show the huge numbers on screen
)

type Game struct {
	screenWidth  int
	screenHeight int
	spaceObjects []*SpaceObject
	time         float64
}

func createEmptyColoredImage(width, height int, color color.Color) *ebiten.Image {
	// todo change with sprite
	img := ebiten.NewImage(width, height)
	img.Fill(color)
	return img
}

func calculateGravitationalForce(so1, so2 SpaceObject) Vector {
	// calculate distance vector and actual distance between so1 and so2
	// The vector points from so2 to  so1
	distanceVector := Vector{so1.position.X - so2.position.X, so1.position.Y - so2.position.Y}
	distance := distanceVector.Length()

	// calculate the gravitational force that is acting on so1
	gravForce := (gravitation * so2.mass * so1.mass) / (distance * distance)

	// Normalize the distance vector, so its length equals 1.
	// This gives us a vector that determines the direction of the gravitational force without
	// manipulating the magnitude.
	// The vector points from so2 to so1
	forceDirection := distanceVector.Normalize()

	// We use negative signs because the vector should be pointing towards so2,
	// but currently is pointing towards so1.
	// We multiply the gravitational force with the respective direction
	// to get the magnitude of the force in each direction.
	force := Vector{-gravForce * forceDirection.X, -gravForce * forceDirection.Y}

	return force
}

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}

func NewGame() *Game {

	game := &Game{
		spaceObjects: make([]*SpaceObject, 3),
		time:         0,
	}
	game.spaceObjects[0] = &SpaceObject{
		name:     "Earth",
		mass:     5.9722e24,
		position: Vector{0, 0},
		velocity: Vector{0, -20},
		img:      createEmptyColoredImage(2, 2, color.RGBA{255, 0, 0, 1}),
		color:    color.RGBA{255, 0, 0, 1},
	}
	game.spaceObjects[1] = &SpaceObject{
		name:     "Moon",
		mass:     5.9722e22,
		position: Vector{5e9, 0},
		velocity: Vector{0, -100},
		img:      createEmptyColoredImage(2, 2, color.RGBA{0, 255, 0, 1}),
		color:    color.RGBA{0, 255, 0, 1},
	}
	game.spaceObjects[2] = &SpaceObject{
		name:     "Spacecraft",
		mass:     5.9722e22,
		position: Vector{-5e9, 1e9},
		velocity: Vector{-10, 150},
		img:      createEmptyColoredImage(2, 2, color.RGBA{0, 0, 255, 1}),
		color:    color.RGBA{0, 0, 255, 1},
	}
	/*game.spaceObjects[0] = &SpaceObject{
		name:     "Mars",
		mass:     6.417e23,
		position: Vector{0, 0},
		velocity: Vector{0, -10},
		img:      createEmptyColoredImage(2, 2, color.RGBA{255, 0, 0, 1}),
		color:    color.RGBA{255, 0, 0, 1},
	}
	game.spaceObjects[1] = &SpaceObject{
		name:     "Spacecraft",
		mass:     815,
		position: Vector{1e8, 1e8},
		velocity: Vector{0, -700},
		img:      createEmptyColoredImage(2, 2, color.White),
		color:    color.White,
	}*/
	//game.spaceObjects[0] = CreateRandomSpaceObject()
	//game.spaceObjects[1] = CreateRandomSpaceObject()
	return game
}

func (g *Game) Update() error {

	// iterate over every spaceobject and calculate how it is influenced by all other objects
	for i, so1 := range g.spaceObjects {
		for j, so2 := range g.spaceObjects {

			// skip if we would compare the same object
			if j == i {
				continue
			}

			// calculate the force so2 is putting on so1 and update the velocity this force applies
			force := calculateGravitationalForce(*so1, *so2)
			so1.UpdateVelocity(force)

		}
	}

	// after updating the velocites, we now update all positions
	for _, so := range g.spaceObjects {
		// Update position using the spaceobjects velocity
		so.UpdatePosition()
		// scale current postion to window
		so.scaledPosition = so.position.Scale(XScale, YScale).Translate(float64(g.screenWidth/2.0), float64(g.screenHeight/2.0))
	}

	g.time += dt
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// iterate over every spaceobject and draw it
	for _, so := range g.spaceObjects {

		// temp
		if g.time <= dt {
			so.pathImg = ebiten.NewImageFromImage(screen)
		}

		// draw image on screen
		soImgOptions := &ebiten.DrawImageOptions{}
		soImgOptions.GeoM.Translate(so.scaledPosition.X, so.scaledPosition.Y)
		screen.DrawImage(so.img, soImgOptions)

		// update so internal path image and draw it on screen
		so.UpdatePathImage()
		screen.DrawImage(so.pathImg, nil)

		fmt.Printf("SO: %s, Position: (%.2f, %.2f), Velocity: (%.2f, %.2f)\n", so.name, so.scaledPosition.X, so.scaledPosition.Y, so.velocity.X, so.velocity.Y)
	}

	size := 12.0

	str := strconv.FormatFloat(g.spaceObjects[0].velocity.Length(), 'f', 2, 64)

	textOp := &text.DrawOptions{}
	//textOp.GeoM.Translate(float64(x)+float64(tileSize)/2, float64(y)+float64(tileSize)/2)
	//textOp.ColorScale.ScaleWithColor(tileColor(v))
	//textOp.PrimaryAlign = text.AlignCenter
	//textOp.SecondaryAlign = text.AlignCenter
	text.Draw(screen, str, &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   size,
	}, textOp)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowSize(1080, 720)
	ebiten.SetWindowTitle("swingby")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
