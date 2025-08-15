package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	North = 0
	South = 1
	East  = 2
	West  = 3
)

var dirNames = []string{"NORTH", "SOUTH", "EAST", "WEST"}

var deltas = [4][2]int{{-1, 0}, {1, 0}, {0, 1}, {0, -1}}

var prioNormal = []int{South, East, North, West}
var prioInv = []int{West, North, East, South}

var charToDir = map[byte]int{
	'N': North,
	'S': South,
	'E': East,
	'W': West,
}

type State struct {
	r, c, dir int
	inv, breaker bool
	gridkey string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := scanner.Text()
	var H, W int
	fmt.Sscanf(line, "%d %d", &H, &W)
	grid := make([][]byte, H)
	for i := 0; i < H; i++ {
		scanner.Scan()
		grid[i] = []byte(scanner.Text())
	}

	// Find start and replace '@' with ' '
	var startr, startc int
	for r := 0; r < H; r++ {
		for c := 0; c < W; c++ {
			if grid[r][c] == '@' {
				startr = r
				startc = c
				grid[r][c] = ' '
			}
		}
	}

	curR := startr
	curC := startc
	curDir := South
	curInv := false
	curBreaker := false
	path := []string{}
	visited := make(map[State]bool)

	for {
		if grid[curR][curC] == '$' {
			for _, m := range path {
				fmt.Println(m)
			}
			return
		}

		// Create grid key
		var sb strings.Builder
		for _, row := range grid {
			sb.Write(row)
		}
		gkey := sb.String()

		state := State{curR, curC, curDir, curInv, curBreaker, gkey}
		if visited[state] {
			fmt.Println("LOOP")
			return
		}
		visited[state] = true

		// Try to move
		dR := deltas[curDir][0]
		dC := deltas[curDir][1]
		nR := curR + dR
		nC := curC + dC
		blocked := true
		if nR >= 0 && nR < H && nC >= 0 && nC < W {
			ch := grid[nR][nC]
			if ch != '#' {
				if ch != 'X' || curBreaker {
					blocked = false
				}
			}
		}

		if blocked {
			// Change direction
			prio := prioNormal
			if curInv {
				prio = prioInv
			}
			found := false
			for _, nDir := range prio {
				ndR := deltas[nDir][0]
				ndC := deltas[nDir][1]
				nnR := curR + ndR
				nnC := curC + ndC
				nBlocked := true
				if nnR >= 0 && nnR < H && nnC >= 0 && nnC < W {
					nCh := grid[nnR][nnC]
					if nCh != '#' {
						if nCh != 'X' || curBreaker {
							nBlocked = false
						}
					}
				}
				if !nBlocked {
					curDir = nDir
					found = true
					break
				}
			}
			if !found {
				fmt.Println("LOOP")
				return
			}
		} else {
			// Move
			if grid[nR][nC] == 'X' {
				grid[nR][nC] = ' '
			}
			curR = nR
			curC = nC
			path = append(path, dirNames[curDir])

			// Apply effect
			char := grid[curR][curC]
			if char >= '1' && char <= '9' {
				// Teleport
				d := char
				for rR := 0; rR < H; rR++ {
					for cC := 0; cC < W; cC++ {
						if grid[rR][cC] == d && (rR != curR || cC != curC) {
							curR = rR
							curC = cC
							goto teleDone
						}
					}
				}
			teleDone:
			} else if dIr, ok := charToDir[char]; ok {
				curDir = dIr
			} else if char == 'I' {
				curInv = !curInv
			} else if char == 'B' {
				curBreaker = !curBreaker
			}
		}
	}
}