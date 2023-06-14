package main

import (
	"fmt"
	"math"
	_ "net/http/pprof"
	"os"
	_ "runtime/pprof"
	"strconv"
	"strings"
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

var LOCAL = os.Getenv("LOCAL") == "true"

const MAX_DEPTH = 2

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
	turn            int
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
		println("local mode")
		mainLocal()
	} else {
		debug("cg mode")
		mainCG()
	}
}

func mainLocal() {
	// start profiling

	//f, err := os.Create("cpu.prof")
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = pprof.StartCPUProfile(f)
	//if err != nil {
	//	panic(err)
	//}

	// enable memory profiling

	//go func() {
	//	http.ListenAndServe(":6060", nil)
	//}()

	// stop after 10 seconds
	//time.AfterFunc(5*time.Second, func() {
	//	println("stopping profiling after 5 seconds")
	//	pprof.StopCPUProfile()
	//	f.Close()
	//})

	//defer pprof.StopCPUProfile()

	state := state{
		playersPosition: [2]coord{{0, 4}, {8, 4}},
		boardRemoved:    [HEIGHT][WIDTH]bool{},
		turn:            0,
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
		turn:            0,
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

func findBestMove(currentState state, myPlayerId int) (bestAction *action, bestScore int) {
	//debugAny("possible moves", possibleActions
	stateScoreCache = make(map[string]int)

	bestScore, bestAction = minimax(currentState, MAX_DEPTH, myPlayerId, true)

	return
}

func applyMove(currentState state, movePosition coord, playerId int) (nextState state) {
	nextState = currentState
	nextState.turn++
	nextState.playersPosition[playerId] = movePosition
	return
}

func applyAction(state state, action *action, playerId int) (nextState state) {
	nextState = applyMove(state, action.movePosition, playerId)
	nextState.boardRemoved[action.removeTile.y][action.removeTile.x] = true
	return
}

func getPossibleActions(currentState state, playerId int) []action {
	actions := make([]action, 0, 8*WIDTH*HEIGHT)

	myPosition := currentState.playersPosition[playerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range adjacentTiles {
		if !isTileOccupied(&currentState, &adjacentTile) && !isTileRemoved(&currentState, &adjacentTile) {
			nextState := applyMove(currentState, adjacentTile, playerId)

			//debugAny(fmt.Sprintf("next state for %v", adjacentTile), nextState)

			// a player can remove any tile that is not occupied by a pawn and not already removed
			//for y := 0; y < HEIGHT; y++ {
			//	for x := 0; x < WIDTH; x++ {
			//		c := coord{x, y}
			//		if !isTileOccupied(&nextState, &c) && !isTileRemoved(&nextState, &c) {
			//			actions = append(actions, action{adjacentTile, c})
			//		}
			//	}
			//}
			// it's better to remove the tile that is the closest to the opponent

			opponentPosition := currentState.playersPosition[1-playerId]
			adjacentTilesToOpponent := getAdjacentTiles(opponentPosition)
			for _, adjacentTileToOpponent := range adjacentTilesToOpponent {
				if !isTileOccupied(&nextState, &adjacentTileToOpponent) && !isTileRemoved(&nextState, &adjacentTileToOpponent) {
					actions = append(actions, action{adjacentTile, adjacentTileToOpponent})
				}
			}
		}
	}

	return actions
}

func getPossibleActionsCount(currentState state, playerId int) int {
	count := 0

	myPosition := currentState.playersPosition[playerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range adjacentTiles {
		if !isTileOccupied(&currentState, &adjacentTile) && !isTileRemoved(&currentState, &adjacentTile) {
			count++
		}
	}

	return count
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

//func getPossibleRemoves(currentState state, possibleRemoves *[]coord) {
//
//	// empty possible removes
//	*possibleRemoves = (*possibleRemoves)[:0]
//
//	// a player can remove any tile that is not occupied by a pawn and not already removed
//	for y := 0; y < HEIGHT; y++ {
//		for x := 0; x < WIDTH; x++ {
//			c := coord{x, y}
//			if !isTileOccupied(currentState, c) && !isTileRemoved(currentState, c) {
//				*possibleRemoves = append(*possibleRemoves, c)
//			}
//		}
//	}
//}

var cacheAdjacentTiles = make(map[int][]coord)

func getAdjacentTiles(position coord) (adjacentTiles []coord) {

	adjacentTiles, ok := cacheAdjacentTiles[position.y*WIDTH+position.x]
	if ok {
		return
	}

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

	for iCoord := 0; iCoord < len(coords); iCoord++ {
		if coords[iCoord].x >= 0 && coords[iCoord].x < WIDTH && coords[iCoord].y >= 0 && coords[iCoord].y < HEIGHT {
			adjacentTiles = append(adjacentTiles, coords[iCoord])
		}
	}

	cacheAdjacentTiles[position.y*WIDTH+position.x] = adjacentTiles

	return
}

func isTileOccupied(currentState *state, position *coord) bool {
	return currentState.playersPosition[0] == *position || currentState.playersPosition[1] == *position
}

func isTileRemoved(currentState *state, position *coord) bool {
	return currentState.boardRemoved[position.y][position.x]
}

var distanceFromPlayer [2][WIDTH * HEIGHT]int

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

	for i := 0; i < WIDTH*HEIGHT; i++ {
		distanceFromPlayer[0][i] = -1
		distanceFromPlayer[1][i] = -1
	}

	// for each player, find the distance to each tile using BFS
	for playerId := 0; playerId < 2; playerId++ {
		distanceFromPlayer[playerId][currentState.playersPosition[playerId].y*WIDTH+currentState.playersPosition[playerId].x] = 0

		queue := make([]coord, 0, WIDTH*HEIGHT)
		queue = append(queue, currentState.playersPosition[playerId])

		for len(queue) > 0 {
			currentPosition := queue[0]
			queue = queue[1:]

			// for each adjacent tile, if it is not occupied and not already visited, add it to the queue
			adjacentTiles := getAdjacentTiles(currentPosition)
			for iAdjacentTile := 0; iAdjacentTile < len(adjacentTiles); iAdjacentTile++ {
				adj := &adjacentTiles[iAdjacentTile]

				if !isTileOccupied(&currentState, adj) && !isTileRemoved(&currentState, adj) && distanceFromPlayer[playerId][adj.y*WIDTH+adj.x] == -1 {
					distanceFromPlayer[playerId][adj.y*WIDTH+adj.x] = distanceFromPlayer[playerId][currentPosition.y*WIDTH+currentPosition.x] + 1
					queue = append(queue, *adj)
				}
			}
		}
	}

	// for each tile, find the closest player
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if distanceFromPlayer[0][y*WIDTH+x] == -1 && distanceFromPlayer[1][y*WIDTH+x] == -1 {
				// if the tile is not reachable by any player, it is not part of the partition
				continue
			}

			if distanceFromPlayer[0][y*WIDTH+x] == -1 && distanceFromPlayer[1][y*WIDTH+x] != -1 {
				partition[y][x] = 1
			} else if distanceFromPlayer[0][y*WIDTH+x] != -1 && distanceFromPlayer[1][y*WIDTH+x] == -1 {
				partition[y][x] = 0
			} else if distanceFromPlayer[0][y*WIDTH+x] < distanceFromPlayer[1][y*WIDTH+x] {
				partition[y][x] = 0
			} else if distanceFromPlayer[0][y*WIDTH+x] > distanceFromPlayer[1][y*WIDTH+x] {
				partition[y][x] = 1
			}
		}
	}

	//log the grid of the partition
	//if LOCAL {
	//
	//	debug("partition")
	//
	//	for y := 0; y < HEIGHT; y++ {
	//		line := ""
	//		for x := 0; x < WIDTH; x++ {
	//			// for each cell, padding of 2 characters
	//
	//			if currentState.playersPosition[0] == (coord{x, y}) {
	//				line += "A"
	//			} else if currentState.playersPosition[1] == (coord{x, y}) {
	//				line += "B"
	//			} else if currentState.boardRemoved[y][x] {
	//				line += "X"
	//			} else if partition[y][x] == -1 {
	//				line += "."
	//			} else {
	//				line += strconv.Itoa(partition[y][x])
	//			}
	//		}
	//		debug(line)
	//	}
	//}

	return
}

func getScore(currentState state, myPlayerId int) int {
	myPossibleActions := getPossibleActionsCount(currentState, myPlayerId)
	opponentPossibleActions := getPossibleActionsCount(currentState, 1-myPlayerId)

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
	if opponentPossibleActions == 0 {
		bonusEnd += 1000000
		bonusEnd -= currentState.turn * 1000
	}

	if myPossibleActions == 0 {
		bonusEnd -= 1000000
		bonusEnd += currentState.turn * 1000
	}

	return bonusEnd + myPlayerCellsCount - opponentCellsCount + myPossibleActions - opponentPossibleActions
}

var stateScoreCache = make(map[string]int)

func hashState(currentState state) string {
	var strBuilder strings.Builder

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if currentState.boardRemoved[y][x] {
				strBuilder.WriteString("X")
			} else {
				strBuilder.WriteString(".")
			}
		}
	}

	for playerId := 0; playerId < 2; playerId++ {
		strBuilder.WriteString(strconv.Itoa(currentState.playersPosition[playerId].x))
		strBuilder.WriteString(strconv.Itoa(currentState.playersPosition[playerId].y))
	}

	return strBuilder.String()
}

func minimax(currentState state, depth int, myPlayerId int, maximizingPlayer bool) (bestMoveValue int, bestMove *action) {
	hashedState := hashState(currentState)

	playerId := 0
	if !maximizingPlayer {
		playerId = 1
	}

	if score, ok := stateScoreCache[hashedState]; ok {
		//debugAny(fmt.Sprintf("cache hit for %s", hashedState), score)
		return score, nil
	}

	if depth == 0 {
		res := getScore(currentState, myPlayerId)
		stateScoreCache[hashedState] = res
		//debugAny(fmt.Sprintf("cache miss for %s", hashedState), res)
		return res, nil
	}

	possibleActions := getPossibleActions(currentState, playerId)

	if len(possibleActions) == 0 {
		res := getScore(currentState, myPlayerId)
		stateScoreCache[hashedState] = res
		//debugAny(fmt.Sprintf("cache miss for %s (no possible actions)", hashedState), res)
		return res, nil
	}

	if maximizingPlayer {
		bestMoveValue = -1000000
		bestMove = nil

		for _, possibleAction := range possibleActions {
			nextState := applyAction(currentState, &possibleAction, playerId)
			value, _ := minimax(nextState, depth-1, myPlayerId, false)

			if value > bestMoveValue {
				bestMoveValue = value
				bestMove = &possibleAction
			}
		}
	} else {
		bestMoveValue = 1000000
		bestMove = nil

		for _, possibleAction := range possibleActions {
			nextState := applyAction(currentState, &possibleAction, playerId)
			value, _ := minimax(nextState, depth-1, myPlayerId, true)

			if value < bestMoveValue {
				bestMoveValue = value
				bestMove = &possibleAction
			}
		}
	}

	stateScoreCache[hashedState] = bestMoveValue
	//debugAny(fmt.Sprintf("cache miss for %s (recursion)", hashedState), bestMoveValue)
	return bestMoveValue, bestMove
}

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a int, b int) int {
	if a < b {
		return a
	}

	return b
}
