package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Direction int

const (
	SOUTH Direction = iota
	EAST
	NORTH
	WEST
)

func (d Direction) String() string {
	switch d {
	case SOUTH:
		return "SOUTH"
	case EAST:
		return "EAST"
	case NORTH:
		return "NORTH"
	case WEST:
		return "WEST"
	default:
		return "UNKNOWN"
	}
}

type Position struct {
	row, col int
}

type State struct {
	pos       Position
	direction Direction
	inverted  bool
	breaker   bool
}

type Robot struct {
	grid        [][]rune
	height      int
	width       int
	startPos    Position
	destPos     Position
	teleporters map[rune][]Position
	state       State
	visited     map[string]bool
	path        []Direction
}

func NewRobot(grid [][]rune, height, width int) *Robot {
	robot := &Robot{
		grid:        grid,
		height:      height,
		width:       width,
		teleporters: make(map[rune][]Position),
		visited:     make(map[string]bool),
		path:        []Direction{},
	}

	// Find start, destination, and teleporters
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			cell := grid[i][j]
			switch cell {
			case '@':
				robot.startPos = Position{i, j}
				robot.state.pos = Position{i, j}
				robot.state.direction = SOUTH
			case '$':
				robot.destPos = Position{i, j}
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				robot.teleporters[cell] = append(robot.teleporters[cell], Position{i, j})
			}
		}
	}

	return robot
}

func (r *Robot) getStateKey() string {
	return fmt.Sprintf("%d,%d,%d,%t,%t", r.state.pos.row, r.state.pos.col, r.state.direction, r.state.inverted, r.state.breaker)
}

func (r *Robot) getNextPosition(dir Direction) Position {
	pos := r.state.pos
	switch dir {
	case SOUTH:
		return Position{pos.row + 1, pos.col}
	case EAST:
		return Position{pos.row, pos.col + 1}
	case NORTH:
		return Position{pos.row - 1, pos.col}
	case WEST:
		return Position{pos.row, pos.col - 1}
	}
	return pos
}

func (r *Robot) isValidPosition(pos Position) bool {
	return pos.row >= 0 && pos.row < r.height && pos.col >= 0 && pos.col < r.width
}

func (r *Robot) canMoveTo(pos Position) bool {
	if !r.isValidPosition(pos) {
		return false
	}

	cell := r.grid[pos.row][pos.col]

	// Can always move to empty spaces and special cells
	if cell == ' ' || cell == '@' || cell == '$' || cell == 'S' || cell == 'E' ||
		cell == 'N' || cell == 'W' || cell == 'B' || cell == 'I' ||
		(cell >= '1' && cell <= '9') {
		return true
	}

	// Cannot move through unbreakable walls
	if cell == '#' {
		return false
	}

	// Can move through breakable walls if in breaker mode
	if cell == 'X' && r.state.breaker {
		return true
	}

	// Cannot move through breakable walls if not in breaker mode
	if cell == 'X' {
		return false
	}

	return true
}

func (r *Robot) getPriorityDirections() []Direction {
	if r.state.inverted {
		return []Direction{WEST, NORTH, EAST, SOUTH}
	}
	return []Direction{SOUTH, EAST, NORTH, WEST}
}

func (r *Robot) findPath() bool {
	for {
		// Check if we've reached the destination
		if r.state.pos == r.destPos {
			return true
		}

		// Check for loop detection
		stateKey := r.getStateKey()
		if r.visited[stateKey] {
			return false // Loop detected
		}
		r.visited[stateKey] = true

		// Process current cell
		cell := r.grid[r.state.pos.row][r.state.pos.col]

		// Handle special cells
		switch cell {
		case 'S':
			r.state.direction = SOUTH
		case 'E':
			r.state.direction = EAST
		case 'N':
			r.state.direction = NORTH
		case 'W':
			r.state.direction = WEST
		case 'I':
			r.state.inverted = !r.state.inverted
		case 'B':
			r.state.breaker = !r.state.breaker
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// Handle teleporter
			teleporters := r.teleporters[cell]
			for _, tp := range teleporters {
				if tp != r.state.pos {
					r.state.pos = tp
					break
				}
			}
		}

		// Try to move in current direction first
		nextPos := r.getNextPosition(r.state.direction)
		if r.canMoveTo(nextPos) {
			// Handle breaking walls
			if r.grid[nextPos.row][nextPos.col] == 'X' && r.state.breaker {
				r.grid[nextPos.row][nextPos.col] = ' ' // Destroy the wall
			}
			r.state.pos = nextPos
			r.path = append(r.path, r.state.direction)
			continue
		}

		// Cannot move in current direction, try priorities
		priorities := r.getPriorityDirections()
		moved := false

		for _, dir := range priorities {
			nextPos := r.getNextPosition(dir)
			if r.canMoveTo(nextPos) {
				// Handle breaking walls
				if r.grid[nextPos.row][nextPos.col] == 'X' && r.state.breaker {
					r.grid[nextPos.row][nextPos.col] = ' ' // Destroy the wall
				}
				r.state.pos = nextPos
				r.state.direction = dir
				r.path = append(r.path, dir)
				moved = true
				break
			}
		}

		// If we cannot move in any direction, we're stuck
		if !moved {
			return false
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Read dimensions
	scanner.Scan()
	dimensions := strings.Fields(scanner.Text())
	height, _ := strconv.Atoi(dimensions[0])
	width, _ := strconv.Atoi(dimensions[1])

	// Read grid
	grid := make([][]rune, height)
	for i := 0; i < height; i++ {
		scanner.Scan()
		line := scanner.Text()

		// Pad line with spaces if it's shorter than expected width
		for len(line) < width {
			line += " "
		}

		// Truncate line if it's longer than expected width
		if len(line) > width {
			line = line[:width]
		}

		grid[i] = []rune(line)
	}

	// Create robot and find path
	robot := NewRobot(grid, height, width)

	if robot.findPath() {
		// Output the path
		for _, move := range robot.path {
			fmt.Println(move)
		}
	} else {
		// Output LOOP if no path found
		fmt.Println("LOOP")
	}
}
