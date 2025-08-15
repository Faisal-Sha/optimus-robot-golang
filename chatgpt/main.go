package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Point struct{ r, c int }

const (
	SOUTH = iota
	EAST
	NORTH
	WEST
)

var dirToDelta = map[int]Point{
	SOUTH: {1, 0},
	EAST:  {0, 1},
	NORTH: {-1, 0},
	WEST:  {0, -1},
}

var dirToName = map[int]string{
	SOUTH: "SOUTH",
	EAST:  "EAST",
	NORTH: "NORTH",
	WEST:  "WEST",
}

func main() {
	runTests := flag.Bool("test", false, "Run built-in tests")
	flag.Parse()

	if *runTests {
		runHarness()
		return
	}

	runSolver()
}

// ---------------- SOLVER ----------------
func runSolver() {
	in := bufio.NewReader(os.Stdin)
	hwLine, _ := readLine(in)
	for strings.TrimSpace(hwLine) == "" {
		hwLine, _ = readLine(in)
	}
	parts := strings.Fields(hwLine)
	if len(parts) < 2 {
		fmt.Println("LOOP")
		return
	}
	H, _ := strconv.Atoi(parts[0])
	W, _ := strconv.Atoi(parts[1])

	grid := make([][]rune, H)
	start := Point{-1, -1}
	teleByDigit := map[rune][]Point{}

	for i := 0; i < H; i++ {
		line, _ := readLine(in)
		if len(line) < W {
			line = line + strings.Repeat(" ", W-len(line))
		}
		row := []rune(line[:W])
		for j, ch := range row {
			if ch == '@' {
				start = Point{i, j}
			}
			if ch >= '1' && ch <= '9' {
				teleByDigit[ch] = append(teleByDigit[ch], Point{i, j})
			}
		}
		grid[i] = row
	}

	telePair := map[Point]Point{}
	for _, arr := range teleByDigit {
		if len(arr) == 2 {
			telePair[arr[0]] = arr[1]
			telePair[arr[1]] = arr[0]
		}
	}

	pos := start
	dir := SOUTH
	breaker := false
	inverted := false
	destroyed := map[int]bool{}
	var moves []string
	visited := map[string]bool{}

	makeDestroyedSig := func() string {
		if len(destroyed) == 0 {
			return ""
		}
		keys := make([]int, 0, len(destroyed))
		for k := range destroyed {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		var b strings.Builder
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Itoa(k))
		}
		return b.String()
	}

	applyTile := func() {
		for {
			cell := grid[pos.r][pos.c]
			switch cell {
			case 'S':
				dir = SOUTH
			case 'E':
				dir = EAST
			case 'N':
				dir = NORTH
			case 'W':
				dir = WEST
			case 'B':
				breaker = !breaker
			case 'I':
				inverted = !inverted
			default:
				if cell >= '1' && cell <= '9' {
					pair := telePair[pos]
					pos = pair
					continue
				}
			}
			break
		}
	}

	canStep := func(next Point) bool {
		ch := grid[next.r][next.c]
		if ch == '#' {
			return false
		}
		if ch == 'X' && !breaker {
			return false
		}
		return true
	}

	nextDirOrder := func() []int {
		if inverted {
			return []int{WEST, NORTH, EAST, SOUTH}
		}
		return []int{SOUTH, EAST, NORTH, WEST}
	}

	applyTile()

	for {
		if grid[pos.r][pos.c] == '$' {
			break
		}

		key := fmt.Sprintf("%d,%d|%d|%t|%t|%s", pos.r, pos.c, dir, breaker, inverted, makeDestroyedSig())
		if visited[key] {
			fmt.Println("LOOP")
			return
		}
		visited[key] = true

		next := Point{pos.r + dirToDelta[dir].r, pos.c + dirToDelta[dir].c}
		if !canInside(next, H, W) || !canStep(next) {
			chosen := -1
			for _, d := range nextDirOrder() {
				nn := Point{pos.r + dirToDelta[d].r, pos.c + dirToDelta[d].c}
				if canInside(nn, H, W) && canStep(nn) {
					chosen = d
					next = nn
					break
				}
			}
			if chosen == -1 {
				fmt.Println("LOOP")
				return
			}
			dir = chosen
		}

		if grid[next.r][next.c] == 'X' && breaker {
			grid[next.r][next.c] = ' '
			destroyed[next.r*W+next.c] = true
		}

		pos = next
		moves = append(moves, dirToName[dir])
		applyTile()
	}

	for _, m := range moves {
		fmt.Println(m)
	}
}

func readLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil && len(s) == 0 {
		return s, err
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}

func canInside(p Point, H, W int) bool {
	return p.r >= 0 && p.r < H && p.c >= 0 && p.c < W
}

// ---------------- TEST HARNESS ----------------

type TestCase struct {
	Input    string
	Expected string
}
func runHarness() {
	tests := []TestCase{
		{
			Input: `5 6
######
#@E $#
# N  #
#X   #
######`,
			Expected: `SOUTH
EAST
NORTH
EAST
EAST`,
		},
		{
			Input: `10 10
##########
#        #
#  S   W #
#        #
#  $     #
#        #
#@       #
#        #
#E     N #
##########`,
			Expected: `SOUTH
SOUTH
EAST
EAST
EAST
EAST
EAST
EAST
NORTH
NORTH
NORTH
NORTH
NORTH
NORTH
WEST
WEST
WEST
WEST
SOUTH
SOUTH`,
		},
		{
			Input: `10 10
##########
# @      #
# B      #
#XXX     #
# B      #
#    BXX$#
#XXXXXXXX#
#        #
#        #
##########`,
			Expected: `SOUTH
SOUTH
SOUTH
SOUTH
EAST
EAST
EAST
EAST
EAST
EAST`,
		},
		{
			Input: `10 10
##########
#    I   #
#        #
#       $# 
#       @#
#        #
#       I#
#        #
#        #
##########`,
			Expected: `SOUTH
SOUTH
SOUTH
SOUTH
WEST
WEST
WEST
WEST
WEST
WEST
WEST
NORTH
NORTH
NORTH
NORTH
NORTH
NORTH
NORTH
EAST
EAST
EAST
EAST
EAST
EAST
EAST
SOUTH
SOUTH`,
		},
		{
			Input: `10 10
##########
#    1   #
#        #
#        #
#        #
#@       #
#        #
#        #
#    1  $# 
##########`,
			Expected: `SOUTH
SOUTH
SOUTH
EAST
EAST
EAST
EAST
EAST
EAST
EAST
SOUTH
SOUTH
SOUTH
SOUTH
SOUTH
SOUTH
SOUTH`,
		},
		{
			Input: `5 5
#####
#   #
# $ #
# @ #
#####`,
			Expected: `LOOP`,
		},
	}

	passCount := 0
	for i, tc := range tests {
		cmd := exec.Command(os.Args[0]) // run same binary
		cmd.Stdin = strings.NewReader(tc.Input)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			fmt.Printf("Test %d ERROR: %v\n", i+1, err)
			continue
		}
		got := strings.TrimSpace(out.String())
		expected := strings.TrimSpace(tc.Expected)
		if got == expected {
			fmt.Printf("Test %d: PASS\n", i+1)
			passCount++
		} else {
			fmt.Printf("Test %d: FAIL\nExpected:\n%s\nGot:\n%s\n", i+1, expected, got)
		}
	}
	fmt.Printf("\nPassed %d/%d tests\n", passCount, len(tests))
}
