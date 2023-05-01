package main

import (
	"fmt"
	"os"
)

/**
 *  	The Goal
The goal of the game is to block the opponent's pawn.

 	Rules
Board:
The game is played on a 9 x 9 boardRemoved.

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

// constant values
const WIDTH = 9
const HEIGHT = 9

type coord struct {
	x int
	y int
}

type state struct {
	playersPosition [2]coord
	boardRemoved    [HEIGHT][WIDTH]bool
}

type move struct {
	movePosition coord
	removeTile   coord
}

func debug(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func debugAny(message string, any interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", message, any)
}

func main() {
	var playerPositionX int
	fmt.Scan(&playerPositionX)

	// playerPositionY: player's coordinates.
	var playerPositionY int
	fmt.Scan(&playerPositionY)

	playerPosition := coord{playerPositionX, playerPositionY}

	myPlayerId := 0

	if playerPositionY == 0 {
		myPlayerId = 1
	}

	opponentPosition := coord{8, 4}

	if myPlayerId == 1 {
		opponentPosition = coord{0, 4}
	}

	currentState := state{
		playersPosition: [2]coord{playerPosition, opponentPosition},
		boardRemoved:    [HEIGHT][WIDTH]bool{},
	}

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

		if opponentLastRemovedTileX != -1 && opponentLastRemovedTileY != -1 {
			currentState.boardRemoved[opponentLastRemovedTileY][opponentLastRemovedTileX] = true
		}

		currentState.playersPosition[1-myPlayerId] = coord{opponentPositionX, opponentPositionY}

		debugAny("current state", currentState)

		bestMove, bestScore := findBestMove(currentState, myPlayerId)

		debugAny("best move", bestMove)
		debugAny("best score", bestScore)

		currentState.playersPosition[myPlayerId] = bestMove.movePosition

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(fmt.Sprintf("%d %d %d %d", bestMove.movePosition.x, bestMove.movePosition.y, bestMove.removeTile.x, bestMove.removeTile.y)) // action: "x y" to move or "x y message" to move and speak
	}
}

func findBestMove(currentState state, myPlayerId int) (bestMove move, bestScore int) {
	bestScore = -1000

	possibleMoves := getPossibleMoves(currentState, myPlayerId)

	for _, move := range possibleMoves {
		score := getScore(currentState, move, myPlayerId)
		if score > bestScore {
			bestScore = score
			bestMove = move

			debugAny("found a better move", bestMove)
			debugAny("found a better score", bestScore)
		}
	}

	return
}

func applyMove(currentState state, movePosition coord, myPlayerId int) (nextState state) {
	nextState = currentState

	nextState.playersPosition[myPlayerId] = movePosition

	return
}

func getPossibleMoves(currentState state, myPlayerId int) (possibleMoves []move) {
	myPosition := currentState.playersPosition[myPlayerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range adjacentTiles {
		if !isTileOccupied(currentState, adjacentTile) && !isTileRemoved(currentState, adjacentTile) {
			nextState := applyMove(currentState, adjacentTile, myPlayerId)

			possibleRemoves := getPossibleRemoves(nextState, myPlayerId)

			for _, possibleRemove := range possibleRemoves {
				possibleMoves = append(possibleMoves, move{adjacentTile, possibleRemove})
			}
		}
	}

	return
}

func getPossibleRemoves(currentState state, myPlayerId int) (possibleRemoves []coord) {
	// a player can remove any tile that is not occupied by a pawn and not already removed
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if !isTileOccupied(currentState, coord{x, y}) && !isTileRemoved(currentState, coord{x, y}) {
				possibleRemoves = append(possibleRemoves, coord{x, y})
			}
		}
	}

	return
}

func getAdjacentTiles(position coord) (adjacentTiles []coord) {
	coords := []coord{
		{position.x - 1, position.y - 1},
		{position.x - 1, position.y},
		{position.x - 1, position.y + 1},
		{position.x, position.y - 1},
		{position.x, position.y + 1},
		{position.x + 1, position.y - 1},
		{position.x + 1, position.y},
		{position.x + 1, position.y + 1},
	}

	for _, coord := range coords {
		if coord.x >= 0 && coord.x < WIDTH && coord.y >= 0 && coord.y < HEIGHT {
			adjacentTiles = append(adjacentTiles, coord)
		}
	}

	return
}

func isTileOccupied(currentState state, position coord) bool {
	for _, playerPosition := range currentState.playersPosition {
		if playerPosition == position {
			return true
		}
	}

	return false
}

func isTileRemoved(currentState state, position coord) bool {
	return currentState.boardRemoved[position.y][position.x]
}

func getScore(currentState state, move move, myPlayerId int) int {
	// a good move is a move that maximize my player possible moves and minimize the opponent possible moves

	nextState := applyMove(currentState, move.movePosition, myPlayerId)
	nextState.boardRemoved[move.removeTile.y][move.removeTile.x] = true

	myPossibleMoves := getPossibleMoves(nextState, myPlayerId)
	opponentPossibleMoves := getPossibleMoves(nextState, 1-myPlayerId)

	return len(myPossibleMoves) - len(opponentPossibleMoves)
}
