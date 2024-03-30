package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
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

type Planet struct {
	mass     float64       // mass of the Planet in kg
	position Vector        // position vector of the planet in m
	velocity Vector        // velocity vector of the planet in m/s
	img      *ebiten.Image // planet image
	pathImg  *ebiten.Image // image of the planets path
}

type Spacecraft struct {
	mass     float64       // mass of the spacecraft in kg
	position Vector        // position vector of the spacecraft in m
	velocity Vector        // velocity vector of the spacecraft in m/s
	img      *ebiten.Image // spacecraft image
	pathImg  *ebiten.Image // image of the spacecrafts path
}

const (
	gravitation float64 = 6.67430e-11               // Gravitational constant (m^3 kg^-1 s^-2)
	dt          float64 = 1.0 / 60.0 * 60 * 60 * 24 // time delta (1 sec / refreshrate * seconds * minutes * hours)
	XScale      float64 = 0.1e-5                    // x scaling to show the huge numbers on screen
	YScale      float64 = 0.1e-5                    // y scaling to show the huge numbers on screen
)

type Game struct {
	screenWidth  int
	screenHeight int
	planet       Planet
	spacecraft   Spacecraft
	time         float64
}

func NewGame() *Game {
	// todo change with sprite
	img := ebiten.NewImage(2, 2)
	img.Fill(color.White)

	return &Game{
		planet: Planet{
			mass:     6.417e23,
			position: Vector{0, 0},
			velocity: Vector{0, -10},
			img:      img,
		},
		spacecraft: Spacecraft{
			mass:     815,
			position: Vector{1e8, 1e8},
			velocity: Vector{0, -700},
			img:      img,
		},
		time: 0,
	}
}

func (g *Game) Update() error {

	// update position of planet
	g.planet.position.X += g.planet.velocity.X * dt
	g.planet.position.Y += g.planet.velocity.Y * dt

	// get position of planet
	planetPos := g.planet.position

	// get position and velocity of spacecraft
	pos := &g.spacecraft.position
	vel := &g.spacecraft.velocity

	// calculate distance vector and actual distance between spacecraft and planet
	distanceVector := Vector{pos.X - planetPos.X, pos.Y - planetPos.Y}
	distance := distanceVector.Length()

	// calculate the gravitational force that is acting on the spacecraft
	gravForce := (gravitation * g.planet.mass * g.spacecraft.mass) / (distance * distance)

	// Normalize the distance vector, so its length equals 1.
	// This gives us a vector that determines the direction of the gravitational force without
	// manipulating the magnitude.
	// The vector points from the star to the spacecraft
	forceDirection := distanceVector.Normalize()

	// We use negative signs because the vector should be pointing towards the planet,
	// but currently is pointing towards the spaceraft.
	// We multiply the gravitational force with the respective direction
	// to get the magnitude of the force in each direction.
	force := Vector{-gravForce * forceDirection.X, -gravForce * forceDirection.Y}

	// Update velocity in each direction using Newtons 2nd Law of motion:
	// F = ma -> a = F/m
	// with a = dv/dt (acceleration is the derivate of velocity)
	// we get dv = F/m * dt
	// this way we can apply dv by adding it to the current velocity
	vel.X += (force.X / g.spacecraft.mass) * dt
	vel.Y += (force.Y / g.spacecraft.mass) * dt

	// Update position using v = dx/dt (velocity is the derivate of distance)
	// v = dx/dt -> dx = v*dt
	// this way we can apply dx by adding it to the current position
	pos.X += vel.X * dt
	pos.Y += vel.Y * dt

	g.time += dt

	return nil
}

func (g *Game) FillPathPixel(x, y float64) {

	newImg := ebiten.NewImage(1, 1)

	pathImg := g.spacecraft.pathImg

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	newImg.Fill(color.White)
	pathImg.DrawImage(newImg, op)
}

func (g *Game) Draw(screen *ebiten.Image) {

	if g.time <= dt {
		g.spacecraft.pathImg = ebiten.NewImageFromImage(screen)
		g.planet.pathImg = ebiten.NewImageFromImage(screen)
	}

	// draw spacecraft
	spacecraftPos := g.spacecraft.position
	//spacecraftVel := g.spacecraft.velocity

	scaledSpacecraftPos := spacecraftPos.Scale(XScale, YScale).Translate(float64(g.screenWidth/2.0), float64(g.screenHeight/2.0))

	spacecraftOptions := &ebiten.DrawImageOptions{}
	spacecraftOptions.GeoM.Translate(scaledSpacecraftPos.X, scaledSpacecraftPos.Y)
	screen.DrawImage(g.spacecraft.img, spacecraftOptions)

	//fmt.Printf("Time: %.2f, Position: (%.2f, %.2f), Velocity: (%.2f, %.2f)\n", g.time, scaledSpacecraftPos.X, scaledSpacecraftPos.Y, spacecraftVel.X, spacecraftVel.Y)

	// draw spacecraft path
	g.FillPathPixel(scaledSpacecraftPos.X, scaledSpacecraftPos.Y)
	screen.DrawImage(g.spacecraft.pathImg, nil)

	// draw planet
	planetPos := g.planet.position
	//planetVel := g.planet.velocity

	scaledPlanetPos := planetPos.Scale(XScale, YScale).Translate(float64(g.screenWidth/2.0), float64(g.screenHeight/2.0))

	planetOptions := &ebiten.DrawImageOptions{}
	planetOptions.GeoM.Translate(scaledPlanetPos.X, scaledPlanetPos.Y)
	screen.DrawImage(g.planet.img, planetOptions)

	//fmt.Printf("Time: %.2f, Position: (%.2f, %.2f), Velocity: (%.2f, %.2f)\n", g.time, scaledPlanetPos.X, scaledPlanetPos.Y, planetVel.X, planetVel.Y)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	//ebiten.SetWindowSize(screenWidth, screenHeight)
	//ebiten.SetWindowTitle("swingby")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
