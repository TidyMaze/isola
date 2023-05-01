package main

import "fmt"

/**
 *  	The Goal
The goal of the game is to block the opponent's pawn.

 	Rules
Board:
The game is played on a 9 x 9 board.

Player 0 always starts at (0, 4) and player 1 at (8, 4).

At each turn:
You must move your pawn to an adjacent tile (diagonal included) :

You can't :
stay put
move to the tile occupied by the opponent's pawn
move to an already removed tile.
Then you must remove a free tile.
You can't remove a tile occupied by a pawn.

Victory Conditions
At his turn, the opponent can't move his pawn.

Loss Conditions
At your turn, you can't move your pawn.
You do not respond in time or output an invalid action.

 	Game Input
Initialization input
Line 1: playerPositionX
Line 2: playerPositionY the coordinates of your pawn.

Input for one game turn
Line 1: opponentPositionX
Line 2: opponentPositionY the coordinates of the opponent's pawn.
Line 3: opponentLastRemovedTileX
Line 4: opponentLastRemovedTileY the coordinates of the last tile removed by the opponent ( -1 -1 if no tile has been removed (first round)).

Output
A single line containing the coordinates where you want to move your pawn, followed by the coordinates of the tile you want to remove.
Example: 1 4 7 4

You can also add a message :
Example: 1 4 7 4;MESSAGE
NB : You can print RANDOM instead of the 4 coordinates. Then a random possible move and tile will be chosen.

Constraints
Response time first turn is ≤ 1000 ms.
Response time per turn is ≤ 100 ms.
 **/

func main() {
	var playerPositionX int
	fmt.Scan(&playerPositionX)

	// playerPositionY: player's coordinates.
	var playerPositionY int
	fmt.Scan(&playerPositionY)

	for {
		var opponentPositionX int
		fmt.Scan(&opponentPositionX)

		// opponentPositionY: opponent's coordinates.
		var opponentPositionY int
		fmt.Scan(&opponentPositionY)

		var opponentLastRemovedTileX int
		fmt.Scan(&opponentLastRemovedTileX)

		// opponentLastRemovedTileY: coordinates of the last removed tile. (-1 -1) if no tile has been removed.
		var opponentLastRemovedTileY int
		fmt.Scan(&opponentLastRemovedTileY)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("RANDOM;MESSAGE") // Write action to stdout
	}
}
