# Isola

This is an AI to play [Isola](https://www.codingame.com/multiplayer/bot-programming/isola), currently ranked 8th/59 in the [Isola contest](https://www.codingame.com/multiplayer/bot-programming/isola/leaderboard).

It uses:
- [Minimax](https://en.wikipedia.org/wiki/Minimax)
- [Alpha-beta pruning](https://en.wikipedia.org/wiki/Alpha%E2%80%93beta_pruning)
- [Iterative deepening](https://en.wikipedia.org/wiki/Iterative_deepening_depth-first_search)
- [Move ordering](https://www.chessprogramming.org/Move_Ordering)

TODO:
- [ ] Add a transposition table
- [ ] Add a quiescence search
- [ ] Reuse the previous search in iterative deepening
- [ ] Improve the evaluation function
- [ ] Improve the move ordering (currently it's random)
- [ ] Improve performance (cache the moves, etc.)
- [ ] Implement negamax
- [ ] Implement MCTS (Monte Carlo Tree Search) and compare the results
