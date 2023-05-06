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

		currentState = applyMove(currentState, bestMove, myPlayerId)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(fmt.Sprintf("%d %d %d %d", bestMove.movePosition.x, bestMove.movePosition.y, bestMove.removeTile.x, bestMove.removeTile.y)) // action: "x y" to move or "x y message" to move and speak
	}
}

func findBestMove(currentState state, myPlayerId int) (bestMove move, bestScore int) {
	bestScore = -1000000

	possibleMoves := getPossibleMoves(currentState, myPlayerId)

	//debugAny("possible moves", possibleMoves)

	for _, move := range possibleMoves {
		//debugAny(fmt.Sprintf("testing move %d", iMove), move)

		nextState := applyMove(currentState, move, myPlayerId)

		score := alphaBeta(nextState, 0, -1000000, 1000000, myPlayerId, 1-myPlayerId)
		if score > bestScore {
			bestScore = score
			bestMove = move

			debugAny("found a better move", bestMove)
			debugAny("found a better score", bestScore)
		}
	}

	return
}

func applyMoveOnly(currentState state, movePosition coord, playerId int) (nextState state) {
	nextState = currentState
	nextState.playersPosition[playerId] = movePosition
	return
}

func applyMove(state state, move move, playerId int) (nextState state) {
	nextState = applyMoveOnly(state, move.movePosition, playerId)
	nextState.boardRemoved[move.removeTile.y][move.removeTile.x] = true
	return
}

func getPossibleMoves(currentState state, myPlayerId int) (possibleMoves []move) {
	myPosition := currentState.playersPosition[myPlayerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range adjacentTiles {
		if !isTileOccupied(currentState, adjacentTile) && !isTileRemoved(currentState, adjacentTile) {
			nextState := applyMoveOnly(currentState, adjacentTile, myPlayerId)

			//debugAny(fmt.Sprintf("next state for %v", adjacentTile), nextState)

			possibleRemoves := getPossibleRemoves(nextState, myPlayerId)

			//debugAny(fmt.Sprintf("possible removes for %v", adjacentTile), possibleRemoves)

			for _, possibleRemove := range possibleRemoves {
				possibleMoves = append(possibleMoves, move{adjacentTile, possibleRemove})
			}
		}
	}

	return
}

//func getPossibleRemoves(currentState state, myPlayerId int) (possibleRemoves []coord) {
//	// a player can remove any tile that is not occupied by a pawn and not already removed
//	for y := 0; y < HEIGHT; y++ {
//		for x := 0; x < WIDTH; x++ {
//			if !isTileOccupied(currentState, coord{x, y}) && !isTileRemoved(currentState, coord{x, y}) {
//				possibleRemoves = append(possibleRemoves, coord{x, y})
//			}
//		}
//	}
//
//	return
//}

func getPossibleRemoves(currentState state, myPlayerId int) (possibleRemoves []coord) {
	// a player can remove any tile that is not occupied by a pawn and not already removed
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			c := coord{x, y}
			if !isTileOccupied(currentState, c) && !isTileRemoved(currentState, c) {
				possibleRemoves = append(possibleRemoves, c)
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

func getPartition(currentState state) (partition [][]int) {
	// we use a BFS to find all the tiles that are reachable from a player

	// initialize the partition to -1
	partition = make([][]int, HEIGHT)
	for y := 0; y < HEIGHT; y++ {
		partition[y] = make([]int, WIDTH)
		for x := 0; x < WIDTH; x++ {
			partition[y][x] = -1
		}
	}

	// initialize the queue with the players' positions
	queue := make([]coord, 0)
	for playerId, playerPosition := range currentState.playersPosition {
		partition[playerPosition.y][playerPosition.x] = playerId
		queue = append(queue, playerPosition)
	}

	// BFS
	for len(queue) > 0 {
		currentPosition := queue[0]
		queue = queue[1:]

		adjacentTiles := getAdjacentTiles(currentPosition)

		for _, adjacentTile := range adjacentTiles {
			if partition[adjacentTile.y][adjacentTile.x] == -1 && !isTileRemoved(currentState, adjacentTile) {
				partition[adjacentTile.y][adjacentTile.x] = partition[currentPosition.y][currentPosition.x]
				queue = append(queue, adjacentTile)
			}
		}
	}

	// log the grid of the partition
	//debug("partition")
	//
	//for y := 0; y < HEIGHT; y++ {
	//	line := ""
	//	for x := 0; x < WIDTH; x++ {
	//		// for each cell, padding of 2 characters
	//
	//		if currentState.boardRemoved[y][x] {
	//			line += "X"
	//		} else if partition[y][x] == -1 {
	//			line += "."
	//		} else {
	//			line += strconv.Itoa(partition[y][x])
	//		}
	//	}
	//	debug(line)
	//}

	return
}

func contains(coords []coord, coord coord) bool {
	for _, c := range coords {
		if c == coord {
			return true
		}
	}

	return false
}

func getScore(currentState state, myPlayerId int) int {
	myPossibleMoves := getPossibleMoves(currentState, myPlayerId)
	opponentPossibleMoves := getPossibleMoves(currentState, 1-myPlayerId)

	// a good move is a move that maximize my player closest coords and minimize opponent closest coords
	partition := getPartition(currentState)

	myPlayerCellsCount := 0
	opponentCellsCount := 0

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if partition[y][x] == myPlayerId {
				myPlayerCellsCount++
			} else if partition[y][x] == 1-myPlayerId {
				opponentCellsCount++
			}
		}
	}

	bonusEnd := 0
	if len(opponentPossibleMoves) == 0 {
		bonusEnd += 1000
	}

	if len(myPossibleMoves) == 0 {
		bonusEnd -= 500
	}

	return bonusEnd + myPlayerCellsCount - opponentCellsCount + len(myPossibleMoves) - len(opponentPossibleMoves)
}

// a minimax algorithm with alpha-beta pruning and negamax
func alphaBeta(currentState state, depth int, alpha int, beta int, myPlayerId int, playerId int) (nodeScore int) {
	if depth == 0 {
		// we reached the end of the tree, we return the score of the current state
		return getScore(currentState, myPlayerId)
	}

	// we get all the possible moves
	possibleMoves := getPossibleMoves(currentState, playerId)

	// if there is no possible move, the game is over, we return the score of the current state
	if len(possibleMoves) == 0 {
		return getScore(currentState, myPlayerId)
	}

	nodeScore = -1000000

	// for each possible move
	for _, possibleMove := range possibleMoves {
		// we get the score of the move by calling alphaBeta recursively
		nextState := applyMove(currentState, possibleMove, playerId)

		// we get the score of the move by calling alphaBeta recursively
		nodeScore = max(nodeScore, -alphaBeta(nextState, depth-1, -beta, -alpha, myPlayerId, 1-playerId))

		if nodeScore >= beta {
			return nodeScore
		}

		alpha = max(alpha, nodeScore)
	}

	return nodeScore
}

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}
