package main

import (
	"fmt"
	"math"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"reflect"
	deb "runtime/debug"
	_ "runtime/pprof"
	"strconv"
	"strings"
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

var LOCAL = os.Getenv("LOCAL") == "true"

// constant values
const WIDTH = 9
const HEIGHT = 9

const GRID_SIZE = WIDTH * HEIGHT

type coord struct {
	x uint8
	y uint8
}

type state struct {
	playersPosition [2]coord
	boardRemoved    compactBoolArray
	turn            uint8
}

/**
 * A compact boolean array that uses 96 bits to at least the 81 bits required to store a WIDTH*HEIGHT boolean array.
 */
type compactBoolArray struct {
	part1 uint64
	part2 uint32
}

func (c *compactBoolArray) set(index uint8, value bool) {
	if index < 64 {
		if value {
			c.part1 |= 1 << index
		} else {
			c.part1 &= ^(1 << index)
		}
	} else {
		if value {
			c.part2 |= 1 << (index - 64)
		} else {
			c.part2 &= ^(1 << (index - 64))
		}
	}
}

func (c *compactBoolArray) get(index uint8) bool {
	if index < 64 {
		return (c.part1 & (1 << index)) != 0
	} else {
		return (c.part2 & (1 << (index - 64))) != 0
	}
}

func (c *compactBoolArray) show() string {
	var result strings.Builder
	for i := uint8(0); i < GRID_SIZE; i++ {
		if c.get(i) {
			result.WriteString("X")
		} else {
			result.WriteString(".")
		}
	}
	return result.String()
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

	deb.SetGCPercent(-1)

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

	initAdjacentTilesCache()

	state := state{
		playersPosition: [2]coord{{2, 6}, {8, 4}},
		boardRemoved:    compactBoolArray{},
		turn:            0,
	}

	startedAt := time.Now()

	deadline := startedAt.Add(10000 * time.Millisecond)

	bestMove, bestScore := findBestMove(&state, 0, deadline)

	debugAny("best move", bestMove)
	debugAny("best score", bestScore)

}

func mainCG() {

	initAdjacentTilesCache()

	var playerPositionX uint8
	fmt.Scan(&playerPositionX)

	// playerPositionY: player's coordinates.
	var playerPositionY uint8
	fmt.Scan(&playerPositionY)

	playerPosition := coord{playerPositionX, playerPositionY}

	myPlayerId := uint8(0)

	if playerPositionY == 0 {
		myPlayerId = 1
	}

	opponentPosition := coord{8, 4}

	if myPlayerId == 1 {
		opponentPosition = coord{0, 4}
	}

	currentState := state{
		playersPosition: [2]coord{playerPosition, opponentPosition},
		boardRemoved:    compactBoolArray{},
		turn:            0,
	}

	for {
		var opponentPositionX uint8
		fmt.Scan(&opponentPositionX)

		startedAt := time.Now()

		deadline := startedAt.Add(1000 * time.Millisecond)

		if currentState.turn > 0 {
			deadline = startedAt.Add(100 * time.Millisecond)
		}

		debugAny(fmt.Sprintf("deadline: %v (%v)", deadline, deadline.Sub(startedAt)), nil)

		// opponentPositionY: opponent's coordinates.
		var opponentPositionY uint8
		fmt.Scan(&opponentPositionY)

		var opponentLastRemovedTileX int
		fmt.Scan(&opponentLastRemovedTileX)

		// opponentLastRemovedTileY: coordinates of the last removed tile. (-1 -1) if no tile has been removed.
		var opponentLastRemovedTileY int
		fmt.Scan(&opponentLastRemovedTileY)

		if opponentLastRemovedTileX != -1 && opponentLastRemovedTileY != -1 {
			index := opponentLastRemovedTileY*WIDTH + opponentLastRemovedTileX
			currentState.boardRemoved.set(uint8(index), true)
		}

		currentState.playersPosition[1-myPlayerId] = coord{opponentPositionX, opponentPositionY}

		debugAny("current state", currentState)

		bestAction, bestScore := findBestMove(&currentState, myPlayerId, deadline)

		debugAny("best action", bestAction)
		debugAny("best score", bestScore)

		currentState = *applyAction(&currentState, bestAction, myPlayerId)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(fmt.Sprintf("%d %d %d %d", bestAction.movePosition.x, bestAction.movePosition.y, bestAction.removeTile.x, bestAction.removeTile.y)) // action: "x y" to action or "x y message" to action and speak
	}
}

func getCurrentDuration(startedAt time.Time) time.Duration {
	return time.Since(startedAt)
}

func isTimeOver(deadline time.Time) bool {
	return time.Now().After(deadline)
}

func findBestMove(currentState *state, myPlayerId uint8, deadline time.Time) (bestAction *action, bestScore int) {
	bestAction = nil
	bestScore = -1000000

	var MaxDepth int

	// iterative deepening
	for MaxDepth = 1; !isTimeOver(deadline) && MaxDepth < 50; MaxDepth++ {
		stateScoreCache = make(map[string]int)

		depthBestScore, depthBestAction, isTimeOverSkip := minimax(currentState, MaxDepth, myPlayerId, true, -1000000, 1000000, deadline)
		if !isTimeOverSkip {
			bestScore = depthBestScore
			bestAction = depthBestAction

			// show the best move found so far
			debugAny(fmt.Sprintf("Depth %d", MaxDepth), fmt.Sprintf("best score: %d, best action: %v", bestScore, bestAction))
		} else {
			break
		}
	}

	debugAny("Depth reached", MaxDepth-1)

	return
}

func applyMove(currentState *state, movePosition coord, playerId uint8) *state {
	nextState := *currentState
	nextState.playersPosition[playerId] = movePosition
	return &nextState
}

func applyAction(state *state, action *action, playerId uint8) *state {
	nextState := applyMove(state, action.movePosition, playerId)
	index := action.removeTile.y*WIDTH + action.removeTile.x
	nextState.boardRemoved.set(uint8(index), true)
	nextState.turn++
	return nextState
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func assertEqual(expected interface{}, actual interface{}, message string) {
	if expected != actual {
		panic(fmt.Sprintf("%v != %v: %v", expected, actual, message))
	}
}

func getPossibleActions(currentState *state, playerId uint8) []action {
	allowNotNeighbor := false

	actions := make([]action, 0)

	myPosition := currentState.playersPosition[playerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range *adjacentTiles {
		if !isTileOccupied(currentState, &adjacentTile) && !isTileRemoved(currentState, &adjacentTile) {
			nextState := applyMove(currentState, adjacentTile, playerId)

			//debugAny(fmt.Sprintf("next state for %v", adjacentTile), nextState)

			if allowNotNeighbor {
				for y := uint8(0); y < HEIGHT; y++ {
					for x := uint8(0); x < WIDTH; x++ {
						c := coord{x, y}
						if !isTileOccupied(nextState, &c) && !isTileRemoved(nextState, &c) {
							actions = append(actions, action{adjacentTile, c})
						}
					}
				}
			} else {
				opponentPosition := currentState.playersPosition[1-playerId]
				adjacentTilesToOpponent := getAdjacentTiles(opponentPosition)

				foundOneRemoveTile := false

				for _, adjacentTileToOpponent := range *adjacentTilesToOpponent {
					if !isTileOccupied(nextState, &adjacentTileToOpponent) && !isTileRemoved(nextState, &adjacentTileToOpponent) {
						actions = append(actions, action{adjacentTile, adjacentTileToOpponent})
						foundOneRemoveTile = true
					}
				}

				if !foundOneRemoveTile {
					for y := uint8(0); y < HEIGHT; y++ {
						for x := uint8(0); x < WIDTH; x++ {
							c := coord{x, y}
							if !isTileOccupied(nextState, &c) && !isTileRemoved(nextState, &c) {
								actions = append(actions, action{adjacentTile, c})
							}
						}
					}
				}
			}
		}
	}

	return actions
}

func getPossibleActionsCount(currentState *state, playerId uint8) int {
	count := 0

	myPosition := currentState.playersPosition[playerId]

	adjacentTiles := getAdjacentTiles(myPosition)

	for _, adjacentTile := range *adjacentTiles {
		if !isTileOccupied(currentState, &adjacentTile) && !isTileRemoved(currentState, &adjacentTile) {
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

var cacheAdjacentTiles = make([][]coord, WIDTH*HEIGHT)

func initAdjacentTilesCache() {
	for y := uint8(0); y < HEIGHT; y++ {
		for x := uint8(0); x < WIDTH; x++ {
			position := coord{x, y}

			adjacentTiles := make([]coord, 0, 8)

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
		}
	}
}

func getAdjacentTiles(position coord) (adjacentTiles *[]coord) {
	return &(cacheAdjacentTiles[position.y*WIDTH+position.x])
}

func isTileOccupied(currentState *state, position *coord) bool {
	return currentState.playersPosition[0] == *position || currentState.playersPosition[1] == *position
}

func isTileRemoved(currentState *state, position *coord) bool {
	return currentState.boardRemoved.get(position.y*WIDTH + position.x)
}

var distanceFromPlayer [2][WIDTH * HEIGHT]int

func getScore(currentState *state, myPlayerId uint8, currentPlayerId uint8) int {
	myPossibleActions := getPossibleActionsCount(currentState, myPlayerId)
	opponentPossibleActions := getPossibleActionsCount(currentState, 1-myPlayerId)

	// a good action is a action that maximize my player closest coords and minimize opponent closest coords
	myPlayerCellsCount, opponentCellsCount := countPartitionCells(currentState, myPlayerId)

	// old for check
	//myPlayerCellsCountOld, opponentCellsCountOld := countPartitionCellsOld(currentState, myPlayerId)

	//debug(fmt.Sprintf("state %v with removed: %v", currentState, currentState.boardRemoved.show()))

	//assertEqual(myPlayerCellsCount, myPlayerCellsCountOld, "myPlayerCellsCount != myPlayerCellsCountOld")
	//assertEqual(opponentCellsCount, opponentCellsCountOld, "opponentCellsCount != opponentCellsCountOld")

	bonusEnd := 0

	myTurn := myPlayerId == currentPlayerId

	if opponentPossibleActions == 0 && !myTurn {
		bonusEnd += 1000000
		bonusEnd -= int(currentState.turn) * 1000
	} else if opponentPossibleActions == 0 {
		bonusEnd += 1000000 / 2
	}

	if myPossibleActions == 0 && myTurn {
		bonusEnd -= 1000000
		bonusEnd += int(currentState.turn) * 1000
	} else if myPossibleActions == 0 {
		bonusEnd -= 1000000 / 2
	}

	return bonusEnd + myPlayerCellsCount - opponentCellsCount + 256*myPossibleActions - 256*opponentPossibleActions
}

func getScorePossibleAction(currentState *state, myPlayerId uint8) int {
	myPossibleActions := getPossibleActionsCount(currentState, myPlayerId)
	opponentPossibleActions := getPossibleActionsCount(currentState, 1-myPlayerId)

	bonusEnd := 0
	if opponentPossibleActions == 0 {
		bonusEnd += 1000000
		bonusEnd -= int(currentState.turn) * 1000
	}

	if myPossibleActions == 0 {
		bonusEnd -= 1000000
		bonusEnd += int(currentState.turn) * 1000
	}

	return bonusEnd + 10*myPossibleActions - 10*opponentPossibleActions
}

func countPartitionCellsOld(currentState *state, myPlayerId uint8) (int, int) {
	// we use a BFS to find all the tiles that are reachable from a player

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
			for _, adj := range *adjacentTiles {
				tileIndex := adj.y*WIDTH + adj.x

				if !isTileOccupied(currentState, &adj) && !isTileRemoved(currentState, &adj) && distanceFromPlayer[playerId][tileIndex] == -1 {
					distanceFromPlayer[playerId][tileIndex] = distanceFromPlayer[playerId][currentPosition.y*WIDTH+currentPosition.x] + 1
					queue = append(queue, adj)
				}
			}
		}
	}

	myPlayerCellsCount := 0
	opponentCellsCount := 0

	// for each tile, find the closest player
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			tileIndex := y*WIDTH + x

			if distanceFromPlayer[0][tileIndex] == -1 && distanceFromPlayer[1][tileIndex] == -1 {
				// if the tile is not reachable by any player, it is not part of the partition
				continue
			}

			playerToOwn := -1

			if distanceFromPlayer[0][tileIndex] == -1 && distanceFromPlayer[1][tileIndex] != -1 {
				playerToOwn = 1
			} else if distanceFromPlayer[0][tileIndex] != -1 && distanceFromPlayer[1][tileIndex] == -1 {
				playerToOwn = 0
			} else if distanceFromPlayer[0][tileIndex] < distanceFromPlayer[1][tileIndex] {
				playerToOwn = 0
			} else if distanceFromPlayer[0][tileIndex] > distanceFromPlayer[1][tileIndex] {
				playerToOwn = 1
			}

			if playerToOwn == -1 {
				// if the tile is reachable by both players at the same distance, it is not part of the partition
				continue
			} else if playerToOwn == int(myPlayerId) {
				myPlayerCellsCount++
			} else if playerToOwn == int(1-myPlayerId) {
				opponentCellsCount++
			}
		}
	}

	return myPlayerCellsCount, opponentCellsCount
}

func contains(slice []coord, element coord) bool {
	// fastest way to check if a slice contains an element
	for i := 0; i < len(slice); i++ {
		if slice[i] == element {
			return true
		}
	}
	return false
}

func countPartitionCells(currentState *state, myPlayerId uint8) (int, int) {
	// we use a BFS to find all the tiles that are reachable from a player

	// -1 for first player
	// 1 for second player
	// 0 for no player
	colorGrid := [HEIGHT][WIDTH]int8{}

	discovered := [2][]coord{}

	discovered[0] = append(discovered[0], currentState.playersPosition[0])
	discovered[1] = append(discovered[1], currentState.playersPosition[1])

	colorGrid[currentState.playersPosition[0].y][currentState.playersPosition[0].x] = -1
	colorGrid[currentState.playersPosition[1].y][currentState.playersPosition[1].x] = 1

	myPlayerCellsCount := 1
	opponentCellsCount := 1

	newDiscovered := [2][]coord{}

	for len(discovered[0]) > 0 || len(discovered[1]) > 0 {

		//debugAny("start of loop discovered", discovered)

		// reset the new discovered tiles
		newDiscovered[0] = newDiscovered[0][:0]
		newDiscovered[1] = newDiscovered[1][:0]

		for playerId := 0; playerId < 2; playerId++ {
			for _, position := range discovered[playerId] {

				adjacentTiles := getAdjacentTiles(position)
				for _, adj := range *adjacentTiles {
					if !isTileOccupied(currentState, &adj) && !isTileRemoved(currentState, &adj) && colorGrid[adj.y][adj.x] == 0 && !contains(newDiscovered[playerId], adj) {
						newDiscovered[playerId] = append(newDiscovered[playerId], adj)
					}
				}
			}
		}

		// for all the discovered tiles that are only discovered by one player, we can assign them to this player
		//testIntersection()
		sharedDiscovered := intersection(newDiscovered[0], newDiscovered[1])

		for _, position := range sharedDiscovered {
			colorGrid[position.y][position.x] = 42
		}

		for _, position := range newDiscovered[0] {
			if colorGrid[position.y][position.x] != 0 {
				continue
			}
			colorGrid[position.y][position.x] = -1
			myPlayerCellsCount++
		}

		for _, position := range newDiscovered[1] {
			if colorGrid[position.y][position.x] != 0 {
				continue
			}
			colorGrid[position.y][position.x] = 1
			opponentCellsCount++
		}

		discovered[0] = discovered[0][:0]
		discovered[1] = discovered[1][:0]

		discovered[0] = append(discovered[0], newDiscovered[0]...)
		discovered[1] = append(discovered[1], newDiscovered[1]...)

		//debugAny("discovered", discovered)

	}

	//debugAny("colorGrid", showColorGrid(colorGrid))

	return myPlayerCellsCount, opponentCellsCount
}

func showColorGrid(colorGrid [HEIGHT][WIDTH]int8) string {
	var result string

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if colorGrid[y][x] == -1 {
				result += "0"
			} else if colorGrid[y][x] == 1 {
				result += "1"
			} else if colorGrid[y][x] == 42 {
				result += "B"
			} else {
				result += "."
			}
		}
		result += "\n"
	}

	return result
}

func testIntersection() {
	a := []coord{
		coord{0, 0},
		coord{1, 0},
		coord{2, 0},
		coord{3, 0},
		coord{4, 0},
		coord{5, 0},
		coord{6, 0},
		coord{7, 0},
		coord{8, 0},
		coord{9, 0},
	}

	b := []coord{
		coord{4, 0},
		coord{1, 1},
		coord{1, 0},
	}

	assert(reflect.DeepEqual(intersection(a, b), []coord{
		coord{1, 0},
		coord{4, 0},
	}), "intersection")
}

func intersection(a []coord, b []coord) []coord {
	var result []coord

	for _, aCoord := range a {
		for _, bCoord := range b {
			if aCoord.x == bCoord.x && aCoord.y == bCoord.y {
				result = append(result, aCoord)
			}
		}
	}

	return result
}

func getColorForPlayer(playerId int) int8 {
	color := int8(0)
	if playerId == 0 {
		color = -1
	} else {
		color = 1
	}
	return color
}

var stateScoreCache = make(map[string]int)

var strBuilder = strings.Builder{}

func hashState(currentState *state) string {
	strBuilder.Reset()
	strBuilder.Grow(WIDTH*HEIGHT + 4)

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			if currentState.boardRemoved.get(uint8(y*WIDTH + x)) {
				strBuilder.WriteString("X")
			} else {
				strBuilder.WriteString(".")
			}
		}
	}

	for playerId := 0; playerId < 2; playerId++ {
		strBuilder.WriteString(strconv.Itoa(int(currentState.playersPosition[playerId].x)))
		strBuilder.WriteString(strconv.Itoa(int(currentState.playersPosition[playerId].y)))
	}
	return strBuilder.String()
}

type actionWithStateAndScore struct {
	action *action
	state  *state
	score  int
}

func minimax(currentState *state, depth int, myPlayerId uint8, maximizingPlayer bool, alpha int, beta int, startedAt time.Time) (bestMoveValue int, bestMove *action, isTimeOverSkip bool) {
	if isTimeOver(startedAt) {
		return 0, nil, true
	}

	hashedState := hashState(currentState)

	playerId := uint8(0)
	if !maximizingPlayer {
		playerId = 1
	}

	if score, ok := stateScoreCache[hashedState]; ok {
		return score, nil, false
	}

	// todo: merge with no possible action
	if depth == 0 {
		res := getScore(currentState, myPlayerId, playerId)
		stateScoreCache[hashedState] = res
		return res, nil, false
	}

	possibleActions := getPossibleActions(currentState, playerId)

	if len(possibleActions) == 0 {
		res := getScore(currentState, myPlayerId, playerId)
		stateScoreCache[hashedState] = res
		return res, nil, false
	}

	// for each possible action, we apply it and score the resulting state
	actionWithStatesAndScores := make([]actionWithStateAndScore, len(possibleActions))
	for i := 0; i < len(possibleActions); i++ {
		possibleAction := &(possibleActions[i])
		nextState := applyAction(currentState, possibleAction, playerId)
		scoreNextState := 0
		actionWithStatesAndScores[i] = actionWithStateAndScore{possibleAction, nextState, scoreNextState}
	}

	// ordering moves by random
	rand.Shuffle(len(actionWithStatesAndScores), func(i, j int) {
		actionWithStatesAndScores[i], actionWithStatesAndScores[j] = actionWithStatesAndScores[j], actionWithStatesAndScores[i]
	})

	if maximizingPlayer {
		bestMoveValue = -1000000
		bestMove = nil

		for i := 0; i < len(actionWithStatesAndScores); i++ {
			possibleAction := &(actionWithStatesAndScores[i])
			nextState := possibleAction.state
			value, _, isTimeOverSkip := minimax(nextState, depth-1, myPlayerId, false, alpha, beta, startedAt)

			if isTimeOverSkip {
				return 0, nil, true
			}

			if value > bestMoveValue {
				bestMoveValue = value
				possibleActionCopy := possibleAction.action
				bestMove = possibleActionCopy
			}

			// Update alpha value
			alpha = max(alpha, bestMoveValue)

			// Perform alpha-beta pruning
			if beta <= alpha {
				//debugAny(fmt.Sprintf("pruning at depth %d", depth), fmt.Sprintf("alpha: %d, beta: %d", alpha, beta))
				break
			}
		}
	} else {
		bestMoveValue = 1000000
		bestMove = nil

		for i := 0; i < len(actionWithStatesAndScores); i++ {
			possibleAction := &(actionWithStatesAndScores[i])
			nextState := possibleAction.state
			value, _, isTimeOverSkip := minimax(nextState, depth-1, myPlayerId, true, alpha, beta, startedAt)

			if isTimeOverSkip {
				return 0, nil, true
			}

			if value < bestMoveValue {
				bestMoveValue = value
				possibleActionCopy := possibleAction.action
				bestMove = possibleActionCopy
			}

			// Update beta value
			beta = min(beta, bestMoveValue)

			// Perform alpha-beta pruning
			if beta <= alpha {
				//debugAny(fmt.Sprintf("pruning at depth %d", depth), fmt.Sprintf("alpha: %d, beta: %d", alpha, beta))
				break
			}
		}
	}

	stateScoreCache[hashedState] = bestMoveValue
	return bestMoveValue, bestMove, false
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
