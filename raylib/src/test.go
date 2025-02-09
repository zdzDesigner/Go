package main

import (
	"os"
	"runtime"

	"github.com/gen2brain/raylib-go/raylib"
)

// Game states
const (
	Logo = iota
	Title
	GamePlay
	Ending
)

func init() {
	rl.SetCallbackFunc(main)
}

func main() {
	screenWidth := int32(400)
	screenHeight := int32(850)

	rl.SetConfigFlags(rl.FlagVsyncHint)

	rl.InitWindow(screenWidth, screenHeight, "Android example")

	rl.InitAudioDevice()

	currentScreen := Logo
	windowShouldClose := false

	texture := rl.LoadTexture("raylib_logo.png") // Load texture (placed on assets folder)
	fx := rl.LoadSound("coin.wav")               // Load WAV audio file (placed on assets folder)
	ambient := rl.LoadMusicStream("ambient.ogg") // Load music

	rl.PlayMusicStream(ambient)

	framesCounter := 0 // Used to count frames

	// rl.SetTargetFPS(60)

	for !windowShouldClose {
		rl.UpdateMusicStream(ambient)

		if runtime.GOOS == "android" && rl.IsKeyDown(rl.KeyBack) || rl.WindowShouldClose() {
			windowShouldClose = true
		}

		switch currentScreen {
		case Logo:
			framesCounter++ // Count frames

			// Wait for 4 seconds (240 frames) before jumping to Title screen
			if framesCounter > 240 {
				currentScreen = Title
			}
		case Title:
			// Press enter to change to GamePlay screen
			if rl.IsGestureDetected(rl.GestureTap) {
				rl.PlaySound(fx)
				currentScreen = GamePlay
			}
		case GamePlay:
			// Press enter to change to Ending screen
			if rl.IsGestureDetected(rl.GestureTap) {
				rl.PlaySound(fx)
				currentScreen = Ending
			}
		case Ending:
			// Press enter to return to Title screen
			if rl.IsGestureDetected(rl.GestureTap) {
				rl.PlaySound(fx)
				currentScreen = Title
			}
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		// rl.DrawText("zdz", 0, 0, 60, rl.LightGray)

		// switch currentScreen {
		// case Logo:
		// 	rl.DrawText("LOGO SCREEN", 20, 20, 40, rl.LightGray)
		// 	rl.DrawTexture(texture, screenWidth/2-texture.Width/2, screenHeight/2-texture.Height/2, rl.White)
		// 	rl.DrawText("WAIT for 4 SECONDS...", 290, 400, 20, rl.Gray)
		// case Title:
		// 	rl.DrawRectangle(0, 0, screenWidth, screenHeight, rl.Green)
		// 	rl.DrawText("TITLE SCREEN", 20, 20, 40, rl.DarkGreen)
		// 	rl.DrawText("TAP SCREEN to JUMP to GAMEPLAY SCREEN", 160, 220, 20, rl.DarkGreen)
		// case GamePlay:
		// 	rl.DrawRectangle(0, 0, screenWidth, screenHeight, rl.Purple)
		// 	rl.DrawText("GAMEPLAY SCREEN", 20, 20, 40, rl.Maroon)
		// 	rl.DrawText("TAP SCREEN to JUMP to ENDING SCREEN", 170, 220, 20, rl.Maroon)
		// case Ending:
		// 	rl.DrawRectangle(0, 0, screenWidth, screenHeight, rl.Blue)
		// 	rl.DrawText("ENDING SCREEN", 20, 20, 40, rl.DarkBlue)
		// 	rl.DrawText("TAP SCREEN to RETURN to TITLE SCREEN", 160, 220, 20, rl.DarkBlue)
		// default:
		// }
		rl.DrawCircleV(rl.NewVector2(30, 200), 140, rl.Red)

		rl.EndDrawing()
	}

	rl.UnloadSound(fx)            // Unload sound data
	rl.UnloadMusicStream(ambient) // Unload music stream data
	rl.CloseAudioDevice()         // Close audio device (music streaming is automatically stopped)
	rl.UnloadTexture(texture)     // Unload texture data
	rl.CloseWindow()              // Close window

	os.Exit(0)
}
