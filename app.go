package main

import (
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	_ "runtime/pprof"
	"time"
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

const LOCAL = true

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

type action struct {
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
	if LOCAL {
		mainLocal()
	} else {
		mainCG()
	}
}

func mainLocal() {
	// start profiling

	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}

	// stop after 10 seconds
	time.AfterFunc(5*time.Second, func() {
		println("stopping profiling after 5 seconds")
		pprof.StopCPUProfile()
		f.Close()
	})

	state := state{
		playersPosition: [2]coord{{0, 4}, {8, 4}},
		boardRemoved:    [HEIGHT][WIDTH]bool{},
	}

	bestMove, bestScore := findBestMove(state, 0)

	debugAny("best move", bestMove)
	debugAny("best score", bestScore)

}

func mainCG() {
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

		bestAction, bestScore := findBestMove(currentState, myPlayerId)

		debugAny("best action", bestAction)
		debugAny("best score", bestScore)

		currentState = applyAction(currentState, bestAction, myPlayerId)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(fmt.Sprintf("%d %d %d %d", bestAction.movePosition.x, bestAction.movePosition.y, bestAction.removeTile.x, bestAction.removeTile.y)) // action: "x y" to action or "x y message" to action and speak
	}
}

func findBestMove(currentState state, myPlayerId int) (bestAction action, bestScore int) {
	bestScore = -1000000

	possibleActions := getPossibleActions(currentState, myPlayerId)

	//debugAny("possible moves", possibleActions)

	for iAction, action := range possibleActions {
		debugAny(fmt.Sprintf("testing action %d/%d", iAction, len(possibleActions)), action)

		nextState := applyAction(currentState, action, myPlayerId)

		score := alphaBeta(nextState, 2, -1000000, 1000000, myPlayerId, 1-myPlayerId)
		if score > bestScore {
			bestScore = score
			bestAction = action

			debugAny("found a better action", bestAction)
			debugAny("found a better score", bestScore)
		}
	}

	return
}

func applyMove(currentState state, movePosition coord, playerId int) (nextState state) {
	nextState = currentState
	nextState.playersPosition[playerId] = movePosition
	return
}

func applyAction(state state, action action, playerId int) (nextState state) {
	nextState = applyMove(state, action.movePosition, playerId)
	nextState.boardRemoved[action.removeTile.y][action.removeTile.x] = true
	return
}

func getPossibleActions(currentState state, playerId int) (possibleActions []action) {
	myPosition := currentState.playersPosition[playerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range adjacentTiles {
		if !isTileOccupied(currentState, adjacentTile) && !isTileRemoved(currentState, adjacentTile) {
			nextState := applyMove(currentState, adjacentTile, playerId)

			//debugAny(fmt.Sprintf("next state for %v", adjacentTile), nextState)

			possibleRemoves := getPossibleRemoves(nextState)

			//// sort possible removes by distance to opponent
			//oppoPosition := nextState.playersPosition[1-playerId]
			//
			//sort.Slice(possibleRemoves, func(i, j int) bool {
			//	return distance(possibleRemoves[i], oppoPosition) < distance(possibleRemoves[j], oppoPosition)
			//})

			//debugAny(fmt.Sprintf("possible removes for %v", adjacentTile), possibleRemoves)

			for _, possibleRemove := range possibleRemoves {
				possibleActions = append(possibleActions, action{adjacentTile, possibleRemove})
			}
		}
	}

	return
}

func distance(coord1 coord, coord2 coord) int {
	return int(math.Abs(float64(coord1.x-coord2.x)) + math.Abs(float64(coord1.y-coord2.y)))
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

func getPossibleRemoves(currentState state) (possibleRemoves []coord) {
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
	adjacentTiles = make([]coord, 0, 8)

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

	distanceFromPlayer := make([][][2]int, HEIGHT)
	for y := 0; y < HEIGHT; y++ {
		distanceFromPlayer[y] = make([][2]int, WIDTH)
		for x := 0; x < WIDTH; x++ {
			distanceFromPlayer[y][x] = [2]int{-1, -1}
		}
	}

	// for each player, find the distance to each tile using BFS
	for playerId := 0; playerId < 2; playerId++ {
		queue := []coord{currentState.playersPosition[playerId]}
		distanceFromPlayer[currentState.playersPosition[playerId].y][currentState.playersPosition[playerId].x][playerId] = 0

		for len(queue) > 0 {
			currentPosition := queue[0]
			queue = queue[1:]

			// for each adjacent tile, if it is not occupied and not already visited, add it to the queue
			adjacentTiles := getAdjacentTiles(currentPosition)
			for iAdjacentTile := 0; iAdjacentTile < len(adjacentTiles); iAdjacentTile++ {
				if !isTileOccupied(currentState, adjacentTiles[iAdjacentTile]) && !isTileRemoved(currentState, adjacentTiles[iAdjacentTile]) && distanceFromPlayer[adjacentTiles[iAdjacentTile].y][adjacentTiles[iAdjacentTile].x][playerId] == -1 {
					distanceFromPlayer[adjacentTiles[iAdjacentTile].y][adjacentTiles[iAdjacentTile].x][playerId] = distanceFromPlayer[currentPosition.y][currentPosition.x][playerId] + 1
					queue = append(queue, adjacentTiles[iAdjacentTile])
				}
			}
		}
	}

	// for each tile, find the closest player
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if distanceFromPlayer[y][x][0] == -1 && distanceFromPlayer[y][x][1] == -1 {
				// if the tile is not reachable by any player, it is not part of the partition
				continue
			}

			if distanceFromPlayer[y][x][0] == -1 && distanceFromPlayer[y][x][1] != -1 {
				partition[y][x] = 1
			} else if distanceFromPlayer[y][x][0] != -1 && distanceFromPlayer[y][x][1] == -1 {
				partition[y][x] = 0
			} else if distanceFromPlayer[y][x][0] < distanceFromPlayer[y][x][1] {
				partition[y][x] = 0
			} else if distanceFromPlayer[y][x][0] > distanceFromPlayer[y][x][1] {
				partition[y][x] = 1
			}
		}
	}

	//log the grid of the partition
	//debug("partition")

	//for y := 0; y < HEIGHT; y++ {
	//	line := ""
	//	for x := 0; x < WIDTH; x++ {
	//		// for each cell, padding of 2 characters
	//
	//		if currentState.playersPosition[0] == (coord{x, y}) {
	//			line += "A"
	//		} else if currentState.playersPosition[1] == (coord{x, y}) {
	//			line += "B"
	//		} else if currentState.boardRemoved[y][x] {
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

func getScore(currentState state, myPlayerId int) int {
	myPossibleActions := getPossibleActions(currentState, myPlayerId)
	opponentPossibleActions := getPossibleActions(currentState, 1-myPlayerId)

	// a good action is a action that maximize my player closest coords and minimize opponent closest coords
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
	if len(opponentPossibleActions) == 0 {
		bonusEnd += 1000
	}

	if len(myPossibleActions) == 0 {
		bonusEnd -= 500
	}

	return bonusEnd + myPlayerCellsCount - opponentCellsCount + len(myPossibleActions) - len(opponentPossibleActions)
}

// a minimax algorithm with alpha-beta pruning and negamax
func alphaBeta(currentState state, depth int, alpha int, beta int, myPlayerId int, playerId int) (nodeScore int) {
	if depth == 0 {
		// we reached the end of the tree, we return the score of the current state
		return getScore(currentState, myPlayerId)
	}

	// we get all the possible moves
	possibleActions := getPossibleActions(currentState, playerId)

	// if there is no possible action, the game is over, we return the score of the current state
	if len(possibleActions) == 0 {
		return getScore(currentState, myPlayerId)
	}

	nodeScore = -1000000

	// for each possible action
	for _, possibleAction := range possibleActions {
		// we get the score of the action by calling alphaBeta recursively
		nextState := applyAction(currentState, possibleAction, playerId)

		// we get the score of the action by calling alphaBeta recursively
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
